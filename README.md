# git-pushup

The git-pushup tool is wrapper around git push that gracefully handles push conflicts:

```
	Updates were rejected because the remote contains work that you do not have locally.
```

There is a good reason why you get that error: changes must be merged and merge conflicts
go beyond simple lexical overlaps, as repository states my have semantic conflicts even
if changes happen in non-overlapping areas of the repository.

That said, there are some cases where you know for sure that if some particular files have
changed (or changed in some way) you don't really care and you can just automatically merge
and push without any further testing (or with a different, possibly cheaper test suite).

In that case, this tool is for you. It works like this:

```console
$ git clone $GITURL
$ do_something
$ git commit -m "blah"
$ git-pushup -I '*.md'
```

Since the command is called git-pushup, it can be also invoked as `git pushup` if
the git-pushup binary is present in the PATH or in the GITEXECPATH directories.
