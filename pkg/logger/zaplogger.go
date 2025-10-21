package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

func NewLogger(logPath string) *zap.Logger {
	_ = os.MkdirAll("./logs", 0755)

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath, 
		MaxSize:    50,      // MB before rotation
		MaxBackups: 0,       // keep all backups
		MaxAge:     0,       // never delete old logs
		Compress:   true,    // compress old logs
	})

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, zap.InfoLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

func InitGlobalLogger() {
	globalLogger = NewLogger("./logs/app.log")
	zap.ReplaceGlobals(globalLogger)
}

func Log() *zap.Logger {
	if globalLogger == nil {
		InitGlobalLogger()
	}
	return globalLogger
}

func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}
