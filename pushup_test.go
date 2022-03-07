package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	ignore "github.com/sabhiram/go-gitignore"
)

// testGit runs a git *do method and raises a fatal error if it fails.
func testGit(t *testing.T, f func(...string) ([]byte, error), arg ...string) {
	t.Helper()
	if _, err := f(arg...); err != nil {
		t.Fatal(err)
	}
}

// mkTestRepo creates a test repo.
//
// The repo will be cleaned up automatically when the test terminates.
func mkTestRepo(t *testing.T, g git, arg ...string) git {
	d, err := ioutil.TempDir("", "pushup-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })

	testGit(t, g.do, append([]string{"init", d}, arg...)...)
	res := g.in(d)

	// hack. TODO(mkm) cleanup
	if len(arg) == 1 && arg[0] == "--bare" {
		res.gitDir = res.workTree
	}

	conf := filepath.Join(res.gitDir, "config")
	testGit(t, g.do, "config", "-f", conf, "user.name", "Test Testovic")
	testGit(t, g.do, "config", "-f", conf, "user.email", "testovic@test.com")

	return res
}

func mkTestRepoPair(t *testing.T, g git) (git, git) {
	ba := mkTestRepo(t, g, "--bare")
	r1 := mkTestRepo(t, g)
	r2 := mkTestRepo(t, g)

	t.Logf("REPOS ba=%v,r1=%v,r2=%v", ba, r1, r2)

	testGit(t, r1.do, "remote", "add", "origin", ba.gitDir)
	testGit(t, r2.do, "remote", "add", "origin", ba.gitDir)

	testCommit(t, r2, "README.md")
	testGit(t, r2.sdo, "push", "-u", "origin", "master")

	testGit(t, r1.sdo, "pull", "origin", "master")
	testGit(t, r1.do, "branch", "-u", "origin/master")
	return r1, r2
}

func randHex() []byte {
	b := make([]byte, 32)
	rand.Read(b)
	return []byte(hex.EncodeToString(b))
}

func testCommit(t *testing.T, g git, filename string) {
	abs := filepath.Join(g.workTree, filename)
	if err := ioutil.WriteFile(abs, randHex(), 0600); err != nil {
		t.Fatal(err)
	}
	testGit(t, g.do, "add", filename)
	testGit(t, g.do, "commit", "-m", "test commit")
}

func TestOk(t *testing.T) {
	r, up := mkTestRepoPair(t, git{})

	testCommit(t, r, "somefile")
	testCommit(t, up, "README.md")

	testGit(t, up.sdo, "push")

	p := pushup{
		git:         r,
		ignoreFiles: []string{"bots/", "*.md"},
		retries:     2,
	}
	if err := p.run(); err != nil {
		t.Fatal(err)
	}

	testGit(t, up.sdo, "pull")

	repoHead, err := r.revParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	upstreamHead, err := up.revParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := repoHead, upstreamHead; got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestConflict(t *testing.T) {
	r, up := mkTestRepoPair(t, git{})

	testCommit(t, r, "somefile")
	testCommit(t, up, "someotherfile")

	testGit(t, up.sdo, "push")

	p := pushup{
		git:         r,
		ignoreFiles: []string{"*.md"},
		retries:     2,
	}
	err := p.run()
	if got, want := err.Error(), `"someotherfile" doesn't match ignore patterns ["*.md"]`; got != want {
		t.Fatalf("got: %q, want: %q", got, want)
	}
}

func TestDryRun(t *testing.T) {
	r, up := mkTestRepoPair(t, git{})

	testCommit(t, r, "somefile")
	testCommit(t, up, "README.md")

	testGit(t, up.sdo, "push")

	p := pushup{
		git:         r,
		ignoreFiles: []string{"bots/", "*.md"},
		retries:     2,
		dryRun:      true,
	}
	if err := p.run(); err != nil {
		t.Fatal(err)
	}

	testGit(t, up.sdo, "pull")

	repoHead, err := r.revParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	upstreamHead, err := up.revParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	if repoHead == upstreamHead {
		t.Errorf("repo HEAD: %q, upstream HEAD: %q", repoHead, repoHead)
	}
}

func TestMatchesPath(t *testing.T) {
	testCases := []struct {
		ok bool
		s  string
		i  []string
	}{
		{true, "README.md", []string{"*.md"}},
		{false, "README.txt", []string{"*.md"}},
		{true, "dir/file", []string{"dir/**"}},
		{true, "dir/file", []string{"dir/"}},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			m := ignore.CompileIgnoreLines(tc.i...)
			if got, want := m.MatchesPath(tc.s), tc.ok; got != want {
				t.Errorf("got: %v, wanted %v", got, want)
			}
		})
	}
}
