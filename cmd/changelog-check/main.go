package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-changelog"
)

func main() {
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	entry := changelog.Entry{
		Issue: os.Args[1],
		Body:  string(in),
	}
	notes := changelog.NotesFromEntry(entry)
	if len(notes) < 1 {
		fmt.Println("no changelog entry")
		os.Exit(1)
	}
	for _, note := range notes {
		switch note.Type {
		case "bug",
			"improvement",
			"feature",
			"breaking-change",
			"new-resource",
			"new-datasource",
			"new-data-source",
			"deprecation":
		default:
			fmt.Fprintf(os.Stderr, "unknown changelog type %q\n", note.Type)
			os.Exit(1)
		}
	}
}
