package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-changelog/parser"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var changelogPath string
	flag.StringVar(&changelogPath, "path", filepath.Join(wd, "CHANGELOG.md"), "path to the changelog file")

	// extractVersion represents version to extract changelog for (e.g. 1.0.0)
	extractVersion := flag.Arg(0)
	flag.Parse()

	if extractVersion == "" {
		fmt.Fprintf(os.Stderr, "Must specify version\n\n")
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(changelogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %s", err)
		os.Exit(1)
	}

	sp, err := parser.NewSectionParser(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read changelog file: %s", err)
		os.Exit(1)
	}
	s, err := sp.Section(extractVersion)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	_, err = os.Stdout.Write(s.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
