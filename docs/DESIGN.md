# Design/Internals

The regression test framework is written in Go with the bulk
The main entry point for running local tests is implemented in
`rtf.go`.

The core of the framework is implemented in the `local` module.
This file provided two objects, `Test` and `Group` which handle
the traversal, naming and execution of tests.

## Results

Apart from the log files (see below) and the output on the console,
the regression test framework also produces two CSV files. `TESTS.csv`
contains a line per test and `SUMMARY.csv` contains a one line summary
of all tests run.

The CSV files are linked by a UUID, which can be thought of as a key.
The structure should make it very easy to store the test results in a
simple database containing two tables, one with test summaries and
another with all test results.

## Logging

For logging we utilise a custom logging package in `./logger`.
We use three logging backends:

- A file based one which writes pretty much any output to a log file.
- A file based one per tests which logs the same as the above but on a
  per test basis.
- A console based logger for printing progress on the terminal.

The file based loggers are pretty standard, while the console based
logger is customised to provided colourful output.

We define several additional log levels:

- Two log-levels to capture stdout/stderr from external commands which
  are executed.  Using separate log-levels allows us to annotate the
  output accordingly.

- Separate log levels for test results. The summary of passed, failed
  and skipped tests are logged to different log levels. This allows
  finer control of what and how the results are displayed on the
  console.
