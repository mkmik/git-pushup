/* The git-pushup tool is wrapper around git push that gracefully handles push conflicts:

	Updates were rejected because the remote contains work that you do not have locally.

There is a good reason why you get that error: changes must be merged and merge conflicts
go beyond simple lexical overlaps, as repository states my have semantic conflicts even
if changes happen in non-overlapping areas of the repository.

That said, there are some cases where you know for sure that if some particular files have
changed (or changed in some way) you don't really care and you can just automatically merge
and push without any further testing (or with a different, possibly cheaper test suite).

In that case, this tool is for you. It works like this:

	$ git clone $GITURL
	$ do_something
	$ git commit -m "blah"
	$ git-pushup -I '*.md'

Since the command is called git-pushup, it can be also invoked as "git pushup" if
the git-pushup binary is present in the PATH or in the GITEXECPATH directories.
*/
package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// NewPushupWithFlags returns a new pushup instance which will be initialized
// once the flags are parsed.
func newPushupWithFlags(fs *pflag.FlagSet) *pushup {
	var p pushup
	p.registerPFlags(fs)
	return &p
}

// RegisterPFlags registers the necessary flags that will initialize this pushup instance.
func (p *pushup) registerPFlags(fs *pflag.FlagSet) {
	pflag.StringArrayVarP(&p.ignoreFiles, "ignore", "I", nil, "Ignore files matching this pattern. Can be repeated. Patterns follow the .gitignore syntax and semantics (see man gitignore).")
	pflag.IntVarP(&p.retries, "retries", "r", 10, "How many times the pull+push should be retried.")
	pflag.BoolVar(&p.rebase, "rebase", true, "Whether to pull --rebase.")
	pflag.StringVar(&p.gitDir, "git-dir", os.Getenv("GIT_DIR"), "Set the path to the repository. This can also be controlled by setting the GIT_DIR environment variable. It can be an absolute path or relative path to current working directory.")
	pflag.BoolVarP(&p.verbose, "verbose", "v", false, "Verbose logs")
	pflag.BoolVarP(&p.dryRun, "dry-run", "N", false, "Dry run: only simulate a push")
}

func main() {
	p := newPushupWithFlags(pflag.CommandLine)
	pflag.Parse()

	if err := p.run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
