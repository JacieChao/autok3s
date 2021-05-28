package common

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// NewLogger returns logger for autok3s
func NewLogger(debug bool, w *os.File) *logrus.Logger {
	logger := logrus.New()
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if w != nil {
		mw := io.MultiWriter(os.Stderr, w)
		logger.SetOutput(mw)
	}

	return logger
}

// GetLogPath returns default log file path
func GetLogPath() string {
	return filepath.Join(CfgPath, "logs")
}

// GetLogFile returns default log file for provider
func GetLogFile(name string) (logFile *os.File, err error) {
	logFilePath := filepath.Join(GetLogPath(), name)
	// check file exist
	_, err = os.Stat(logFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		logFile, err = os.Create(logFilePath)
	} else {
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	}
	return logFile, err
}
