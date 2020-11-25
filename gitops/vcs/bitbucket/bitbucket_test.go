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
package bitbucket

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePRRemote(t *testing.T) {
	t.Skip("Manual")
	user := "********"
	pass := "*************"
	bitbucketUser = &user
	bitbucketPassword = &pass
	err := CreatePR("deploy/test1", "feature/AP-0000", "test")
	if err != nil {
		t.Error("Unexpected error from server: ", err)
	}
}

func TestCreatePRNew(t *testing.T) {
	var buf []byte
	var srverr error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, srverr = ioutil.ReadAll(r.Body)
		http.Error(w, "Created", 201)
		fmt.Fprintln(w, "PR created")
	}))
	defer ts.Close()
	oldendpoint := *apiEndpoint
	defer func() { *apiEndpoint = oldendpoint }()
	*apiEndpoint = ts.URL
	err := CreatePR("deploy/test1", "feature/AP-0000", "test")
	if err != nil {
		t.Error("Unexpected error from server: ", err)
	}
	if srverr != nil {
		t.Error("Unexpected error: ", srverr)
	}
	expectedreq := `{"title":"test","description":"test","state":"OPEN","open":true,"closed":false,"fromRef":{"id":"refs/heads/deploy/test1","repository":{"slug":"repo","project":{"key":"TM"}}},"toRef":{"id":"refs/heads/feature/AP-0000","repository":{"slug":"repo","project":{"key":"TM"}}},"locked":false}`
	if string(buf) != expectedreq {
		t.Error("Unexpected request body: ", string(buf))
	}
}
