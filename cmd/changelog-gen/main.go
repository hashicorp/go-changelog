package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/github"
	"github.com/manifoldco/promptui"
	"os"
	"regexp"
	"strings"
	"text/template"
)

//go:embed changelog.tmpl
var changelogTplDefault string

type Note struct {
	// service touched by pr
	Service string
	//release note type (Bug...)
	Type string
	// release note text
	Description string
	//pr number
	Pr  int
	URL string
}

type GithubCred struct {
	owner string
	repo  string
	token string
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var service, Type, Description, changelogTmpl, url string
	var pr int
	var noUrl bool
	flag.BoolVar(&noUrl, "no-url", true, "do not add github issue url")
	flag.IntVar(&pr, "pr", -1, "pr number")
	flag.StringVar(&service, "service", "", "the service the pr change (not mandatory)")
	flag.StringVar(&Type, "type", "", "The pr type")
	flag.StringVar(&Description, "description", "", "the changelog-gen description entry")
	flag.StringVar(&changelogTmpl, "changelog-template", "", "the path of the file holding the template to use for the entire changelog-gen")
	flag.Parse()

	if pr == -1 {
		pr, url, err = getPrNumberFromGithub(pwd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Must specify pr number or run in a valid github repo")
			fmt.Fprintln(os.Stderr, "")
			flag.Usage()
			os.Exit(1)
		}
	}

	if Type == "" {
		prompt := promptui.Select{
			Label: "Select a change type",
			Items: []string{"enhancement",
				"bug",
				"note",
				"new-resource",
				"new-datasource",
				"deprecation",
				"breaking-change",
				"none"},
		}

		_, Type, err = prompt.Run()

		if err != nil {
			fmt.Fprintln(os.Stderr, "Must specify the change type")
			fmt.Fprintln(os.Stderr, "")
			flag.Usage()
			os.Exit(1)
		}
	}

	if Description == "" {
		prompt := promptui.Prompt{Label: "Description"}
		Description, err = prompt.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Must specify the change description")
			fmt.Fprintln(os.Stderr, "")
			flag.Usage()
			os.Exit(1)
		}
	}
	var tmpl *template.Template
	if changelogTmpl != "" {
		file, err := os.ReadFile(changelogTmpl)
		if err != nil {
			os.Exit(1)
		}
		tmpl, err = template.New("").Parse(string(file))
		if err != nil {
			os.Exit(1)
		}
	} else {
		tmpl, err = template.New("").Parse(changelogTplDefault)
		if err != nil {
			os.Exit(1)
		}
	}

	if noUrl {
		url = ""
	}
	n := Note{Type: Type, Description: Description, Service: service, Pr: pr, URL: url}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, n)
	fmt.Printf("\n%s\n", buf.String())
	if err != nil {
		os.Exit(1)
	}
	err = os.WriteFile(fmt.Sprintf("%s/%d.txt", pwd, pr), buf.Bytes(), 0644)
	if err != nil {
		os.Exit(1)
	}
}

func OpenGit(path string) (*git.Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		if path == "/" {
			return r, err
		} else {
			return OpenGit(path[:strings.LastIndex(path, "/")])
		}
	}
	return r, err
}

func getPrNumberFromGithub(path string) (int, string, error) {
	r, err := OpenGit(path)
	if err != nil {
		return -1, "", err
	}

	ref, err := r.Head()
	if err != nil {
		return -1, "", err
	}

	rem, err := r.Remote("origin")
	if err != nil {
		return -1, "", err
	}

	cli := github.NewClient(nil)

	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 200},
		Sort:        "updated",
		Direction:   "desc",
	}

	ctx := context.Background()
	if len(rem.Config().URLs) <= 0 {
		return -1, "", errors.New("not able to parse repo and org")
	}
	repoUrl := rem.Config().URLs[0]

	reg := regexp.MustCompile(".*github\\.com:(.*)/(.*)\\.git")
	m := reg.FindAllStringSubmatch(repoUrl, -1)
	if len(m) > 1 {
		if len(m[0]) < 2 {
			return -1, "", errors.New("not able to parse repo and org")
		}
	}

	list, _, err := cli.PullRequests.List(ctx, m[0][1], m[0][2], opt)
	if err != nil {
		return -1, "", err
	}
	branchName := ref.Name().Short()
	if branchName == "master" || branchName == "main" {
		return -1, "", errors.New("cannot find a pr # for the main branch")
	}
	for _, pr := range list {
		if pr.Head == nil {
			continue
		}
		if pr.Head.Ref == nil {
			continue
		}

		if *pr.Head.Ref == branchName {
			n := pr.GetNumber()
			if n != 0 {
				return n, pr.GetHTMLURL(), nil
			}
		}
	}
	return -1, "", errors.New("not found")
}
