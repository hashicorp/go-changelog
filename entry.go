package changelog

import (
	"fmt"
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

// Diff returns the slice of Entry values that represent the difference of
// entries in the dir directory within repo from ref1 revision to ref2 revision.
// ref1 and ref2 should be valid git refs as strings and dir should be a valid
// directory path in the repository.
//
// The function calculates the diff by first checking out ref2 and collecting
// the set of all entries in dir. It then checks out ref1 and subtracts the
// entries found in dir. The resulting set of entries is then filtered to
// exclude any entries that came before the commit date of ref1.
//
// Along the way, if any git or filesystem interactions fail, an error is returned.
func Diff(repo, ref1, ref2, dir string) ([]Entry, error) {
	r, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: repo,
	})
	if err != nil {
		return nil, err
	}
	rev2, err := r.ResolveRevision(plumbing.Revision(ref2))
	if err != nil {
		return nil, fmt.Errorf("could not resolve revision %s: %w", ref2, err)
	}
	var rev1 *plumbing.Hash
	if ref1 != "-" {
		rev1, err = r.ResolveRevision(plumbing.Revision(ref1))
		if err != nil {
			return nil, fmt.Errorf("could not resolve revision %s: %w", ref1, err)
		}
	}
	wt, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	if err := wt.Checkout(&git.CheckoutOptions{
		Hash:  *rev2,
		Force: true,
	}); err != nil {
		return nil, fmt.Errorf("could not checkout repository at %s: %w", ref2, err)
	}
	entriesAfterFI, err := wt.Filesystem.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read repository directory %s: %w", dir, err)
	}
	// a set of all entries at rev2 (this release); the set of entries at ref1
	// will then be subtracted from it to arrive at a set of 'candidate' entries.
	entryCandidates := make(map[string]bool, len(entriesAfterFI))
	for _, i := range entriesAfterFI {
		entryCandidates[i.Name()] = true
	}
	if rev1 != nil {
		err = wt.Checkout(&git.CheckoutOptions{
			Hash:  *rev1,
			Force: true,
		})
		entriesBeforeFI, err := wt.Filesystem.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("could not read repository directory %s: %w", dir, err)
		}
		for _, i := range entriesBeforeFI {
			delete(entryCandidates, i.Name())
		}
		// checkout rev2 so that we can read files later
		if err := wt.Checkout(&git.CheckoutOptions{
			Hash:  *rev2,
			Force: true,
		}); err != nil {
			return nil, fmt.Errorf("could not checkout repository at %s: %w", ref2, err)
		}
	}

	entries := make([]Entry, 0, len(entryCandidates))
	for name := range entryCandidates {
		fp := filepath.Join(dir, name)
		f, err := wt.Filesystem.Open(fp)
		if err != nil {
			return nil, fmt.Errorf("error opening file at %s: %w", name, err)
		}
		contents, err := ioutil.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading file at %s: %w", name, err)
		}
		log, err := r.Log(&git.LogOptions{FileName: &fp})
		if err != nil {
			return nil, err
		}
		lastChange, err := log.Next()
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{
			Issue: name,
			Body:  string(contents),
			Date:  lastChange.Author.When,
			Hash:  lastChange.Hash.String(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Issue < entries[j].Issue
	})

	return entries, nil
}
