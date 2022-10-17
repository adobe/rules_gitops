package main

import (
	"flag"
	"log"
	"os"

	filter "github.com/adobe/rules_gitops/testing/it_manifest_filter/pkg"
)

var (
	inf  = flag.String("infile", "", "Input file")
	outf = flag.String("outfile", "", "Out file")
)

func main() {
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

	err := filter.ReplacePDWithEmptyDirs(infile, outfile)
	if err != nil {
		log.Fatalf("Unable to process: %s", err)
	}

}
