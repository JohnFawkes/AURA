package logging

import (
	"fmt"
	"runtime"
	"strings"
)

// getFunctionLocation retrieves the location of the function that called the logging function,
// excluding any functions within the "internal/logging" package. It returns a formatted string
// containing the file name and the function name in the format "[fileName:functionName]".
// This is useful for logging purposes to identify the source of a log entry.
func getFunctionLocation() string {
	frame := runtime.Frame{}
	for i := 3; ; i++ { // Start at 2 to skip the Log function and its direct caller
		pcs := make([]uintptr, 1)
		n := runtime.Callers(i, pcs)
		if n == 0 {
			break // No more callers
		}

		f, _ := runtime.CallersFrames(pcs).Next()
		if !strings.Contains(f.File, "internal/logging") {
			frame = f
			break
		}
	}

	// Split the frame.Function into package and function name
	parts := strings.Split(frame.Function, ".")

	var functionName string
	if len(parts) > 1 {
		functionName = parts[1]
	} else {
		functionName = parts[0]
	}
	callerLocation := fmt.Sprintf("[%s]", functionName)

	return callerLocation
}
