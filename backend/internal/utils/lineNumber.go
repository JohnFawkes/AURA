package utils

import "runtime"

func GetLineNumber() int {
	if _, _, line, ok := runtime.Caller(2); ok {
		return line
	}
	return 0
}

func GetFunctionName() string {
	if pc, _, _, ok := runtime.Caller(2); ok {
		return runtime.FuncForPC(pc).Name()
	}
	return ""
}
