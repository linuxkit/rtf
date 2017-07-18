package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

// LogLevel is The severity level of the logs
// 0 is the lowest severity
type LogLevel int

// LogDispatcher dispatches logs to multiple Loggers
type LogDispatcher interface {
	Log(level LogLevel, msg string)
	Register(name string, backend Logger)
	Unregister(name string)
}

// A Logger formats and writes log entries
type Logger interface {
	LogFormatter
	LogWriter
	Log(timestamp time.Time, level LogLevel, msg string)
	SetLevel(level LogLevel)
}

const (
	// LevelCritical represents the Critical Log Level
	LevelCritical = 100
	// LevelError represents the Error log level
	LevelError = 200
	// LevelWarning represents the Warning log level
	LevelWarning = 300
	// LevelInfo represents the Informational logging level
	LevelInfo = 400
	// LevelDebug represents the Debug log level
	LevelDebug = 500
	// LevelStderr represents the Stderr log level
	LevelStderr = LevelWarning + 6
	// LevelStdout represents the Stdout log level
	LevelStdout = LevelWarning + 7
	// LevelSkip represents the Skip log level
	LevelSkip = LevelWarning + 1
	// LevelPass represents the Pass log level
	LevelPass = LevelWarning + 2
	// LevelCancel represents the Cancel log level
	LevelCancel = LevelWarning + 3
	// LevelFail represents the Fail log level
	LevelFail = LevelWarning + 4
	// LevelSummary represents the Summary log level
	LevelSummary = LevelWarning + 5
)

// LevelNames maps LogLevels to a string representation of their names
var LevelNames = map[LogLevel]string{
	LevelCritical: "CRITICAL",
	LevelError:    "ERROR",
	LevelWarning:  "WARNING",
	LevelInfo:     "INFO",
	LevelDebug:    "DEBUG",
	LevelStderr:   "STDERR",
	LevelStdout:   "STDOUT",
	LevelSkip:     "SKIP",
	LevelPass:     "PASS",
	LevelCancel:   "CANCEL",
	LevelFail:     "FAIL",
	LevelSummary:  "SUMMARY",
}

type logDispatcher struct {
	Backends map[string]Logger
	sync.RWMutex
}

// Log dispatches a Log entry to each backend
func (d *logDispatcher) Log(level LogLevel, msg string) {
	d.RLock()
	defer d.RUnlock()
	timestamp := time.Now()
	for _, b := range d.Backends {
		b.Log(timestamp, level, msg)
	}
}

func (d *logDispatcher) Register(name string, backend Logger) {
	d.Lock()
	defer d.Unlock()
	d.Backends[name] = backend
}

func (d *logDispatcher) Unregister(name string) {
	d.Lock()
	defer d.Unlock()
	delete(d.Backends, name)
}

// NewLogDispatcher returns a new LogDispatcher with the provided backend Loggers
func NewLogDispatcher(backends map[string]Logger) LogDispatcher {
	return &logDispatcher{Backends: backends}
}

// LogFormatter formats log entries in to a string
type LogFormatter interface {
	Format(timestamp time.Time, level LogLevel, msg string) string
}

// LogWriter writes log entries somewhere
type LogWriter interface {
	Write(entry string)
}

type logger struct {
	LogFormatter
	LogWriter
	levelFilter LogLevel
}

// Log formats a log entry and writes it
func (l *logger) Log(timestamp time.Time, level LogLevel, msg string) {
	if level <= l.levelFilter {
		entry := l.Format(timestamp, level, msg)
		l.Write(entry)
	}
}

// SetLevel sets maximum logging level
func (l *logger) SetLevel(level LogLevel) {
	l.levelFilter = level
}

type ioLogWriter struct {
	writer io.Writer
}

// Write writes to the underlying io.Writer
func (i ioLogWriter) Write(entry string) {
	n, err := io.WriteString(i.writer, entry)
	if err != nil {
		panic(err)
	}
	if n == 0 {
		panic("Wrote 0 bytes to file")
	}
}

// ColourMap maps logLevels to colorizing functions
type ColourMap map[LogLevel]func(...interface{}) string

type consoleLogFormatter struct {
	coloured  bool
	colourMap ColourMap
}

var defaultColourMap = ColourMap{
	LevelCritical: color.New(color.FgRed, color.Bold).SprintFunc(),
	LevelError:    color.New(color.FgRed).SprintFunc(),
	LevelWarning:  color.New(color.FgYellow).SprintFunc(),
	LevelInfo:     color.New(color.FgBlue).SprintFunc(),
	LevelDebug:    color.New(color.FgWhite).SprintFunc(),
	LevelStderr:   color.New(color.FgRed).SprintFunc(),
	LevelSkip:     color.New(color.FgYellow, color.Bold).SprintFunc(),
	LevelPass:     color.New(color.FgGreen, color.Bold).SprintFunc(),
	LevelCancel:   color.New(color.FgMagenta, color.Bold).SprintFunc(),
	LevelFail:     color.New(color.FgRed, color.Bold).SprintFunc(),
}

// Format formats the log for writing to console
func (c consoleLogFormatter) Format(timestamp time.Time, level LogLevel, msg string) string {
	l := fmt.Sprintf("[%-8s]", LevelNames[level])
	if c.coloured {
		if cFunc, ok := c.colourMap[level]; ok {
			l = cFunc(l)
		}
	}
	var s string
	switch level {
	case LevelPass, LevelFail, LevelSkip, LevelSummary, LevelCancel:
		s = fmt.Sprintf("%s %s\n", l, msg)
	default:
		s = fmt.Sprintf("%s %s: %s\n", l, timestamp.Format(time.RFC3339Nano), msg)
	}
	return s
}

// NewConsoleLogger returns a new logger that logs to stderr in console log format
func NewConsoleLogger(coloured bool, colourMap *ColourMap) Logger {
	var clf LogFormatter
	if coloured {
		if colourMap == nil {
			colourMap = &defaultColourMap
		}
		clf = consoleLogFormatter{
			coloured:  coloured,
			colourMap: *colourMap,
		}
	} else {
		clf = consoleLogFormatter{}
	}

	return &logger{
		clf,
		ioLogWriter{
			writer: os.Stderr,
		},
		LevelWarning,
	}
}

// NewFileLogger returns a new logger that logs to file in a console log format
func NewFileLogger(f *os.File) Logger {
	return &logger{
		consoleLogFormatter{
			coloured: false,
		},
		ioLogWriter{
			writer: f,
		},
		LevelWarning,
	}
}
