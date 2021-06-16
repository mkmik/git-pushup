package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"

	ignore "github.com/sabhiram/go-gitignore"
)

// Type pushup holds all the input parameters and state for executing the pushup tool.
type pushup struct {
	git
	ignoreFiles []string
	retries     int
	rebase      bool
	verbose     bool
	dryRun      bool
}

// Run contains the main entrypoint for git-pushup.
func (p *pushup) run() error {
	if len(p.ignoreFiles) == 0 {
		return fmt.Errorf("need at least one ignore pattern")
	}
	for i := 0; i < p.retries; i++ {
		log.Printf("attempt %d of %d", i+1, p.retries)
		log.Printf("pushing...")

		// --dry-run skips the actual push, pretends the push failed,
		// pulls from upstream, and checks the rules once before exiting.
		if !p.dryRun {
			b, err := p.git.sdo("push", "--porcelain")
			if err == nil {
				log.Printf("git push out:\n%s", b)
				return nil
			}
			log.Printf("git push error: %v", err)
			log.Printf("git push out:\n%s", b)
		}
		before, err := p.git.revParse("HEAD")
		if err != nil {
			return err
		}
		log.Printf("pulling...")
		if b, err := p.git.sdo("pull", "--rebase"); err != nil {
			log.Printf("pull out:\n%s", b)
			return err
		}
		after, err := p.git.revParse("HEAD")
		if err != nil {
			return err
		}

		if err := p.checkRules(before, after); err != nil {
			return err
		}

		// it doesn't make sense for --dry-run to repeat
		if p.dryRun {
			return nil
		}
	}
	return fmt.Errorf("giving up after %d retries", p.retries)
}

// checkRules returns an error if at list one file doesn't match one of the "ignoreFiles" patterns.
func (p *pushup) checkRules(before, after string) error {
	r := fmt.Sprintf("%s..%s", before, after)
	log.Printf("before..after: %s", r)

	n, err := p.git.do("diff", "--name-only", r)
	if err != nil {
		return err
	}

	m, err := ignore.CompileIgnoreLines(p.ignoreFiles...)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(n))
	for scanner.Scan() {
		log.Println(scanner.Text())
		f := scanner.Text()
		if !m.MatchesPath(f) {
			return fmt.Errorf("%q doesn't match ignore patterns %q", f, p.ignoreFiles)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
