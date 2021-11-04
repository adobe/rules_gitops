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
package test_callback

import (
	"github.com/adobe/rules_gitops/testing/it_sidecar/client"
	"testing"
)

var (
	setup         client.K8STestSetup
	isCallbackRun bool
)

func TestMain(m *testing.M) {
	callback := func() error {
		isCallbackRun = true
		return nil
	}

	setup := client.K8STestSetup{
		PortForwardServices: map[string]int{},
		ReadyCallback:       callback,
	}

	setup.TestMain(m)
}

// TestReadyCallback validates that the pre-test ReadyCallback is run. Note that this test scenario assumes
// that a K8STestSetup in TestMain will invoke the test.
func TestReadyCallback(t *testing.T) {
	if !isCallbackRun {
		t.Fatalf("ready callback should have been run")
	}
}
