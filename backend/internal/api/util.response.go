package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

func Util_Response_SendJsonError(w http.ResponseWriter, elapsed string, err logging.StandardError) {
	if err.Function == "" {
		err.Function = Util_GetFunctionName()
	}
	if err.LineNumber == 0 {
		err.LineNumber = Util_GetLineNumber()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	response := JSONResponse{
		Status:  "error",
		Elapsed: elapsed,
		Error:   &err,
	}

	// Log the error using the logging package
	logMsg := err.Message

	if err.HelpText != "" {
		logMsg += "\n" + err.HelpText
	}

	if err.Details != nil {
		if detailsStr, ok := err.Details.(string); ok {
			logMsg += "\n" + detailsStr
		} else if detailsMap, ok := err.Details.(map[string]any); ok {
			if jsonBytes, jsonErr := json.MarshalIndent(detailsMap, "", "  "); jsonErr == nil {
				logMsg += "\n" + string(jsonBytes)
			} else {
				logMsg += "\n" + fmt.Sprintf("%v", detailsMap)
			}
		} else {
			logMsg += "\n" + fmt.Sprintf("%v", err.Details)
		}
	}

	logging.LOG.Error(logMsg)

	// Encode the response as JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Handle encoding error if necessary
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func Util_Response_SendJson(w http.ResponseWriter, statusCode int, response JSONResponse) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode the data into JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Handle encoding error if necessary
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func Util_GetLineNumber() int {
	if _, _, line, ok := runtime.Caller(2); ok {
		return line
	}
	return 0
}

func Util_GetFunctionName() string {
	if pc, _, _, ok := runtime.Caller(2); ok {
		return runtime.FuncForPC(pc).Name()
	}
	return ""
}
