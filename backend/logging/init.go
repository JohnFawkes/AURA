package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync/atomic"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var LOGGER *zerolog.Logger                      // Global logger instance
var LogFolder string                            // Global log folder path
var LogLevel zerolog.Level = zerolog.DebugLevel // -1: Trace, 0: Debug, 1: Info, 2: Warn, 3: Error
var LogFilePath string                          // Global log file path

var (
	devMode   atomic.Bool
	nopLogger = zerolog.New(io.Discard)
)

// SetDevMode should be called from main during startup.
func SetDevMode(enabled bool) {
	if enabled {
		fmt.Println("Dev mode enabled: TRACE logging is active")
	}
	devMode.Store(enabled)
}

// Dev returns a TRACE event only when dev mode is enabled.
// When disabled, it returns a no-op event (safe to call .Msg()).
func Dev() *zerolog.Event {
	if !devMode.Load() || LOGGER == nil {
		return nopLogger.Trace()
	}
	return LOGGER.Trace().Timestamp().Str("dev", "aura")
}

// DevMsg is a convenience wrapper.
func DevMsg(msg string) {
	Dev().Msg(msg)
}

func DevMsgf(format string, v ...any) {
	Dev().Msgf(format, v...)
}

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

	// rotate logs using lumberjack; output is written as JSON lines (jsonl)
	logFile := &lumberjack.Logger{
		Filename:   LogFilePath,
		MaxSize:    25,    // megabytes
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

	// Use the global LogLevel
	zerolog.SetGlobalLevel(LogLevel)
	logger := zerolog.New(multi).Level(LogLevel)
	LOGGER = &logger
}

func ApplyLogLevel() {
	zerolog.SetGlobalLevel(LogLevel)
	if LOGGER != nil {
		l := LOGGER.Level(LogLevel)
		LOGGER = &l
	}
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
	ApplyLogLevel()
}
