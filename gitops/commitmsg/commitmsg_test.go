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
package commitmsg_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/adobe/rules_gitops/gitops/commitmsg"
)

func TestRoundtrip(t *testing.T) {
	targets := []string{"target1", "target2"}
	msg := commitmsg.Generate(targets)
	t2 := commitmsg.ExtractTargets(msg)
	if !reflect.DeepEqual(targets, t2) {
		t.Errorf("Unexpected targets after parsing: %v", t2)
	}
}

func ExampleGenerate() {
	targets := []string{"target1", "target2"}
	msg := commitmsg.Generate(targets)
	fmt.Println(msg)
	// Output:
	// --- gitops targets begin ---
	// target1
	// target2
	// --- gitops targets end ---
}
