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
package client

import (
	"testing"
)

var (
	setup             K8STestSetup
	readyCallbacksRun = map[string]bool{}
)

func TestMain(m *testing.M) {
	firstCallback := func() error {
		readyCallbacksRun["first"] = true
		return nil
	}

	secondCallback := func() error {
		readyCallbacksRun["second"] = true
		return nil
	}

	setup := K8STestSetup{
		PortForwardServices: map[string]int{},
		ReadyCallbacks: []Callback{
			firstCallback,
			secondCallback,
		},
	}

	setup.TestMain(m)
}

// TestReadyCallback validates that the pre-test ReadyCallback is run. Note that this test scenario assumes
// that a K8STestSetup in TestMain will invoke the test.
func TestReadyCallback(t *testing.T) {
	if len(readyCallbacksRun) != 2 {
		t.Fatalf("all ready callbacks should have been run")
	}
}
