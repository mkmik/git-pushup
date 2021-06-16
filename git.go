package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// git contains the state/flags for running git.
type git struct {
	gitDir   string
	workTree string
}

// mkenv returns an env variable encoded in the standard key=value format.
func mkenv(k, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

// Do runs a git command as a subprocess and returns the output.
func (g git) do(arg ...string) ([]byte, error) {
	return g.wdo(os.Stderr, arg...)
}

// Sdo is like do but silent on stderr.
func (g git) sdo(arg ...string) ([]byte, error) {
	return g.wdo(ioutil.Discard, arg...)
}

func (g git) wdo(stderr io.Writer, arg ...string) ([]byte, error) {
	if false {
		log.Printf("Executing: git %q ; in %q", arg, g.gitDir)
	}
	c := exec.Command("git", arg...)
	c.Stderr = stderr
	c.Env = os.Environ()
	if g.gitDir != "" {
		c.Env = append(c.Env, mkenv("GIT_DIR", g.gitDir))
	}
	if wt := g.workTree; wt != "" {
		c.Dir = wt
	}
	return c.Output()
}

// in returns a new git executor rooted in a different repository.
func (g git) in(d string) git {
	g.gitDir = filepath.Join(d, ".git")
	w, err := g.getWorkTree()
	if err != nil {
		panic(err)
	}
	g.workTree = w
	return g
}

func (g git) revParse(r string) (string, error) {
	b, err := g.do("rev-parse", r)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// getWorkTree returns the working tree directory.
func (g git) getWorkTree() (string, error) {
	if _, err := os.Stat(g.gitDir); os.IsNotExist(err) {
		g.gitDir = strings.TrimSuffix(g.gitDir, "/.git")
	} else if err != nil {
		return "", err
	}
	g.workTree = ""
	b, err := g.do("worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	s := string(b)
	l := strings.SplitN(s, "\n", 2)[0]
	p := strings.SplitN(l, " ", 2)[1]
	return p, nil
}
