package logging

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/gilden/fortis/correlationID"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
}

// Setup configures the logger based on options in the config.json.
func Setup(config *viper.Viper) error {
	Logger.Formatter = &logrus.TextFormatter{DisableColors: false}

	Logger.SetLevel(logrus.InfoLevel)
	// Set up logging to a file if specified in the config
	logFile := config.GetString("logging.file")
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		mw := io.MultiWriter(os.Stderr, f)
		Logger.Out = mw
	}
	return nil
}

func requestFields(r *http.Request) logrus.Fields {
	requestID, typeCheck := correlationID.FromContext(r.Context())
	if !typeCheck {
	}
	return logrus.Fields{
		"request-id": requestID,
		"url":        r.URL.Path,
		"method":     r.Method,
	}
}

func SetFormatter(formatter logrus.Formatter) {
	Logger.Formatter = (formatter)
}

func SetOutput(output io.Writer) {
	Logger.Out = (output)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Warning(args ...interface{}) {
	Logger.Warning(args...)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

func WithRequest(req *http.Request) *logrus.Entry {
	return Logger.WithFields(requestFields(req))
}

func WithContext(ctx context.Context) *logrus.Entry {
	requestID, typeCheck := correlationID.FromContext(ctx)
	if !typeCheck {
	}
	return Logger.WithFields(logrus.Fields{
		"request-id": requestID,
	})
}
