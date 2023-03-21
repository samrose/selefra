package logger

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
)

// Logger impl the hclog.Logger interface
var _ hclog.Logger = (*Logger)(nil)

func (l *Logger) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel:
		return
	case hclog.Trace:
		l.Trace(msg, args...)
	case hclog.Debug:
		l.Debug(msg, args...)
	case hclog.Info:
		l.Info(msg, args...)
	case hclog.Warn:
		l.Warn(msg, args...)
	case hclog.Error:
		l.Error(msg, args...)
	}
}

func (l *Logger) Trace(msg string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(msg, args...))
}

func (l *Logger) IsTrace() bool {
	return false
}

func (l *Logger) IsDebug() bool {
	return l.logger.Core().Enabled(zapcore.DebugLevel)
}

func (l *Logger) IsInfo() bool {
	return l.logger.Core().Enabled(zapcore.InfoLevel)
}

func (l *Logger) IsWarn() bool {
	return l.logger.Core().Enabled(zapcore.WarnLevel)
}

func (l *Logger) IsError() bool {
	return l.logger.Core().Enabled(zapcore.ErrorLevel)
}

func (l *Logger) ImpliedArgs() []interface{} {
	return nil
}

func (l *Logger) With(args ...interface{}) hclog.Logger {
	return l
}

func (l *Logger) Name() string {
	return "selefra-cli"
}

func (l *Logger) Named(name string) hclog.Logger {
	return l
}

func (l *Logger) ResetNamed(name string) hclog.Logger {
	return l
}

func (l *Logger) SetLevel(level hclog.Level) {
	return
}

func (l *Logger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(l.StandardWriter(opts), "", 0)
}

func (l *Logger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return os.Stdin
}
