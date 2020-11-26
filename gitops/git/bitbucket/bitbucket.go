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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	apiEndpoint       = flag.String("bitbucket_api_pr_endpoint", "https://bitbucket.tubemogul.info/rest/api/1.0/projects/TM/repos/repo/pull-requests", "bitbucket pull request api endpoint with project and repo")
	bitbucketUser     = flag.String("bitbucket_user", os.Getenv("BITBUCKET_USER"), "bitbucket api user")
	bitbucketPassword = flag.String("bitbucket_password", os.Getenv("BITBUCKET_PASSWORD"), "bitbucket api user password")
)

type project struct {
	Key string `json:"key,omitempty"`
}

type repository struct {
	Slug    string  `json:"slug,omitempty"`
	Project project `json:"project"`
}

type pullrequestEndpoint struct {
	ID         string     `json:"id,omitempty"`
	Repository repository `json:"repository,omitempty"`
}

type account struct {
	User user `json:"user"`
}

type user struct {
	Name string `json:"name,omitempty"`
}

type pullrequest struct {
	Title       string               `json:"title,omitempty"`
	Description string               `json:"description,omitempty"`
	State       string               `json:"state,omitempty"`
	Open        bool                 `json:"open"`
	Closed      bool                 `json:"closed"`
	FromRef     *pullrequestEndpoint `json:"fromRef,omitempty"`
	ToRef       *pullrequestEndpoint `json:"toRef,omitempty"`
	Locked      bool                 `json:"locked"`
	Reviewers   []account            `json:"reviewers,omitempty"`
}

// CreatePR creates a pull request using branch names from and to
func CreatePR(from, to, title string) error {
	repo := repository{
		Slug:    "repo",
		Project: project{"TM"},
	}
	prReq := pullrequest{
		Title:       title,
		Description: title,
		State:       "OPEN",
		Open:        true,
		Closed:      false,
		FromRef: &pullrequestEndpoint{
			ID:         "refs/heads/" + from,
			Repository: repo,
		},
		ToRef: &pullrequestEndpoint{
			ID:         "refs/heads/" + to,
			Repository: repo,
		},
		Locked:    false,
		Reviewers: []account{},
	}
	json, err := json.Marshal(&prReq)
	if err != nil {
		return fmt.Errorf("Unable to marshal CreatePR request: %w", err)
	}
	req, err := http.NewRequest("POST", *apiEndpoint, bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(*bitbucketUser, *bitbucketPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to send CreatePR request: %w", err)
	}
	log.Printf("bitbucket api response: %s", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Print("bitbucket response: ", string(body))
	// 201 created
	// 409 already exists
	if resp.StatusCode == 201 {
		log.Print("PR was created")
		return nil
	}
	if resp.StatusCode == 409 {
		log.Print("reusing existing PR")
		return nil
	}
	return fmt.Errorf("Unrecognized bitbucket response %d", resp.StatusCode)
}
