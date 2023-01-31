package main

import (
	"flag"
	"log"
	"os"

	"github.com/adobe/rules_gitops/gitops/k8ssort"
)

var filter = flag.String("filter", "", "Show only manifests that match this `Kind`")
var reverse = flag.Bool("reverse", false, "Reverse the sorting order")

func main() {
	flag.Parse()
	b, err := k8ssort.Sort(os.Stdin, *filter, *reverse)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(b)
}
