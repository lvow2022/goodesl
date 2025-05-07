package goodesl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var logger = log.New(os.Stdout, "[goodesl] ", log.LstdFlags)
var debugMode = false

func SetDebug(debug bool) {
	debugMode = debug

}

func Debugf(format string, args ...interface{}) {
	if debugMode {
		_, file, line, ok := runtime.Caller(1)
		prefix := ""
		if ok {
			prefix = fmt.Sprintf("%s:%d: ", filepath.Base(file), line)
		}
		logger.Output(2, prefix+fmt.Sprintf(format, args...))
	}
}

func Debug(val any) {
	if debugMode {
		_, file, line, ok := runtime.Caller(1)
		prefix := ""
		if ok {
			prefix = fmt.Sprintf("%s:%d: ", filepath.Base(file), line)
		}
		logger.Output(2, prefix+fmt.Sprint(val))
	}
}
