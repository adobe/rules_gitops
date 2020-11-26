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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	oe "os/exec"
	"path/filepath"

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
func Clone(repo, dir, mirrorDir, primaryBranch string) (*Repo, error) {
	if err := os.RemoveAll(dir); err != nil {
		return nil, fmt.Errorf("Unable to clone repo: %w", err)
	}
	if mirrorDir != "" {
		exec.Mustex("", "git", "clone", "-n", "--reference", mirrorDir, repo, dir)
	} else {
		exec.Mustex("", "git", "clone", "-n", repo, dir)
	}
	exec.Mustex(dir, "git", "config", "--local", "core.sparsecheckout", "true")
	if err := ioutil.WriteFile(filepath.Join(dir, ".git/info/sparse-checkout"), []byte("cloud/\n"), 0644); err != nil {
		return nil, fmt.Errorf("Unable to create .git/info/sparse-checkout: %w", err)
	}
	exec.Mustex(dir, "git", "checkout", primaryBranch)

	return &Repo{
		Dir: dir,
	}, nil
}

// Repo is a clone of a git repository. Create with Clone, and don't
// forget to clean it up after.
type Repo struct {
	// Dir is the location of the git repo.
	Dir string
}

// Clean cleans up the repo
func (r *Repo) Clean() error {
	return os.RemoveAll(r.Dir)
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
func (r *Repo) Commit(message string) bool {
	exec.Mustex(r.Dir, "git", "add", "cloud")
	if r.IsClean() {
		return false
	}
	exec.Mustex(r.Dir, "git", "commit", "-a", "-m", message)
	return true
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
	args := append([]string{"push", "origin", "-f", "--set-upstream"}, branches...)
	exec.Mustex(r.Dir, "git", args...)
}
