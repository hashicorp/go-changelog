package changelog

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
)

func BenchmarkDiff(b *testing.B) {
	repo := cloneRepoLocal(b, "https://github.com/hashicorp/nomad.git")
	//repo := "/tmp/git/nomad"
	ref1, ref2, dir := "v1.11.0", "v1.11.1", ".changelog"

	b.Run("memory", func(b *testing.B) {
		b.Log("begin")
		for b.Loop() {
			_, err := Diff(repo, ref1, ref2, dir)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("filesystem", func(b *testing.B) {
		b.Log("begin")
		for b.Loop() {
			_, err := DiffLocal(repo, ref1, ref2, dir)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func cloneRepoLocal(b *testing.B, url string) (dir string) {
	b.Helper()
	tmp := b.TempDir()
	b.Logf("cloning repo %q to %q", url, tmp)
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stderr,
	})
	if err != nil {
		b.Fatalf("failed to clone repo: %v", err)
	}
	return tmp
}
