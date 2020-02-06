package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-changelog"
)

func main() {
	ctx := context.Background()
	pr := os.Args[1]
	prNo, err := strconv.Atoi(pr)
	if err != nil {
		log.Fatalf("Error parsing PR %q as a number: %w", pr, err)
	}

	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	if owner == "" {
		log.Fatalf("GITHUB_OWNER not set")
	}
	if repo == "" {
		log.Fatalf("GITHUB_REPO not set")
	}

	client := github.NewClient(nil)

	pullRequest, _, err := client.PullRequests.Get(ctx, owner, repo, prNo)
	if err != nil {
		log.Fatalf("Error retrieving pull request github.com/"+
			"%s/%s/%d: %w", owner, repo, prNo, err)
	}
	entry := changelog.Entry{
		Issue: pr,
		Body:  pullRequest.GetBody(),
	}
	notes := changelog.NotesFromEntry(entry)
	if len(notes) < 1 {
		log.Fatalf("no changelog entry found in %s: %s", entry.Issue,
			string(entry.Body))
	}
	for _, note := range notes {
		switch note.Type {
		case "none",
			"bug",
			"note",
			"enhancement",
			"new-resource",
			"new-datasource",
			"deprecation",
			"breaking-change",
			"feature":
		default:
			log.Fatalf("unknown changelog type %q", note.Type)
		}
	}
}
