/*
Copyright 2026 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/adobe/rules_gitops/gitops/git"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPushWithRetrySucceedsOnFirstPush(t *testing.T) {
	remoteDir := t.TempDir()
	runGit(t, "", "init", "--bare", "--initial-branch=main", remoteDir)

	seedDir := t.TempDir()
	runGit(t, "", "clone", remoteDir, seedDir)
	writeFile(t, seedDir, "README.md", "seed")
	runGit(t, seedDir, "add", ".")
	runGit(t, seedDir, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "-m", "seed")
	runGit(t, seedDir, "push", "origin", "main")

	cloneDir := t.TempDir()
	repo, err := git.Clone(remoteDir, cloneDir, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
	runGit(t, repo.Dir, "config", "user.email", "t@test")
	runGit(t, repo.Dir, "config", "user.name", "t")

	repo.SwitchToBranch("deploy/app", "main")
	writeFile(t, cloneDir, "out.yaml", "v1")
	if !repo.Commit("v1", ".") {
		t.Fatal("expected Commit to report changes")
	}

	regenerateCalled := false
	got := pushWithRetry(repo, func() []string {
		regenerateCalled = true
		return []string{"deploy/app"}
	}, []string{"deploy/app"})

	if regenerateCalled {
		t.Error("regenerate should not be called when the first push succeeds")
	}
	if len(got) != 1 || got[0] != "deploy/app" {
		t.Errorf("got %v, want [deploy/app]", got)
	}
}

func TestPushWithRetryRecoversFromRejection(t *testing.T) {
	remoteDir := t.TempDir()
	runGit(t, "", "init", "--bare", "--initial-branch=main", remoteDir)
	runGit(t, remoteDir, "config", "receive.denyNonFastForwards", "true")

	seedDir := t.TempDir()
	runGit(t, "", "clone", remoteDir, seedDir)
	writeFile(t, seedDir, "README.md", "seed")
	runGit(t, seedDir, "add", ".")
	runGit(t, seedDir, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "-m", "seed")
	runGit(t, seedDir, "push", "origin", "main")

	cloneA := t.TempDir()
	repoA, err := git.Clone(remoteDir, cloneA, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone (A) failed: %v", err)
	}
	cloneB := t.TempDir()
	repoB, err := git.Clone(remoteDir, cloneB, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone (B) failed: %v", err)
	}
	for _, r := range []*git.Repo{repoA, repoB} {
		runGit(t, r.Dir, "config", "user.email", "t@test")
		runGit(t, r.Dir, "config", "user.name", "t")
	}

	const deployBranch = "deploy/app"

	repoA.SwitchToBranch(deployBranch, "main")
	writeFile(t, cloneA, "out.yaml", "from-a")
	if !repoA.Commit("from A", ".") {
		t.Fatal("expected repoA.Commit to report changes")
	}
	if err := repoA.Push([]string{deployBranch}); err != nil {
		t.Fatalf("repoA.Push should succeed: %v", err)
	}

	repoB.SwitchToBranch(deployBranch, "main")
	writeFile(t, cloneB, "out.yaml", "from-b")
	if !repoB.Commit("from B", ".") {
		t.Fatal("expected repoB.Commit to report changes")
	}

	regenerateCalls := 0
	got := pushWithRetry(repoB, func() []string {
		regenerateCalls++
		repoB.Fetch("deploy/*")
		repoB.ResyncBranchWithRemote(deployBranch)
		writeFile(t, cloneB, "out.yaml", "from-b-retry")
		if !repoB.Commit("from B retry", ".") {
			t.Fatal("expected repoB.Commit to report changes after resync")
		}
		return []string{deployBranch}
	}, []string{deployBranch})

	if regenerateCalls != 1 {
		t.Errorf("regenerate called %d times, want 1", regenerateCalls)
	}
	if len(got) != 1 || got[0] != deployBranch {
		t.Errorf("got %v, want [%s]", got, deployBranch)
	}
}

func TestPushWithRetryStopsWhenRegenerateHasNothingToPush(t *testing.T) {
	remoteDir := t.TempDir()
	runGit(t, "", "init", "--bare", "--initial-branch=main", remoteDir)
	runGit(t, remoteDir, "config", "receive.denyNonFastForwards", "true")

	seedDir := t.TempDir()
	runGit(t, "", "clone", remoteDir, seedDir)
	writeFile(t, seedDir, "README.md", "seed")
	runGit(t, seedDir, "add", ".")
	runGit(t, seedDir, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "-m", "seed")
	runGit(t, seedDir, "push", "origin", "main")

	cloneA := t.TempDir()
	repoA, err := git.Clone(remoteDir, cloneA, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone (A) failed: %v", err)
	}
	cloneB := t.TempDir()
	repoB, err := git.Clone(remoteDir, cloneB, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone (B) failed: %v", err)
	}
	for _, r := range []*git.Repo{repoA, repoB} {
		runGit(t, r.Dir, "config", "user.email", "t@test")
		runGit(t, r.Dir, "config", "user.name", "t")
	}

	const deployBranch = "deploy/app"

	repoA.SwitchToBranch(deployBranch, "main")
	writeFile(t, cloneA, "out.yaml", "from-a")
	if !repoA.Commit("from A", ".") {
		t.Fatal("expected repoA.Commit to report changes")
	}
	if err := repoA.Push([]string{deployBranch}); err != nil {
		t.Fatalf("repoA.Push should succeed: %v", err)
	}

	repoB.SwitchToBranch(deployBranch, "main")
	writeFile(t, cloneB, "out.yaml", "from-b")
	if !repoB.Commit("from B", ".") {
		t.Fatal("expected repoB.Commit to report changes")
	}

	got := pushWithRetry(repoB, func() []string {
		return nil
	}, []string{deployBranch})

	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
