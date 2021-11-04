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
package test_no_callback

import (
	"github.com/adobe/rules_gitops/testing/it_sidecar/client"
	"testing"
)

var (
	setup         client.K8STestSetup
)

// The setup should run without error since ReadyCallback is optional
func TestMain(m *testing.M) {
	setup := client.K8STestSetup{
		PortForwardServices: map[string]int{},
	}

	setup.TestMain(m)
}
