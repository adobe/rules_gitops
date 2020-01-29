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
package exec

import (
	"log"
	"os/exec"
	"strings"
)

// Ex is a shortcut for executing the command in specified dir
func Ex(dir, name string, arg ...string) (output string, err error) {
	log.Println("executing:", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	b, err := cmd.CombinedOutput()
	log.Printf("%s", string(b))
	return string(b), err
}

// Mustex executes the command name arg... in directory dir
// it will exit with fatal error if execution was not successful
func Mustex(dir, name string, arg ...string) {
	_, err := Ex(dir, name, arg...)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

}
