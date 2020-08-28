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
package resolver_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	resolver "github.com/adobe/rules_gitops/resolver/pkg"
)

func TestNoError(t *testing.T) {
	testcases := []struct {
		name   string
		imgmap map[string]string
	}{
		{"happypath", map[string]string{
			"salist":      "docker.io/rtb/sacli/cmd/salist/image@sha256:5711bcf54511ab2fef6e08d9c9f9ae3f3a269e66834048465cc7502adb0d489b",
			"filewatcher": "docker.io/kube/filewatcher/image:tag",
		}},
		{"cwf", map[string]string{
			"helloworld-image": "docker.io/kube/hello/image:tag",
		}},
		{"flinkapp", map[string]string{
			"flinkapp-image": "docker.io/kube/flink/image:tag",
		}},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			infn := fmt.Sprintf("testdata/%s.yaml", testcase.name)
			expectedfn := fmt.Sprintf("testdata/%s.expected.yaml", testcase.name)
			inf, err := os.Open(infn)
			if err != nil {
				t.Errorf("Unable to open file %s", infn)
				return
			}
			defer inf.Close()
			expectedb, err := ioutil.ReadFile(expectedfn)
			if err != nil {
				t.Errorf("Unable to read file %s", expectedfn)
				return
			}
			expected := strings.TrimSpace(string(expectedb))
			var outbuf bytes.Buffer
			err = resolver.ResolveImages(inf, &outbuf, testcase.imgmap)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
				return
			}
			if strings.TrimSpace(outbuf.String()) != expected {
				t.Errorf("Unexpected output: %s", outbuf.String())
			}
		})
	}
}
