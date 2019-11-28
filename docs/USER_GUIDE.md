# User's guide

Tests are located in a directory, defaulting to `./cases`

Within a project tests may be grouped by placing them in a
common sub-directory (forming a _Test Group_).  The order in which
tests and test groups are executed is determined by alphabetical order
and one may prefix a directory with a number to force a order. Each
test is given a unique name based on the place in the directory
hierarchy where it is located.  For the naming of tests, any
prefix-numbers, used to control the order of execution, are
stripped. For example, a test located in:

```
cases/foo/bar/001_example_test/
```

will be named:

```
foo.bar.example_test
```

While traversing the directory tree, directories starting with `_` are
ignored.

## Run control

Apart from the simple command line examples given in the README,
the regression test framework provides fine grained control over which 
tests are run.  The primary tool for this are _labels_.

A test may define that a specific label "foobar" must be present for
it to be run or, by prefixing the label name with a '!', that the test
should not be run if a label is defined. The labels are defined with
`test.sh` or `test.ps1` (see below).

The regression test framework is completely agnostic to which labels
are used for a given set of test cases, though it defines a number of
labels based on the OS it is executing on. See the output of `rtf
list` on your host.

A good strategy for using labels is to not define any labels for tests
which should always be executed (e.g. on every CI run).  Then use
labels to control execution for tests which are specific for a given
platform (e.g. `osx` or `win` for OS X and Windows installer tests).
If a particualr test is known not to run on a given platform you can
use, e.g. `!win`, to indicate it.  Finally, define separate labels,
e.g., long running tests or extensive tests which should only be run
on release candidates.

You can control labels for executing tests by using the `-l` flag.
For example, if you have some long running tests which you do not wish
to execute on every CI run and have them marked with a label `long`,
then you can execute them with:

```
./rtf -l long run
```

You can see which tests would get executed using the `-l` flag for the
`list` command as well:

```
./rtf -l long list
```

In addition to control which tests are run via labels it is also
possible to specify an individual test or a group name on the command
line to just run these test (subject to labels).  Here are two
examples:

```
./rtf run foo.bar.example_test
./rtf run foo.bar
```

The first runs a single test, while the second is running all tests
within the `bar` group. Note, that this is currently implemented as
a simple prefix match, so, if you have tests such as `foo.bar` and
`foo.bar_baz` and use `./rtf run foo.bar`, it will execute both
`foo.bar` and `foo.bar_baz`.

## Parallel Execution

You may have tests execute in parallel by using the `-p` flag:
```
./rtf run -p
```

All tests across all groups will run in parallel.  You should ensure
that individual tests have no dependencies on each other since you
cannot guarantee any test has completed before another has started


## Writing tests

Tests are simple scripts which return `0` on success and a non-zero
code on failure.  A special return code (`253` or `RT_TEST_CANCEL`)
can be use to indicate that the test was cancelled (for whatever
reason).  Each test must be located in its own sub-directory (together
with any files it may require).

Currently, a test is a simple shell script called `test.sh` or
`test.ps1`. On Windows, `test.ps1` is chosen in preference over
`test.sh` if both are present. `test.sh` should also work provided
that MSYS2 `bash` is installed. On Unix type systems `test.sh` is
chosen over `test.ps1`, but `test.ps1` should also work provided that
that `powershell` is installed.

There are template [`test.sh`](../etc/templates/test.sh) and
[`test.ps1`](../etc/templates/test.ps1) files which can be used for
writing tests. A test script contains a number of special comments
(`SUMMARY`, `LABELS`, `REPEAT`, `ISSUE`, and, `AUTHOR`) which are used
by the regression test framework. The `SUMMARY` line should contain a
*short* summary of what the test does. The `LABELS` is a (optional)
list of labels to control when a test should be executed.  `AUTHOR`
should contain name and email address of the test author.  If a test
is from multiple authors, multiple `AUTHOR` lines can be used. `ISSUE`
can be used to link a test to one or more issues on a bug tracker,
i.e., this test exists because of these issues. To link to multiple
issues, multiple `ISSUE` lines can be used.  Finally, the `REPEAT`
line may contain a single number to indicate that a test should be
executed multiple times.  The `REPEAT` line may also contain
`<label>:<number>` entries to runtest multiple times if a label is
present.

Optionally, if a test is a benchmark, you can echo the benchmark
result in `test.sh` or `test.ps1` in a line *starting* with
`RT_BENCHMARK_RESULT:`. The remainder of that line will then be logged
in the results.

A few guidelines for writing tests:

- A test should always clean up whatever is created during test
  execution. This includes containers, docker images, and files. The
  template contains a `clean_up()` function which can be used for this
  purpose.

- An individual tests should not rely on a artefact left behind by a
  previous test (even if one can control the order in which tests are
  executed)

- Tests should be self contained, i.e. they should not rely on any
  files outside the directory they are located in.

The regression test framework currently passes the following
environment variables into a test script:

- `RT_ROOT`: Points to the root of the test framework.

- `RT_LIB`: Points to the common shell library found in
  [`lib.sh`](../lib/lib.sh) or [`lib.ps1`](../lib/lib.ps1)

- `RT_UTILS`: Points to the directory where the helper applications
  are available

- `RT_PROJECT_ROOT`: Points to the root of the project. This can be
  used, e.g. to source common shell functions defined by a project.

- `RT_OS`, `RT_OS_VER`: OS and OS version information. `RT_OS` is one
  of `osx` or `win`.

- `RT_RESULTS`: Points to the directory where results data is stored.
  Can be used in conjunction with `RT_TEST_NAME` to store additional
  data, e.g., benchmark results.

- `RT_TEST_NAME`: Name of the test being run. A test implementation
  can use this in conjunction with `RT_RESULTS` to store additional
  data, e.g., benchmark results.

- `RT_LABELS`: A colon separated list of labels defined when the test
  is run. Can be used by the test to trigger different behaviour based
  on a Labels presence, e.g., run longer for release
  tests. `./lib/lib.sh` provides a shell function, `rt_label_set`, to
  check if a label is set.

Users can specify additional environment variables using the `-e` or
`--env` command line option to `rtf`.  This may be useful for
scenarios where `rtf` is executed remotely.


### General utilities for writing tests

Under `rt/lib/utils` are a number of small, cross-platform utilities
which may be useful when writing tests (the environment variable 
`RT_UTILS` points to the directory):

- `rt-filegen`: A small standalone utility to create a file of fixed
  size with random content.

- `rt-filerandgen`: A variation of the above. The filesize is with a
  set maximum.

- `rt-filemd5`: Returns the MD5 checksum of a file.

- `rt-crexec`: A utility to execute multiple commands concurrently in a
  random order.

- `rt-urltest`: A utility like curl, with retry and optional `grep` for a
  keyword on the result.

- `rt-elevate.exe`: A utility to run commands as administrator on Windows


## Creating a new Test Group

Any directory containing sub-directories with tests under the
top-level project directory forms a test group.  The execution of a
group can be customised by defining a `group.sh` or `group.ps1` file
inside the group directory.  Like with tests, the group script may
provide a `SUMMARY` and a set of `LABELS`.

If a group directory contains a `group.sh`/`group.ps1` file, it is
executed with the `init` argument before the first test of the group
is executed, and with the `deinit` argument after the last test of the
group was executed.  Test writers can thus place group specific
initialisation code into `group.sh`/`group.ps1`. The group script is
executed with the same environment variables set as `test.sh` scripts.

There are template [`group.sh`](../etc/templates/group.sh) and
[`group.ps1`](../etc/templates/group.ps1) files.

The top-level `group.sh`/`group.ps1` should also create a
`VERSION.txt` file in `RT_RESULTS`, containing some form of version
information.  If the tests are run against a local build, this could
be the git sha value, or when run as part of CI the version of the
build being tested etc.

In addition, the top-level group may also contain two optional
scripts, `pre-test.sh` and `post-test.sh` (or Powershell equivalent),
which are executed before and after each test is run.  Both get the
test name as first argument and `post-test.sh` gets passed the test
result as the second argument.  The idea is that these scripts may be
used to collect additional logging or collect debug information if a
test fails.  They can store the per test information in files prefixed
with `"${RT_RESULT}/$1"`.
