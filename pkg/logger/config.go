package logger

import (
	"github.com/natefinch/lumberjack"
	"github.com/selefra/selefra/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"path/filepath"
	"strings"
)

type Config struct {
	Source              string `yaml:"source,omitempty" json:"source,omitempty"`
	FileLogEnabled      bool   `yaml:"file_log_enabled,omitempty" json:"file_log_enabled,omitempty"`
	ConsoleLogEnabled   bool   `yaml:"enable_console_log,omitempty" json:"enable_console_log,omitempty"`
	EncodeLogsAsJson    bool   `yaml:"encode_logs_as_json,omitempty" json:"encode_logs_as_json,omitempty"`
	Directory           string `yaml:"directory,omitempty" json:"directory,omitempty"`
	Level               string `yaml:"level,omitempty" json:"level,omitempty"`
	LevelIdentUppercase bool   `yaml:"level_ident_uppercase,omitempty" json:"level_ident_uppercase,omitempty"`
	MaxAge              int    `yaml:"max_age,omitempty" json:"max_age,omitempty"`
	ShowLine            bool   `yaml:"show_line,omitempty" json:"show_line,omitempty"`
	ConsoleNoColor      bool   `yaml:"console_no_color,omitempty" json:"console_no_color,omitempty"`
	MaxSize             int    `yaml:"max_size,omitempty" json:"max_size,omitempty"`
	MaxBackups          int    `yaml:"max_backups,omitempty" json:"max_backups,omitempty"`
	TimeFormat          string `yaml:"time_format,omitempty" json:"time_format,omitempty"`
	Prefix              string `yaml:"prefix,omitempty" json:"prefix"`
}

func (c *Config) EncodeLevel() zapcore.LevelEncoder {
	switch {
	case c.LevelIdentUppercase && c.ConsoleNoColor:
		return zapcore.CapitalLevelEncoder
	case c.LevelIdentUppercase && !c.ConsoleNoColor:
		return zapcore.CapitalColorLevelEncoder
	case !c.LevelIdentUppercase && c.ConsoleNoColor:
		return zapcore.LowercaseLevelEncoder
	case !c.LevelIdentUppercase && !c.ConsoleNoColor:
		return zapcore.LowercaseColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

func (c *Config) TranslationLevel() zapcore.Level {
	switch strings.ToLower(c.Level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (c *Config) GetEncoder() zapcore.Encoder {
	if c.EncodeLogsAsJson {
		return zapcore.NewJSONEncoder(c.GetEncoderConfig())
	}
	return zapcore.NewConsoleEncoder(c.GetEncoderConfig())
}

func (c *Config) GetEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    c.EncodeLevel(),
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func (c *Config) GetLogWriter(level string) zapcore.WriteSyncer {
	filename := filepath.Join(global.WorkSpace(), c.Directory, c.Source+".log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackups,
		LocalTime:  true,
		Compress:   false,
	}
	return zapcore.AddSync(lumberjackLogger)
}

func (c *Config) GetEncoderCore() []zapcore.Core {
	cores := make([]zapcore.Core, 0, 7)
	for level := c.TranslationLevel(); level <= zapcore.FatalLevel; level++ {
		cores = append(cores, zapcore.NewCore(c.GetEncoder(), c.GetLogWriter(c.TranslationLevel().String()), c.GetLevelPriority(level)))
	}
	return cores
}

func (c *Config) GetLevelPriority(level zapcore.Level) zap.LevelEnablerFunc {
	switch level {
	case zapcore.DebugLevel:
		return func(level zapcore.Level) bool {
			return level == zap.DebugLevel
		}
	case zapcore.InfoLevel:
		return func(level zapcore.Level) bool {
			return level == zap.InfoLevel
		}
	case zapcore.WarnLevel:
		return func(level zapcore.Level) bool {
			return level == zap.WarnLevel
		}
	case zapcore.ErrorLevel:
		return func(level zapcore.Level) bool {
			return level == zap.ErrorLevel
		}
	case zapcore.DPanicLevel:
		return func(level zapcore.Level) bool {
			return level == zap.DPanicLevel
		}
	case zapcore.PanicLevel:
		return func(level zapcore.Level) bool {
			return level == zap.PanicLevel
		}
	case zapcore.FatalLevel:
		return func(level zapcore.Level) bool {
			return level == zap.FatalLevel
		}
	default:
		return func(level zapcore.Level) bool {
			return level == zap.DebugLevel
		}
	}
}
