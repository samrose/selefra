package logger

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"go.uber.org/zap"
)

// SelefraSDKClientLogger is the implement of schema.ClientLogger
type SelefraSDKClientLogger struct {
	wrappedLog *Logger
}

// To match the log system on the SDK, connect the logs of the two systems
var _ schema.ClientLogger = (*SelefraSDKClientLogger)(nil)

func NewSchemaLogger(wrappedLog ...*Logger) *SelefraSDKClientLogger {
	if len(wrappedLog) == 0 {
		wrappedLog = append(wrappedLog, defaultLogger)
	}
	return &SelefraSDKClientLogger{
		wrappedLog: wrappedLog[0],
	}
}

func (s *SelefraSDKClientLogger) Debug(msg string, fields ...zap.Field) {
	s.wrappedLog.Debug(msg, fields)
}

func (s *SelefraSDKClientLogger) DebugF(msg string, args ...any) {
	s.wrappedLog.Debug(msg, args)
}

func (s *SelefraSDKClientLogger) Info(msg string, fields ...zap.Field) {
	s.wrappedLog.Info(msg, fields)
}

func (s *SelefraSDKClientLogger) InfoF(msg string, args ...any) {
	s.wrappedLog.Info(msg, args)
}

func (s *SelefraSDKClientLogger) Warn(msg string, fields ...zap.Field) {
	s.wrappedLog.Warn(msg, fields)
}

func (s *SelefraSDKClientLogger) WarnF(msg string, args ...any) {
	s.wrappedLog.Warn(msg, args)
}

func (s *SelefraSDKClientLogger) Error(msg string, fields ...zap.Field) {
	s.wrappedLog.Error(msg, fields)
}

func (s *SelefraSDKClientLogger) ErrorF(msg string, args ...any) {
	s.wrappedLog.Error(msg, args)
}

func (s *SelefraSDKClientLogger) Fatal(msg string, fields ...zap.Field) {
	s.wrappedLog.Fatal(msg, fields)
}

func (s *SelefraSDKClientLogger) FatalF(msg string, args ...any) {
	s.wrappedLog.Fatal(msg, args)
}

func (s *SelefraSDKClientLogger) LogDiagnostics(prefix string, d *schema.Diagnostics) {
	if d == nil {
		return
	}

	for _, diagnostic := range d.GetDiagnosticSlice() {

		var msg string
		if prefix != "" {
			msg = fmt.Sprintf("%s, %s", prefix, diagnostic.Content())
		} else {
			msg = diagnostic.Content()
		}

		switch diagnostic.Level() {
		case schema.DiagnosisLevelTrace:
			s.Debug(msg)
		case schema.DiagnosisLevelDebug:
			s.Debug(msg)
		case schema.DiagnosisLevelInfo:
			s.Info(msg)
		case schema.DiagnosisLevelWarn:
			s.Warn(msg)
		case schema.DiagnosisLevelError:
			s.Error(msg)
		case schema.DiagnosisLevelFatal:
			s.Fatal(msg)
		}
	}
}
