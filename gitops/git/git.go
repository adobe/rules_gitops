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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	oe "os/exec"
	"path/filepath"
	"strings"

	"github.com/adobe/rules_gitops/gitops/exec"
)

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
	// Only enable sparse-checkout when restricting to a subdir
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

// Push pushes all local changes to the remote repository
// all changes should be already commited
func (r *Repo) Push(branches []string) {
	args := append([]string{"push", r.RemoteName, "-f", "--set-upstream"}, branches...)
	exec.Mustex(r.Dir, "git", args...)
}

// isRootPath is an internal helper to detect "full repo" case.
func isRootPath(gitopsPath string) bool {
	return gitopsPath == "" || gitopsPath == "."
}