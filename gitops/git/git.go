/*
Copyright 2020 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/
package git

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	oe "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adobe/rules_gitops/gitops/exec"
)

// PushRejectedError indicates a non-fast-forward rejection (e.g. a concurrent run already
// pushed to the same branch), distinguishable from other git failures via errors.As.
type PushRejectedError struct {
	Output string
}

func (e *PushRejectedError) Error() string {
	return fmt.Sprintf("push rejected, remote has new commits: %s", e.Output)
}

func nonFastForwardRejection(output string) (rejected bool, reason string) {
	// Parses `git push --porcelain` summary lines: "<flag>\t<from>:<to>\t<summary>".
	for _, line := range strings.Split(output, "\n") {
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) != 3 || fields[0] != "!" {
			continue
		}
		summary := fields[2]
		if strings.Contains(summary, "non-fast-forward") ||
			strings.Contains(summary, "fetch first") ||
			strings.Contains(summary, "stale info") {
			return true, summary
		}
	}
	return false, ""
}

var (
	git = "git"
)

// Clone clones a repository. Pass the full repository name, such as
// "https://aleksey.pesternikov@bitbucket.tubemogul.info/scm/tm/repo.git" as the repo.
// Cloned directory will be clean of local changes with primaryBranch branch checked out.
// repo: https://aleksey.pesternikov@bitbucket.tubemogul.info/scm/tm/repo.git
// dir: /tmp/cloudrepo
// mirrorDir: optional (if not empty) local mirror of the repository
func Clone(repo, dir, mirrorDir, primaryBranch, gitopsPath string) (*Repo, error) {
	if err := os.RemoveAll(dir); err != nil {
		return nil, fmt.Errorf("Unable to clone repo: %w", err)
	}
	remoteName := "origin"
	args := []string{"clone", "--no-checkout", "--single-branch", "--branch", primaryBranch, "--filter=blob:none", "--no-tags", "--origin", remoteName}
	if mirrorDir != "" {
		args = append(args, "--reference", mirrorDir)
	}
	args = append(args, repo, dir)
	exec.Mustex("", "git", args...)
	// Enable sparse-checkout when restricting to a subdir
	if !isRootPath(gitopsPath) {
		exec.Mustex(dir, "git", "config", "--local", "core.sparsecheckout", "true")
		genPath := fmt.Sprintf("%s/\n", gitopsPath)
		if err := ioutil.WriteFile(filepath.Join(dir, ".git/info/sparse-checkout"), []byte(genPath), 0644); err != nil {
			return nil, fmt.Errorf("Unable to create .git/info/sparse-checkout: %w", err)
		}
	}
	exec.Mustex(dir, "git", "checkout", primaryBranch)
	return &Repo{
		Dir:        dir,
		RemoteName: remoteName,
	}, nil
}

// Repo is a clone of a git repository. Create with Clone, and don't
// forget to clean it up after.
type Repo struct {
	// Dir is the location of the git repo.
	Dir string
	// RemoteName is the name of the remote that tracks upstream repository.
	RemoteName string
}

// Clean cleans up the repo
func (r *Repo) Clean() error {
	return os.RemoveAll(r.Dir)
}

// Fetch branches from the remote repository based on a specified pattern.
// The branches will be be added to the list tracked remote branches ready to be pushed.
func (r *Repo) Fetch(pattern string) {
	exec.Mustex(r.Dir, "git", "remote", "set-branches", "--add", r.RemoteName, pattern)
	exec.Mustex(r.Dir, "git", "fetch", "--force", "--filter=blob:none", "--no-tags", r.RemoteName)
}

// UpdatePrimaryBranch fast-forwards the local primaryBranch ref to match the remote.
func (r *Repo) UpdatePrimaryBranch(primaryBranch string) {
	exec.Mustex(r.Dir, "git", "fetch", r.RemoteName, primaryBranch)
	exec.Mustex(r.Dir, "git", "checkout", primaryBranch)
	exec.Mustex(r.Dir, "git", "reset", "--hard", r.RemoteName+"/"+primaryBranch)
}

// ResyncBranchWithRemote discards any local, unpushed commits on branch and resets it to
// the remote's current tip. No-op if the branch doesn't exist on the remote yet.
func (r *Repo) ResyncBranchWithRemote(branch string) {
	remoteRef := r.RemoteName + "/" + branch
	if _, err := exec.Ex(r.Dir, "git", "rev-parse", "--verify", remoteRef); err != nil {
		return
	}
	exec.Mustex(r.Dir, "git", "checkout", "-B", branch, remoteRef)
}

// SwitchToBranch switch the repo to specified branch and checkout primaryBranch files over it.
// if branch does not exist it will be created
func (r *Repo) SwitchToBranch(branch, primaryBranch string) (new bool) {
	if _, err := exec.Ex(r.Dir, "git", "checkout", branch); err != nil {
		// error checking out, create new
		exec.Mustex(r.Dir, "git", "branch", branch, primaryBranch)
		exec.Mustex(r.Dir, "git", "checkout", branch)
		return true
	}
	return false
}

// RecreateBranch discards a branch content and reset it from primaryBranch.
func (r *Repo) RecreateBranch(branch, primaryBranch string) {
	exec.Mustex(r.Dir, "git", "checkout", primaryBranch)
	exec.Mustex(r.Dir, "git", "branch", "-f", branch, primaryBranch)
	exec.Mustex(r.Dir, "git", "checkout", branch)
}

// GetLastCommitMessage fetches the commit message from the most recent change of the branch
func (r *Repo) GetLastCommitMessage() (msg string) {
	msg, err := exec.Ex(r.Dir, "git", "log", "-1", "--pretty=%B")
	if err != nil {
		return ""
	}
	return msg
}

// Commit all changes to the current branch. returns true if there were any changes
func (r *Repo) Commit(message, gitopsPath string) bool {
	if isRootPath(gitopsPath) {
		exec.Mustex(r.Dir, "git", "add", ".")
	} else {
		exec.Mustex(r.Dir, "git", "add", gitopsPath)
	}
	if r.IsClean() {
		return false
	}
	exec.Mustex(r.Dir, "git", "commit", "-a", "-m", message)
	return true
}

// RestoreFile restores the specified file in the repository to its original state
func (r *Repo) RestoreFile(fileName string) {
	exec.Mustex(r.Dir, "git", "checkout", "--", fileName)
}

// GetChangedFiles returns a list of files that have been changed in the repository
func (r *Repo) GetChangedFiles() []string {
	s, err := exec.Ex(r.Dir, "git", "diff", "--name-only")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	var files []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		files = append(files, sc.Text())
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	return files
}

// IsClean returns true if there is no local changes (nothing to commit)
func (r *Repo) IsClean() bool {
	cmd := oe.Command("git", "status", "--porcelain")
	cmd.Dir = r.Dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	return len(b) == 0
}

// Push pushes all local changes to the remote repository.
// All changes should already be committed. Returns a *PushRejectedError for a
// non-fast-forward rejection; any other failure is returned as a plain error.
func (r *Repo) Push(branches []string) error {
	args := append([]string{"push", r.RemoteName, "-f", "--set-upstream", "--porcelain"}, branches...)
	output, err := exec.Ex(r.Dir, "git", args...)
	if err != nil {
		if rejected, reason := nonFastForwardRejection(output); rejected {
			return &PushRejectedError{Output: reason}
		}
		return fmt.Errorf("git push failed: %w (output: %s)", err, output)
	}
	return r.verifyPushed(branches)
}

// verifyAttempts, verifyRetryDelay and lsRemoteTimeout bound how long verifyPushed spends on
// ls-remote: failures are retried, and each attempt is cut off after lsRemoteTimeout.
var (
	verifyAttempts   = 3
	verifyRetryDelay = time.Second
	lsRemoteTimeout  = 30 * time.Second
)

// pushURL returns the URL `git push` uses for the remote, which can differ from the fetch URL
// (remote.<name>.pushurl). With several push URLs only the first is returned.
// It runs git directly rather than through exec.Ex, which logs arguments and output: the URL may
// embed credentials.
func (r *Repo) pushURL() string {
	cmd := oe.Command("git", "remote", "get-url", "--push", r.RemoteName)
	cmd.Dir = r.Dir
	output, err := cmd.Output()
	if err != nil {
		return r.RemoteName
	}
	if fields := strings.Fields(string(output)); len(fields) > 0 {
		return fields[0]
	}
	return r.RemoteName
}

// containsCommit reports whether remoteTip is already in the local object store and descends
// from (or equals) commit, i.e. a later push by someone else built on top of ours.
func (r *Repo) containsCommit(remoteTip, commit string) bool {
	if remoteTip == "" {
		return false
	}
	if _, err := exec.Ex(r.Dir, "git", "cat-file", "-e", remoteTip+"^{commit}"); err != nil {
		return false
	}
	_, err := exec.Ex(r.Dir, "git", "merge-base", "--is-ancestor", commit, remoteTip)
	return err == nil
}

// lsRemote lists refs on the push remote, retrying transient failures. The command is not
// logged by exec.Ex because its arguments contain the push URL.
func (r *Repo) lsRemote(refs []string) (output string, err error) {
	args := append([]string{"ls-remote", r.pushURL()}, refs...)
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), lsRemoteTimeout)
		cmd := oe.CommandContext(ctx, "git", args...)
		cmd.Dir = r.Dir
		// Helpers such as ssh can outlive the killed git and keep its output pipe open.
		cmd.WaitDelay = time.Second
		var b []byte
		b, err = cmd.CombinedOutput()
		cancel()
		output = string(b)
		if err == nil || attempt >= verifyAttempts {
			return output, err
		}
		time.Sleep(time.Duration(attempt) * verifyRetryDelay)
	}
}

// verifyPushed checks the remote tip of each branch. Some git clients exit 0 for a push the
// remote refused, so the exit status alone can't be trusted. The check is best effort: if the
// remote can't be queried the push is assumed to have landed, since git already reported success.
func (r *Repo) verifyPushed(branches []string) error {
	refs := make([]string, len(branches))
	for i, b := range branches {
		refs[i] = "refs/heads/" + b
	}
	output, err := r.lsRemote(refs)
	if err != nil {
		log.Printf("WARNING: unable to verify pushed branches, assuming the push landed: %v (output: %s)", err, output)
		return nil
	}
	remote := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		if fields := strings.Fields(line); len(fields) == 2 {
			remote[fields[1]] = fields[0]
		}
	}
	for i, b := range branches {
		local, err := exec.Ex(r.Dir, "git", "rev-parse", refs[i])
		if err != nil {
			return fmt.Errorf("git rev-parse %s failed: %w (output: %s)", refs[i], err, local)
		}
		local = strings.TrimSpace(local)
		if got := remote[refs[i]]; got != local && !r.containsCommit(got, local) {
			return &PushRejectedError{Output: fmt.Sprintf("remote %s is at %q after push, expected %s", b, got, local)}
		}
	}
	return nil
}

// isRootPath is an internal helper to detect "full repo" case.
func isRootPath(gitopsPath string) bool {
	return gitopsPath == "" || gitopsPath == "."
}
