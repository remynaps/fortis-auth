package logging

import (
	"context"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"gitlab.com/gilden/fortis/correlationID"
)

var Logger = logrus.New()

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
