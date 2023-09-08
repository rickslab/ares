package logger

import (
	"context"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewEntry(ctx context.Context, fields map[string]any) *logrus.Entry {
	return logrus.WithContext(ctx).WithFields(fields)
}

func getFileAndLine() (string, int) {
	for i := 3; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		if !strings.Contains(file, "logrus") {
			return file, line
		}
	}
	return "", 0
}
