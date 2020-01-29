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
package commitmsg

import (
	"log"
	"strings"
)

const begin = "--- gitops targets begin ---"
const end = "--- gitops targets end ---"

// ExtractTargets extracts list of gitops targets used in a commit
func ExtractTargets(msg string) (packages []string) {
	betweenMarkers := false
	for _, s := range strings.Split(msg, "\n") {
		switch s {
		case begin:
			betweenMarkers = true
		case end:
			betweenMarkers = false
		default:
			if betweenMarkers {
				packages = append(packages, s)
			}
		}
	}
	if betweenMarkers {
		log.Print("Unable to find end marker in commit message")
	}
	return
}

// Generate generates a commit message from a list of targets
func Generate(targets []string) string {
	var sb strings.Builder
	sb.WriteByte('\n')
	sb.WriteString(begin)
	sb.WriteByte('\n')

	for _, t := range targets {
		sb.WriteString(t)
		sb.WriteByte('\n')
	}

	sb.WriteString(end)
	sb.WriteByte('\n')
	return sb.String()
}
