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

func TestNonFastForwardRejection(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantReject bool
		wantReason string
	}{
		{
			name: "fetch first",
			output: "To /path/to/remote\n" +
				"!\trefs/heads/deploy/app:refs/heads/deploy/app\t[rejected] (fetch first)\n" +
				"Done\nerror: failed to push some refs to '/path/to/remote'\nhint: Updates were rejected",
			wantReject: true,
			wantReason: "[rejected] (fetch first)",
		},
		{
			name: "non-fast-forward",
			output: "To /path/to/remote\n" +
				"!\trefs/heads/deploy/app:refs/heads/deploy/app\t[rejected] (non-fast-forward)\n" +
				"Done\n",
			wantReject: true,
			wantReason: "[rejected] (non-fast-forward)",
		},
		{
			name: "server-side deny (branch protection)",
			output: "To /path/to/remote\n" +
				"!\trefs/heads/deploy/app:refs/heads/deploy/app\t[remote rejected] (non-fast-forward)\n" +
				"Done\nremote: error: denying non-fast-forward refs/heads/deploy/app\n",
			wantReject: true,
			wantReason: "[remote rejected] (non-fast-forward)",
		},
		{
			name: "stale info",
			output: "To /path/to/remote\n" +
				"!\trefs/heads/deploy/app:refs/heads/deploy/app\t[rejected] (stale info)\n" +
				"Done\n",
			wantReject: true,
			wantReason: "[rejected] (stale info)",
		},
		{
			name: "permission denied",
			output: "remote: Permission to foo/bar.git denied\n" +
				"fatal: unable to access 'https://example.com/': The requested URL returned error: 403",
			wantReject: false,
		},
		{
			name:       "unrelated failure",
			output:     "fatal: unable to access 'https://example.com/': Could not resolve host",
			wantReject: false,
		},
		{
			name: "server-side hook decline is not a race",
			output: "To /path/to/remote\n" +
				"!\trefs/heads/deploy/app:refs/heads/deploy/app\t[remote rejected] (pre-receive hook declined)\n" +
				"Done\n",
			wantReject: false,
		},
		{
			name: "keywords outside the porcelain field are ignored",
			output: "remote: this push would not be a fast-forward, see fetch first for details\n" +
				"fatal: unable to access 'https://example.com/': Could not resolve host",
			wantReject: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReject, gotReason := nonFastForwardRejection(tt.output)
			if gotReject != tt.wantReject {
				t.Errorf("nonFastForwardRejection() rejected = %v, want %v", gotReject, tt.wantReject)
			}
			if gotReject && gotReason != tt.wantReason {
				t.Errorf("nonFastForwardRejection() reason = %q, want %q", gotReason, tt.wantReason)
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
