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
package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsNonFastForwardRejection(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "fetch first",
			output: "! [rejected]  deploy/app -> deploy/app (fetch first)\nhint: Updates were rejected",
			want:   true,
		},
		{
			name:   "non-fast-forward",
			output: "! [rejected]  deploy/app -> deploy/app (non-fast-forward)",
			want:   true,
		},
		{
			name:   "server-side deny (branch protection)",
			output: "remote: error: denying non-fast-forward refs/heads/deploy/app\n! [remote rejected] deploy/app -> deploy/app (non-fast-forward)",
			want:   true,
		},
		{
			name:   "stale info",
			output: "! [rejected]  deploy/app -> deploy/app (stale info)",
			want:   true,
		},
		{
			name:   "permission denied",
			output: "remote: Permission to foo/bar.git denied\nfatal: unable to access",
			want:   false,
		},
		{
			name:   "unrelated failure",
			output: "fatal: unable to access 'https://example.com/': Could not resolve host",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonFastForwardRejection(tt.output); got != tt.want {
				t.Errorf("isNonFastForwardRejection() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

// TestPushRetryRecoversFromRejection simulates two create_gitops_prs runs racing to update
// the same deployment branch: repoA pushes first and succeeds; repoB, cloned before that
// push landed, gets rejected. It then verifies the retry path (Fetch + ResyncBranchWithRemote
// + regenerate + Commit) recovers and pushes cleanly.
func TestPushRetryRecoversFromRejection(t *testing.T) {
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
	repoA, err := Clone(remoteDir, cloneA, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone (A) failed: %v", err)
	}
	cloneB := t.TempDir()
	repoB, err := Clone(remoteDir, cloneB, "", "main", ".")
	if err != nil {
		t.Fatalf("Clone (B) failed: %v", err)
	}

	for _, r := range []*Repo{repoA, repoB} {
		runGit(t, r.Dir, "config", "user.email", "t@test")
		runGit(t, r.Dir, "config", "user.name", "t")
	}

	const deployBranch = "deploy/app-stage"

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
	err = repoB.Push([]string{deployBranch})
	var rejected *PushRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("expected a *PushRejectedError, got: %v", err)
	}

	repoB.Fetch("deploy/*")
	repoB.ResyncBranchWithRemote(deployBranch)
	writeFile(t, cloneB, "out.yaml", "from-b-retry")
	if !repoB.Commit("from B retry", ".") {
		t.Fatal("expected repoB.Commit to report changes after resync")
	}
	if err := repoB.Push([]string{deployBranch}); err != nil {
		t.Fatalf("retried repoB.Push should succeed: %v", err)
	}
}
