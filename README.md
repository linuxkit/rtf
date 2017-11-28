RT Framework
============

**A regression testing framework**

This project contains a generic, cross-platform regression test
framework.

The generic regression test framework is implemented in Go whereas
the test are written as shell scripts. The library code is in the `rt`
directory and some common utilities and helper programs are contained in
the `utils` directory.

- `rtf` - a local test runner

For more details, see the documentation in `./docs/USER_GUIDE.md`.

## Installation

```
go get -u github.com/linuxkit/rtf
```

## Development

To run the test suite please use:
```
make test
```

## Prerequisites

On a Mac and Linux, this should pretty much just work out of the
box. The tests are written in `sh` and expect it to be installed in
`/bin/sh`. Some optional utilities (see [`./bin`](./bin)) are written
in `python` and if your tests use them, you need to have python
installed.

On Windows, it depends on how your tests are written. If they are
written as `powershell` scripts, `rtf` should just work. If they are
written as shell scripts, you need to have the MSYS2 variant of `bash`
installed and `bash.exe` must be in your path. The simplest way to
install it is to install `git` via
[chocolatey](https://chocolatey.org/). Note, neither `bash` from WSL
nor cygwin is currently supported.

```
choco install git
```

If your tests use the optional utilities in, you also need to install `python`:
```
choco install python
```

## Quickstart

The regression test framework allows running tests on a local host (or
inside a VM) as well as against a suitably configured remote host.

If you don't have the source code in your `GOPATH`, you may have to
set the `RT_ROOT` environment variable to point to it.

To run tests locally, simply execute the `rtf run` command. It will
executed all the test cases in the supplied cases directory. This
defaults to `./cases`

To list all current tests run `rtf list`, or to get a one line
summary for each test use `rtf info`.

When running tests, by default a line per test is printed on the console
with a pass/fail indication. Detailed logs, by default, are stored in
`./_results/<UUID>/`. In that directory, `TESTS.log` contains detailed
logs of all tests, `TESTS.csv` contains a line per test and
`SUMMARY.csv` contains a one line summary of the all tests run. The
directory also contains a log file for each tests, with the same
contents as `TESTS.log`.

If you prefer a bit more information in the log files use:
```
rtf -v run -x
```

This executes the tests with `-x`, logging all commands executed to `stderr`;
and with `-v`, causing `stderr` to be displayed to the console.

For a CI system, where the output is displayed on a web page use:
```
rtf -vvv run -x
```

This prints the same information logged to the log file to the console.
