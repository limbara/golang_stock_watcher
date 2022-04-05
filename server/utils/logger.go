package utils

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Logger() (*zap.Logger, error) {
	file, err := getLogFile()
	if err != nil {
		err = fmt.Errorf("Logger getErrorFile Error : %w", err)
		return nil, err
	}

	stdPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	enc := zap.NewProductionEncoderConfig()
	enc.TimeKey = "time"
	enc.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000Z0700")
	fileEncoder := zapcore.NewJSONEncoder(enc)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, file, stdPriority),
		zapcore.NewCore(consoleEncoder, os.Stdout, lowPriority),
	)

	return zap.New(core), nil
}

func getLogFile() (*os.File, error) {
	appEnv, err := LoadAppEnv()
	if err != nil {
		return nil, err
	}

	path := "./storage/error"
	if appEnv.LogPath != "" {
		path = appEnv.LogPath
	}

	currentTime := time.Now()
	filePath := fmt.Sprintf("%s/%s.log", path, currentTime.Format("2006-01-02"))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0655) // Create your file
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 644)
	if err != nil {
		err = fmt.Errorf("Logger OpenFile Error : %w", err)
		return nil, err
	}

	return file, nil
}
