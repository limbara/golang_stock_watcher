package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Singleton?
var loggerInstance *zap.Logger

// Get CustomLogger stored in global variable. panic if nil
func Logger() *zap.Logger {
	if loggerInstance == nil {
		panic(errors.New("logger was't BootstrapDB correctly"))
	}

	return loggerInstance
}

// Set CustomLogger stored in global variable
func BootstrapLogger(logPath string) error {
	file, err := getLogFile(logPath)
	if err != nil {
		return fmt.Errorf("Error BootstrapLogger : %w", err)
	}

	loggerInstance = createZapLogger(file)

	return nil
}

func createZapLogger(file *os.File) *zap.Logger {
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

	return zap.New(core)
}

func getLogFile(logPath string) (*os.File, error) {
	if logPath == "" {
		return nil, errors.New("empty logPath")
	}

	currentTime := time.Now()
	filePath := fmt.Sprintf("%s/%s.log", logPath, currentTime.Format("2006-01-02"))

	// Create Log Path Directory if not exist
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.MkdirAll(logPath, 0755)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		err = fmt.Errorf("Logger OpenFile Error : %w", err)
		return nil, err
	}

	return file, nil
}
