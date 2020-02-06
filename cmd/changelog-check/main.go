package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-changelog"
)

func main() {
	ctx := context.Background()
	pr := os.Args[1]
	prNo, err := strconv.ParseInt(pr, 10, 64)
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

	var file *github.CommitFile
	opt := &github.ListOptions{
		PerPage: 100,
	}
	for {
		commitFiles, resp, err := client.PullRequests.ListFiles(ctx,
			owner, repo, int(prNo), opt)
		if err != nil {
			log.Fatalf("Error retrieving pull request github.com/"+
				"%s/%s/%d: %w", owner, repo, prNo, err)
		}
		for _, f := range commitFiles {
			if strings.HasPrefix(f.GetFilename(), ".changelog/") {
				if file != nil {
					log.Fatalf("Two changelog files found for PR %s: %w and %w",
						f.GetFilename(), file.GetFilename())
				}
				file = f
				break
			}
		}
		if file != nil {
			break
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	if file == nil {
		log.Fatalf("No changelog file found for PR %s", pr)
	}
	resp, err := http.Get(file.GetRawURL())
	if err != nil {
		log.Fatalf("Error retrieving changelog file contents from %q: %w",
			file.GetRawURL(), err)
	}
	body, err := readResponse(resp)
	if err != nil {
		log.Fatalf("Error reading changelog file contents from %q: %w",
			file.GetRawURL(), err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP error reading changelog file contents from %q: %d %s",
			file.GetRawURL(), resp.StatusCode, body)
	}
	issue := strings.TrimPrefix(file.GetFilename(), ".changelog/")
	issue = strings.TrimSuffix(issue, ".txt")
	entry := changelog.Entry{
		Issue: issue,
		Body:  string(body),
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

func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
