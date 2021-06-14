package changelog

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

var TypeValues = []string{"enhancement",
	"feature",
	"bug",
	"note",
	"new-resource",
	"new-datasource",
	"deprecation",
	"breaking-change",
}

type Entry struct {
	Issue string
	Body  string
}

func Diff(repo, ref1, ref2, dir string) ([]Entry, error) {
	r, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: repo,
	})
	if err != nil {
		return nil, err
	}
	rev2, err := r.ResolveRevision(plumbing.Revision(ref2))
	if err != nil {
		return nil, err
	}
	var rev1 *plumbing.Hash
	if ref1 != "-" {
		rev1, err = r.ResolveRevision(plumbing.Revision(ref1))
		if err != nil {
			return nil, err
		}
	}
	wt, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Hash:  *rev2,
		Force: true,
	})
	if err != nil {
		return nil, err
	}
	entriesAfterFI, err := wt.Filesystem.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entriesAfter := make(map[string][]byte, len(entriesAfterFI))
	for _, i := range entriesAfterFI {
		f, err := wt.Filesystem.Open(filepath.Join(dir, i.Name()))
		if err != nil {
			return nil, err
		}
		contents, err := ioutil.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		entriesAfter[i.Name()] = contents
	}
	if rev1 != nil {
		err = wt.Checkout(&git.CheckoutOptions{
			Hash:  *rev1,
			Force: true,
		})
		if err != nil {
			return nil, err
		}
		entriesBeforeFI, err := wt.Filesystem.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, i := range entriesBeforeFI {
			delete(entriesAfter, i.Name())
		}
	}
	entries := make([]Entry, 0, len(entriesAfter))
	for name, contents := range entriesAfter {
		entries = append(entries, Entry{
			Issue: name,
			Body:  string(contents),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Issue < entries[j].Issue
	})
	return entries, nil
}

func TypeValid(Type string) bool {
	for _, a := range TypeValues {
		if a == Type {
			return true
		}
	}
	return false
}
