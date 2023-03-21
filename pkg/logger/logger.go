package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

// Logger Logs used in the CLI to save local logs
type Logger struct {
	logger *zap.Logger
}

func Default() *Logger {
	return defaultLogger
}

func DebugF(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func InfoF(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func ErrorF(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func FatalF(msg string, args ...any) {
	defaultLogger.Fatal(msg, args...)
}

var defaultLogger, _ = NewLogger(Config{
	FileLogEnabled:    true,
	ConsoleLogEnabled: false,
	EncodeLogsAsJson:  true,
	ConsoleNoColor:    true,
	Source:            "client",
	Directory:         "logs",
	// TODO Specifies the log level
	Level: "info",
})

func NewLogger(c Config) (*Logger, error) {
	// TODO The logs are stored in the current directory
	logDir := filepath.Join("./", c.Directory)
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(logDir, 0755)
	}
	if err != nil {
		return nil, nil
	}
	errorStack := zap.AddStacktrace(zap.ErrorLevel)

	development := zap.Development()

	logger := zap.New(zapcore.NewTee(c.GetEncoderCore()...), errorStack, development)

	if c.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	return &Logger{logger: logger}, nil
}
