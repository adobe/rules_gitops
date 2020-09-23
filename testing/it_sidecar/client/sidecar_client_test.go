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

// This file currently only tests a scenario involving hooks test setup using K8STestSetup.
var (
	setup K8STestSetup

	// NOTE: these global variables are used specifically for this test scenario, and should
	// be kept in mind when more tests are added in the future.
	isTestRun bool
	isHookRun bool
)

func TestMain(m *testing.M) {
	setup := K8STestSetup{PortForwardServices: map[string]int{}}

	hook := func() {
		isHookRun = true
	}

	setup.TestMainWithHook(m, hook)
}

// TestSetupHook validates that the pre-test hook is run. Note that this test scenario assumes
// that the K8STestSetup TestMain will invoke the test, as a success will be detected when either
// the test is not invoked or if the hook is successfully run.
func TestSetupHook(t *testing.T) {
	isTestRun = true
	if !isHookRun {
		t.Fatalf("Pre hook was not run!")
	}
}
