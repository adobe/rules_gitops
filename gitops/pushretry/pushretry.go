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

// pushretry runs a command (an image pusher) and retries it on any non-zero exit.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Config controls retry behavior.
type Config struct {
	MaxAttempts int
	Delay       time.Duration
	MaxDelay    time.Duration
}

func envInt(name string, def int) int {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return def
}

func defaultConfig() Config {
	return Config{
		MaxAttempts: envInt("GITOPS_PUSH_MAX_ATTEMPTS", 5),
		Delay:       time.Duration(envInt("GITOPS_PUSH_RETRY_DELAY_SECONDS", 5)) * time.Second,
		MaxDelay:    time.Duration(envInt("GITOPS_PUSH_MAX_RETRY_DELAY_SECONDS", 60)) * time.Second,
	}
}

// exitCode extracts the process exit code from err, or 1 if it is not an exit error.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() > 0 {
		return ee.ExitCode()
	}
	return 1
}

// run executes attempt until it succeeds or cfg.MaxAttempts is reached and returns the
// last exit code. sleep is injected for tests.
func run(cfg Config, attempt func() error, sleep func(time.Duration)) int {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	delay := cfg.Delay
	for n := 1; ; n++ {
		err := attempt()
		if err == nil {
			return 0
		}
		code := exitCode(err)
		if n >= cfg.MaxAttempts {
			log.Printf("push failed after %d attempt(s): %v", n, err)
			return code
		}
		wait := delay
		if delay > 0 {
			wait += time.Duration(rand.Int63n(int64(delay/2) + 1))
		}
		log.Printf("push attempt %d/%d failed: %v; retrying in %s", n, cfg.MaxAttempts, err, wait)
		sleep(wait)
		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s -- command [args...]\n", os.Args[0])
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	attempt := func() error {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		return cmd.Run()
	}
	os.Exit(run(defaultConfig(), attempt, time.Sleep))
}
