RT Framework
============

**A regression testing framework**

This project contains a generic, cross-platform regression test
framework.

The generic regression test framework is implemented in Go whereas
the test are written as shell scripts. The library code is in the `rt`
directory and some common utilities and helper programs are contained in
the `utils` directory.

-   `rtf` - a local test runner

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

On a Mac and Linux, this should pretty much just work out of the box.

On Windows, you need to have form of `bash`.
The tests must be run from a `bash` shell and `bash.exe` must be
in your path. The simplest way is to install via
[chocolatey](https://chocolatey.org/)

```
choco install git
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
rtf -x run
```

This executes the tests with `-x` and thus logs all commands executed.

For a CI system, where the output is displayed on a web page use:
```
rtf -x -vvv run
```

This prints the same information logged to the log file to the console.
