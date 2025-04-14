package logging

import (
	"fmt"
	"os"
	"path"
)

var LogFolder string

func init() {
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
