package changelog

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

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
