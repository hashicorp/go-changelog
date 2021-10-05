package changelog

import (
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Entry struct {
	Issue string
	Body  string
	Date  time.Time
	Hash  string
}

type changelog struct {
	content []byte
	hash    string
	date    time.Time
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
	entriesAfter := make(map[string]changelog, len(entriesAfterFI))
	for _, i := range entriesAfterFI {
		fp := filepath.Join(dir, i.Name())
		f, err := wt.Filesystem.Open(fp)
		if err != nil {
			return nil, err
		}
		contents, err := ioutil.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		log, err := r.Log(&git.LogOptions{FileName: &fp})
		if err != nil {
			return nil, err
		}
		lastChange, err := log.Next()
		if err != nil {
			return nil, err
		}
		entriesAfter[i.Name()] = changelog{content: contents, date: lastChange.Author.When, hash: lastChange.Hash.String()}
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
	for name, cl := range entriesAfter {
		entries = append(entries, Entry{
			Issue: name,
			Body:  string(cl.content),
			Date:  cl.date,
			Hash:  cl.hash,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Issue < entries[j].Issue
	})

	return entries, nil
}
