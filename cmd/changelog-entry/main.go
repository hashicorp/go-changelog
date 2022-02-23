package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/hashicorp/go-changelog"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var pr, entriesDir, entryTmpl, entryType, entryBody string
	flag.StringVar(&pr, "pr", "", "the pull request number for which to generate a changelog entry")
	flag.StringVar(&entriesDir, "entries-dir", ".changelog", "the directory within the repo containing changelog entry files")
	flag.StringVar(&entryTmpl, "entry-template", filepath.Join(entriesDir, "entry.tmpl"), "the path of the file holding the template to use for the changelog entry to be generated")
	flag.StringVar(&entryType, "entry-type", "", "the type of changelog entry")
	flag.StringVar(&entryBody, "body", "", "the body of the changelog entry")
	flag.Parse()

	if pr == "" {
		fmt.Fprintln(os.Stderr, "Must specify pull request number associated with the changelog entry to be generated.")
		fmt.Fprintln(os.Stderr, "")
		flag.Usage()
		os.Exit(1)
	}

	_, err = strconv.Atoi(pr)
	if err != nil {
		log.Fatalf("Error parsing PR %q as a number: %s", pr, err)
	}

	if entryType == "" {
		fmt.Fprintln(os.Stderr, "Must specify type for the changelog entry.")
		fmt.Fprintln(os.Stderr, "")
		flag.Usage()
		os.Exit(1)
	}

	// TODO: this list should be configurable from an ordered list in the
	// implementing project which would also be used to enumerate and order the
	// sections of the changelog template file
	switch entryType {
	case "none",
		"bug",
		"note",
		"enhancement",
		"deprecation",
		"breaking-change",
		"feature":
	default:
		log.Fatalf("Unknown release-note type: %v", entryType)
	}

	tmpl, err := template.ParseFiles(filepath.Join(pwd, entryTmpl))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %q as a Go template: %s\n", entryTmpl, err)
		os.Exit(1)
	}

	filename := pr + ".txt"
	f, err := os.Create(filepath.Join(filepath.Base(entriesDir), filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, changelog.Note{
		Type: entryType,
		Body: entryBody,
	})
}
