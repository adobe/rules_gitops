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
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/adobe/rules_gitops/templating/fasttemplate"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	stampInfoFile      arrayFlags
	output             string
	format, formatFile string
)

func init() {
	flag.Var(&stampInfoFile, "stamp-info-file", "Paths to info_file and version_file files for stamping.")
	flag.StringVar(&output, "output", "", "The output file")
	flag.StringVar(&formatFile, "format-file", "", "The file containing stamp variables placeholders")
	flag.StringVar(&format, "format", "", "The format string containing stamp variables")
}

func workspaceStatusDict(filenames []string) map[string]interface{} {
	d := map[string]interface{}{}
	for _, f := range filenames {
		content, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalf("Unable to read %s: %v", f, err)
		}
		for _, l := range strings.Split(string(content), "\n") {
			sv := strings.SplitN(l, " ", 2)
			if len(sv) == 2 {
				d[sv[0]] = sv[1]
			}
		}
	}
	return d
}

func main() {
	var err error
	flag.Parse()
	stamps := workspaceStatusDict(stampInfoFile)
	if formatFile != "" {
		if format != "" {
			log.Fatal("only one of --format or --format-file should be used")
		}
		imp, err := ioutil.ReadFile(formatFile)
		if err != nil {
			log.Fatalf("Unable to read file %s: %v", formatFile, err)
		}
		format = string(imp)
	}

	outf := os.Stdout
	if output != "" {
		outf, err = os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Unable to create output file %s: %v", output, err)
		}
		defer outf.Close()
	}
	_, err = fasttemplate.Execute(format, "{", "}", outf, stamps)
	if err != nil {
		log.Fatalf("Unable to execute template %s: %v", format, err)
	}
}
