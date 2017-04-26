package logger

import (
	"os"
	"testing"
	"time"
)

func TestFileLogger(t *testing.T) {
	f, err := os.Create("foo.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	l := NewFileLogger(f)
	l.SetLevel(LevelDebug)
	generateLogEntries(l)
}

func TestConsoleLogger(t *testing.T) {
	l := NewConsoleLogger(true, nil)
	l.SetLevel(LevelDebug)
	generateLogEntries(l)
}

func generateLogEntries(l Logger) {
	l.Log(time.Now(), LevelCritical, "test")
	l.Log(time.Now(), LevelError, "test")
	l.Log(time.Now(), LevelWarning, "test")
	l.Log(time.Now(), LevelInfo, "test")
	l.Log(time.Now(), LevelDebug, "test")
	l.Log(time.Now(), LevelStderr, "test")
	l.Log(time.Now(), LevelStdout, "test")
	l.Log(time.Now(), LevelPass, "test")
	l.Log(time.Now(), LevelFail, "test")
	l.Log(time.Now(), LevelSkip, "test")
	l.Log(time.Now(), LevelCancel, "test")
	l.Log(time.Now(), LevelSummary, "test")
}
