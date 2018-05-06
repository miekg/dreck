package log

import (
	"fmt"
	golog "log"
)

const d = "dreck: "

// logf calls log.Printf prefixed with level.
func logf(level, format string, v ...interface{}) {
	s := level + d + fmt.Sprintf(format, v...)
	golog.Print(s)
}

// log calls log.Print prefixed with level.
func log(level string, v ...interface{}) { s := level + d + fmt.Sprint(v...); golog.Print(s) }

// Info is equivalent to log.Print, but prefixed with "[INFO] ".
func Info(v ...interface{}) { log(info, v...) }

// Infof is equivalent to log.Printf, but prefixed with "[INFO] ".
func Infof(format string, v ...interface{}) { logf(info, format, v...) }

// Warning is equivalent to log.Print, but prefixed with "[WARNING] ".
func Warning(v ...interface{}) { log(warning, v...) }

// Warningf is equivalent to log.Printf, but prefixed with "[WARNING] ".
func Warningf(format string, v ...interface{}) { logf(warning, format, v...) }

// Error is equivalent to log.Print, but prefixed with "[ERROR] ".
func Error(v ...interface{}) { log(err, v...) }

// Errorf is equivalent to log.Printf, but prefixed with "[ERROR] ".
func Errorf(format string, v ...interface{}) { logf(err, format, v...) }

const (
	err     = "[ERROR] "
	warning = "[WARNING] "
	info    = "[INFO] "
)
