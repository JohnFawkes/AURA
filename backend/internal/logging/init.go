package logging

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var LOGGER *zerolog.Logger                      // Global logger instance
var LogFolder string                            // Global log folder path
var LogLevel zerolog.Level = zerolog.DebugLevel // -1: Trace, 0: Debug, 1: Info, 2: Warn, 3: Error
var LogFilePath string                          // Global log file path

func CreateLogFolder() {
	// Create the log directory if it doesn't exist
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	logPath := path.Join(configPath, "logs")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.MkdirAll(logPath, 0755)
		if err != nil {
			fmt.Printf("Error creating log directory: %v\n", err)
		}
	}
	LogFolder = logPath

}

func init() {
	CreateLogFolder()

	LogFilePath = path.Join(LogFolder, "aura.log")

	// Set default log level
	level := zerolog.DebugLevel

	// rotate logs using lumberjack; output is written as JSON lines (jsonl)
	logFile := &lumberjack.Logger{
		Filename:   LogFilePath,
		MaxSize:    100,   // megabytes
		MaxBackups: 7,     // number of backups
		MaxAge:     14,    // days
		LocalTime:  true,  // use local time
		Compress:   false, // disabled by default
	}

	// Create a console writer for pretty printing to console
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006/01/02 15:04:05",
		FormatLevel: func(i any) string {
			if s, ok := i.(string); ok {
				switch strings.ToLower(s) {
				case "error":
					return "\033[31m[ERROR]\033[0m" // Red
				case "warn":
					return "\033[33m[WARN]\033[0m" // Yellow
				case "info":
					return "\033[32m[INFO]\033[0m" // Green
				case "debug":
					return "\033[36m[DEBUG]\033[0m" // Cyan
				case "trace":
					return "\033[35m[TRACE]\033[0m" // Magenta
				default:
					return fmt.Sprintf("[%s]", strings.ToUpper(s))
				}
			}
			return fmt.Sprintf("[%v]", i)
		},
	}

	// Combine log file writer and console writer
	multi := zerolog.MultiLevelWriter(logFile, consoleWriter)

	logger := zerolog.New(multi).Level(level)
	LOGGER = &logger
}

func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "trace":
		LogLevel = zerolog.TraceLevel
	case "debug":
		LogLevel = zerolog.DebugLevel
	case "info":
		LogLevel = zerolog.InfoLevel
	case "warn", "warning":
		LogLevel = zerolog.WarnLevel
	case "error":
		LogLevel = zerolog.ErrorLevel
	default:
		LogLevel = zerolog.InfoLevel
	}
}
