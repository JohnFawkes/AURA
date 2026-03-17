package downloadqueue

import (
	"aura/config"
	"aura/logging"
	"aura/utils"
	"context"
	"os"
	"path"
	"time"
)

type Status string

const (
	LAST_STATUS_SUCCESS    Status = "Success"
	LAST_STATUS_WARNING    Status = "Warning"
	LAST_STATUS_ERROR      Status = "Error"
	LAST_STATUS_IDLE       Status = "Idle - Queue Empty"
	LAST_STATUS_PROCESSING Status = "Processing"
)

var (
	LatestInfo = struct {
		Time     time.Time
		Status   Status
		Message  string
		Errors   []string
		Warnings []string
	}{}

	FolderPath string = ""
)

type FileIssues struct {
	Errors   []string
	Warnings []string
}

func init() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Download Queue Init")
	defer ld.Log()
	logAction := ld.AddAction("Initializing Download Queue", logging.LevelTrace)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer logAction.Complete()

	FolderPath = path.Join(config.ConfigPath, "download-queue")

	// Create the download queue folder if it doesn't exist
	Err := utils.CreateFolderIfNotExists(ctx, FolderPath)
	if Err.Message != "" {
		os.Exit(1)
	}
}
