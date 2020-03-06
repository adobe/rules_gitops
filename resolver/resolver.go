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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	resolver "github.com/adobe/rules_gitops/resolver/pkg"
)

type imagesFlags map[string]string

func (i *imagesFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *imagesFlags) Set(value string) error {
	v := strings.SplitN(value, "=", 2)
	if len(v) != 2 {
		return errors.New("image parameter should be in form imagename=imagevalue")
	}
	(*i)[strings.TrimSpace(v[0])] = strings.TrimSpace(v[1])
	return nil
}

var (
	inf    = flag.String("infile", "", "Input file")
	outf   = flag.String("outfile", "", "Out file")
	images = make(imagesFlags)
)

func main() {
	flag.Var(&images, "image", "imagename=imagevalue")
	flag.Parse()
	infile := os.Stdin
	if *inf != "" {
		f, err := os.Open(*inf)
		if err != nil {
			log.Fatalf("Unable to open file %s for reading: %s", *inf, err)
		}
		defer f.Close()
		infile = f
	}
	outfile := os.Stdout
	if *outf != "" {
		f, err := os.Create(*outf)
		if err != nil {
			log.Fatalf("Unable to create file %s for reading: %s", *outf, err)
		}
		defer f.Close()
		outfile = f
	}

	err := resolver.ResolveImages(infile, outfile, images)
	if err != nil {
		log.Fatalf("Unable to process: %s", err)
	}

}
