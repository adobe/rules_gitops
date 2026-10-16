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
	"strconv"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		cfg          Config
		failures     int
		wantCode     int
		wantAttempts int
		wantSleeps   int
	}{
		{"success first try", Config{MaxAttempts: 5}, 0, 0, 1, 0},
		{"recovers after transient failures", Config{MaxAttempts: 5}, 3, 0, 4, 3},
		{"recovers on last attempt", Config{MaxAttempts: 3}, 2, 0, 3, 2},
		{"gives up after max attempts", Config{MaxAttempts: 3}, 100, 7, 3, 2},
		{"retries disabled", Config{MaxAttempts: 1}, 100, 7, 1, 0},
		{"zero attempts runs once", Config{MaxAttempts: 0}, 100, 7, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0
			sleeps := 0
			attempt := func() error {
				attempts++
				if attempts <= tt.failures {
					return helperCmd(t, 7).Run()
				}
				return nil
			}
			got := run(tt.cfg, attempt, func(time.Duration) { sleeps++ })
			if got != tt.wantCode {
				t.Errorf("exit code = %d, want %d", got, tt.wantCode)
			}
			if attempts != tt.wantAttempts {
				t.Errorf("attempts = %d, want %d", attempts, tt.wantAttempts)
			}
			if sleeps != tt.wantSleeps {
				t.Errorf("sleeps = %d, want %d", sleeps, tt.wantSleeps)
			}
		})
	}
}

func TestRunBackoff(t *testing.T) {
	cfg := Config{MaxAttempts: 6, Delay: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond}
	var waits []time.Duration
	run(cfg, func() error { return helperCmd(t, 1).Run() }, func(d time.Duration) { waits = append(waits, d) })
	base := []time.Duration{10, 20, 40, 50, 50}
	if len(waits) != len(base) {
		t.Fatalf("got %d waits, want %d", len(waits), len(base))
	}
	for i, b := range base {
		b *= time.Millisecond
		if waits[i] < b || waits[i] > b+b/2 {
			t.Errorf("wait[%d] = %s, want within [%s, %s]", i, waits[i], b, b+b/2)
		}
	}
}

func TestExitCodeNonExitError(t *testing.T) {
	err := exec.Command("/nonexistent/binary").Run()
	if got := exitCode(err); got != 1 {
		t.Errorf("exitCode = %d, want 1", got)
	}
}

func TestDefaultConfigFromEnv(t *testing.T) {
	t.Setenv("GITOPS_PUSH_MAX_ATTEMPTS", "2")
	t.Setenv("GITOPS_PUSH_RETRY_DELAY_SECONDS", "0")
	t.Setenv("GITOPS_PUSH_MAX_RETRY_DELAY_SECONDS", "9")
	cfg := defaultConfig()
	if cfg.MaxAttempts != 2 || cfg.Delay != 0 || cfg.MaxDelay != 9*time.Second {
		t.Errorf("unexpected config: %+v", cfg)
	}
	t.Setenv("GITOPS_PUSH_MAX_ATTEMPTS", "garbage")
	if got := defaultConfig().MaxAttempts; got != 5 {
		t.Errorf("invalid env should fall back to default, got %d", got)
	}
}

// TestHelperProcess is re-executed by helperCmd to produce a real non-zero exit code.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("PUSHRETRY_HELPER") != "1" {
		return
	}
	code, _ := strconv.Atoi(os.Getenv("PUSHRETRY_EXIT"))
	os.Exit(code)
}

func helperCmd(t *testing.T, code int) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "PUSHRETRY_HELPER=1", "PUSHRETRY_EXIT="+strconv.Itoa(code))
	return cmd
}

// TestRunWithRealCommand retries a real shell command that fails until its third invocation.
func TestRunWithRealCommand(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "count")
	attempt := func() error {
		cmd := exec.Command("sh", "-c", `n=$(cat "$1" 2>/dev/null || echo 0); n=$((n+1)); echo $n > "$1"; [ $n -ge 3 ]`, "sh", marker)
		return cmd.Run()
	}
	if got := run(Config{MaxAttempts: 5}, attempt, func(time.Duration) {}); got != 0 {
		t.Fatalf("exit code = %d, want 0", got)
	}
	b, _ := os.ReadFile(marker)
	if string(b) != "3\n" {
		t.Errorf("attempts recorded = %q, want 3", b)
	}
}
