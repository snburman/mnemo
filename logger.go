package mnemo

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

const (
	Err LogLevel = iota
	Debug
	Info
	Warn
	Fatal
	Panic
)

var logger = newcharmLogger("Mnemo")

type (
	LogLevel int
	Logger   interface {
		Info(string)
		Debug(string)
		Warn(string)
		Error(string)
		Fatal(string)
	}
	charmLogger struct {
		*log.Logger
	}
)

// newcharmLogger wraps charmbracelet/log to implement the Logger interface
func newcharmLogger(prefix string) *charmLogger {
	l := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          prefix,
	})
	return &charmLogger{l}
}

func (l *charmLogger) Info(msg string) {
	l.Logger.Info(msg)
}

func (l *charmLogger) Debug(msg string) {
	l.Logger.Debug(msg)
}

func (l *charmLogger) Warn(msg string) {
	l.Logger.Warn(msg)
}

func (l *charmLogger) Error(msg string) {
	l.Logger.Error(msg)
}

func (l *charmLogger) Fatal(msg string) {
	l.Logger.Fatal(msg)
}
