# Contributing

Thank you for your interest in making [FerretDB](https://github.com/FerretDB/FerretDB) better!

## Finding something to work on

We are interested in all contributions, big or small, in code or documentation.
But unless you are fixing a very small issue like a typo,
we kindly ask you first to [create an issue](https://github.com/FerretDB/github-actions/issues/new/choose),
to leave a comment on an existing issue if you want to work on it,
or to [join our Slack chat](https://github.com/FerretDB/FerretDB/README.md#community) and leave a message for us there.
This way, you will get help from us and avoid wasted efforts if something can't be worked on right now
or someone is already working on it.

You can find a list of good issues for first-time contributors [there](https://github.com/FerretDB/github-actions/contribute).

## Setting up the environment

### Requirements

**TODO**

### Making a working copy

Fork the [FerretDB/github-actions repository on GitHub](https://github.com/FerretDB/github-actions/fork).
To have all the tags in the repository and what they point to, copy all branches by removing checkmark for `copy the main branch only` before forking.

After forking `FerretDB/github-actions` on GitHub, you can clone the repository:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/github-actions.git
cd github-actions
git remote add upstream https://github.com/FerretDB/github-actions.git

To run development commands, you should first install the [`task`](https://taskfile.dev/) tool.
You can do this by changing the directory to `tools` (`cd tools`) and running `go generate -x`.
That will install `task` into the `bin` directory (`bin/task` on Linux and macOS, `bin\task.exe` on Windows).
You can then add `./bin` to `$PATH` either manually (`export PATH=./bin:$PATH` in `bash`)
or using something like [`direnv` (`.envrc` files)](https://direnv.net),
or replace every invocation of `task` with explicit `bin/task`.
You can also [install `task` globally](https://taskfile.dev/#/installation),
but that might lead to the version skew.

With `task` installed,
you should install development tools with `task init`.

If something does not work correctly,
you can clean whole cache and re-install all development 
tools with `task init-clean`.

You can see all available `task` tasks with `task -l`.

## Contributing code

### Commands for contributing code

With `task` installed (see above), you may do the following:

1. Format code with `task fmt`.
2. Run linters against code with `task lint`.
3. Run godocs server at 127.0.0.1:6060 to check documentation formatting.

### Code overview

**TODO**

### Code style and conventions

Above everything else, we value consistency in the source code.
If you see some code that doesn't follow some best practice but is consistent,
please keep it that way;
but please also tell us about it, so we can improve all of it.
If, on the other hand, you see code that is inconsistent without apparent reason (or comment),
please improve it as you work on it.

Our code most of the standard Go conventions,
documented on [CodeReviewComments wiki page](https://github.com/golang/go/wiki/CodeReviewComments).
Some of our idiosyncrasies:

[//]: # ("TODO: Is this needed here vvv")
1. We use type switches over BSON types in many places in our code.
   The order of `case`s follows this order: <https://pkg.go.dev/github.com/FerretDB/FerretDB/internal/types#hdr-Mapping>
   It may seem random, but it is only pseudo-random and follows BSON spec: <https://bsonspec.org/spec.html>

### Submitting code changes

Before submitting a pull request, please make sure that:

1. Tests are added for new functionality or fixed bugs.
2. `task all` passes.
3. Comments are added or updated for all new and changed top-level declarations (functions, types, etc).
   Both exported and unexported declarations should have comments.
4. Comments are rendered correctly in the `task godocs` output.
