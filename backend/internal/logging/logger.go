package logging

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-chi/chi/middleware"
)

var LOG *LogEntry                                  // Global log entry variable
var logLevel int                                   // Log level for the application
var grey = color.New(color.FgHiBlack).SprintFunc() // Function to colorize text in grey
var logger = log.New(os.Stdout, "", 0)             // Custom logger with no prefix and no automatic timestamp

var cwd, _ = os.Getwd() // Get the current working directory

// Define log levels as constants
// - TRACE: Trace level logging
// - DEBUG: Debug level logging
// - INFO: Info level logging
// - WARN: Warning level logging
// - ERROR: Error level logging
const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
)

// Lengths to format the log output
const (
	TimeLength                = 19
	LevelAndMethodLength      = 6
	FunctionNameAndPathLength = 40
	ByteAndElapsedLength      = 10
)

// NewLogEntry creates a new log entry for the given HTTP request
// - r: The HTTP request to create a log entry for
// Returns a new CustomLogEntry instance (middleware.LogEntry)
func (clf *LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &LogEntry{Request: r}
}

func (cle *LogEntry) Trace(message string) {
	cle.TraceWithLog(Log{Message: message})
}
func (cle *LogEntry) TraceWithLog(log Log) {
	cle.Log(log, "TRACE")
}

func (cle *LogEntry) Debug(message string) {
	cle.DebugWithLog(Log{Message: message})
}
func (cle *LogEntry) DebugWithLog(log Log) {
	cle.Log(log, "DEBUG")
}

func (cle *LogEntry) Info(message string) {
	cle.InfoWithLog(Log{Message: message})
}
func (cle *LogEntry) InfoWithLog(log Log) {
	cle.Log(log, "INFO")
}

func (cle *LogEntry) Warn(message string) {
	cle.WarnWithLog(Log{Message: message})
}
func (cle *LogEntry) WarnWithLog(log Log) {
	cle.Log(log, "WARN")
}

func (cle *LogEntry) Error(message string) {
	cle.ErrorWithLog(StandardError{Message: message})
}
func (cle *LogEntry) ErrorWithLog(err StandardError) {
	// Create a log entry for the error
	cle.Log(Log{Message: err.Message}, "ERROR")

}
func (cle *LogEntry) Panic(v any, stack []byte) {
	// Handle panic logging here as ERROR message
	cle.Log(Log{Message: "Panic"}, "ERROR", stack)
}

func (cle *LogEntry) Log(log Log, level string, params ...any) {

	// Convert level to uppercase
	level = strings.ToUpper(level)

	// Convert the level string to an integer
	var levelInt int
	switch level {
	case "TRACE":
		levelInt = TRACE
	case "DEBUG":
		levelInt = DEBUG
	case "INFO":
		levelInt = INFO
	case "WARN":
		levelInt = WARN
	case "ERROR":
		levelInt = ERROR
	default:
		levelInt = INFO
	}

	// Only log the message if the level is greater than or equal to the current log level
	if levelInt >= logLevel {

		functionName := getFunctionLocation()

		elapsed := ""
		if len(log.Elapsed) > 0 {
			elapsed = "(" + log.Elapsed + ")"
		}

		var consoleString string
		var fileString string
		if functionName != "" {
			consoleString = fmt.Sprintf("%s %s %s %s %s",
				grey(fixStringLength(time.Now().Format("2006/01/02 15:04:05"), TimeLength)),
				colorLevel(fixStringLength(level, LevelAndMethodLength)),
				color.CyanString(fixStringLength(functionName, FunctionNameAndPathLength)),
				log.Message,
				color.CyanString(fixStringLength(elapsed, ByteAndElapsedLength)),
			)
			fileString = fmt.Sprintf("%s %s %s %s %s",
				fixStringLength(time.Now().Format("2006/01/02 15:04:05"), TimeLength),
				fixStringLength(level, LevelAndMethodLength),
				fixStringLength(functionName, FunctionNameAndPathLength),
				log.Message,
				fixStringLength(elapsed, ByteAndElapsedLength),
			)
		} else {
			consoleString = fmt.Sprintf("%s %s %s %s",
				grey(fixStringLength(time.Now().Format("2006/01/02 15:04:05"), TimeLength)),
				colorLevel(fixStringLength(level, LevelAndMethodLength)),
				log.Message,
				color.CyanString(fixStringLength(elapsed, ByteAndElapsedLength)),
			)
			fileString = fmt.Sprintf("%s %s %s %s",
				fixStringLength(time.Now().Format("2006/01/02 15:04:05"), TimeLength),
				fixStringLength(level, LevelAndMethodLength),
				log.Message,
				fixStringLength(elapsed, ByteAndElapsedLength),
			)
		}

		logger.Printf("%s\n", consoleString)

		// Check to see if a log exists for today's date
		// If not, create a new log file for today's date
		// Otherwise, append to the existing log file
		logFile := GetTodayLogFile()
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()

		// Write the log message to the log file
		if _, err := fmt.Fprintf(f, "%s\n", fileString); err != nil {
			return
		}
	}
}

func (cle *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {

	// Extract the Path of request
	// If the Path is *.js, *.css, or manifest.json, skip logging
	// Otherwise, log the request
	if shouldSkipLogging(cle.Request) {
		return
	}

	var msg HTTPRequestLog
	var colorMsg HTTPRequestLog

	// Set the Date and Time of the log
	colorMsg.Time, msg.Time = getLogTime()

	// Set the Method of the request
	colorMsg.Method, msg.Method = getLogMethod(cle.Request.Method)

	// Set the Status of the request
	colorMsg.Status, msg.Status = getLogStatus(status)

	// Set the Path of the request
	colorMsg.Path, msg.Path = getLogPath(cle.Request.URL.Path)

	// Set the Bytes written
	colorMsg.Bytes, msg.Bytes = getLogBytes(bytes)

	// Set the Elapsed time
	colorMsg.Elapsed, msg.Elapsed = getLogTimeElapsed(elapsed)

	// Get the IP address of the client
	colorMsg.IP, msg.IP = getLogIP(cle.Request)

	// Log with colors and custom information
	//logger.Printf("%s %-10s\t%-45s %-10s\n",

	logger.Printf("%s %s %s %s %s %s %s",
		colorMsg.Time,
		colorMsg.Method,
		colorMsg.Path,
		colorMsg.Status,
		colorMsg.Bytes,
		colorMsg.Elapsed,
		colorMsg.IP)

	// Check to see if a log exists for today's date
	// If not, create a new log file for today's date
	// Otherwise, append to the existing log file
	logFile := GetTodayLogFile()
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Check to see if the logs directory exists
		// If not, create the logs directory
		if _, err := os.Stat("./logs"); os.IsNotExist(err) {
			os.Mkdir("./logs", 0755)
		}
		// Create a new log file for today's date
		f, err = os.Create(logFile)
		if err != nil {
			return
		}
	}
	defer f.Close()

	// Write the log message to the log file
	if _, err := fmt.Fprintf(f, "%s %s %s %s %s %s %s\n",
		msg.Time,
		msg.Method,
		msg.Path,
		msg.Status,
		msg.Bytes,
		msg.Elapsed,
		msg.IP); err != nil {
		return
	}

}

func SetLogLevel(level string) {
	switch strings.ToUpper(level) {
	case "TRACE":
		logLevel = TRACE
	case "DEBUG":
		logLevel = DEBUG
	case "INFO":
		logLevel = INFO
	case "WARN":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	default:
		logLevel = INFO
	}
}
