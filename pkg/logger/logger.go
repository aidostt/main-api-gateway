package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
)

func Debug(msg ...interface{}) {
	logrus.Debug(msg...)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Info(msg ...interface{}) {
	logrus.Info(msg...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

// getStackTrace returns a formatted stack trace with a given depth
func getStackTrace(depth int) string {
	stackBuf := make([]uintptr, depth)
	length := runtime.Callers(3, stackBuf[:]) // skip 3 levels (Callers, getStackTrace, Error/Errorf)
	stackBuf = stackBuf[:length]

	stackTrace := ""
	for _, pc := range stackBuf {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pc)
		stackTrace += fmt.Sprintf("%s:%d %s\n", file, line, fn.Name())
	}

	return stackTrace
}

func Error(msg ...interface{}) {
	stackTrace := getStackTrace(3)
	logrus.Error(fmt.Sprint(msg...) + "\n" + stackTrace)
}

func Errorf(format string, args ...interface{}) {
	stackTrace := getStackTrace(3)
	logrus.Errorf(format+"\n"+stackTrace, args...)
}
