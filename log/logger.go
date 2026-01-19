package log

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func Init(appName string) {
	logDir := filepath.Join(".", "logs")
	_ = os.MkdirAll(logDir, 0755)

	fileName := filepath.Join(
		logDir,
		appName+"-"+time.Now().Format("2006-01-02")+".log",
	)

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    20, // MB
		MaxBackups: 7,
		MaxAge:     14, // days
		Compress:   true,
	})

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		MessageKey:    "msg",
		CallerKey:     "caller",
		StacktraceKey: "stack",
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(
		encoder,
		writer,
		zap.InfoLevel,
	)

	Log = zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}
