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

After forking FerretDB on GitHub, you can clone the repository:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/FerretDB.git
cd FerretDB
git remote add upstream https://github.com/FerretDB/FerretDB.git
```

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


