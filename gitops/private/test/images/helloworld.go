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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	port = flag.Int("port", 8080, "IP port")
)

func printenv(w http.ResponseWriter, r *http.Request) {
	for _, e := range os.Environ() {
		fmt.Fprintf(w, "%s\n", e)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<html><body>Hello World!</body></html>")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", home)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok\n")
	})
	http.HandleFunc("/env", printenv)
	fmt.Printf("Serving on port %d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
