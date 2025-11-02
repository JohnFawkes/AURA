package routes_download_queue

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

func GetDownloadQueueResults(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Download Queue Results", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var inProgressEntries []api.DBMediaItemWithPosterSets
	var errorEntries []api.DBMediaItemWithPosterSets
	var warningEntries []api.DBMediaItemWithPosterSets

	// Get the download-queue folder path
	queueFolderPath := api.GetDownloadQueueFolderPath(ctx)
	if queueFolderPath == "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Read all JSON files in the download-queue folder
	files, err := os.ReadDir(queueFolderPath)
	if err != nil {
		logAction.SetError("Failed to read download queue folder",
			"Ensure that the download-queue folder exists and is accessible",
			map[string]any{"error": err.Error()})
		return
	}

	if len(files) == 0 {
		logAction.AppendResult("queue_empty", "No files in download queue")
		api.Util_Response_SendJSON(w, ld, map[string]any{
			"in_progress_entries": inProgressEntries,
			"error_entries":       errorEntries,
			"warning_entries":     warningEntries,
		})
		return
	}

	// Process each file
	for _, file := range files {
		if file.IsDir() || path.Ext(file.Name()) != ".json" {
			continue // Skip non-JSON files
		}

		// Create a sub-action for processing this file
		_, subAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Processing file: %s", file.Name()), logging.LevelDebug)

		// Get the full file path
		filePath := path.Join(queueFolderPath, file.Name())
		subAction.AppendResult("file_name", file.Name())
		subAction.AppendResult("file_path", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			subAction.AppendWarning("file_read_error", "Failed to read file")
			continue
		}

		var savedItem api.DBMediaItemWithPosterSets
		err = json.Unmarshal(data, &savedItem)
		if err != nil {
			subAction.AppendWarning("json_unmarshal_error", "Failed to unmarshal JSON data")
			continue
		}

		// If file starts with "error_" or "warning_", categorize accordingly
		if strings.HasPrefix(file.Name(), "error_") {
			errorEntries = append(errorEntries, savedItem)
		} else if strings.HasPrefix(file.Name(), "warning_") {
			warningEntries = append(warningEntries, savedItem)
		} else {
			// If there is a file, then that means it hasn't been processed yet
			inProgressEntries = append(inProgressEntries, savedItem)
		}

		subAction.Complete()
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"in_progress_entries": inProgressEntries,
		"error_entries":       errorEntries,
		"warning_entries":     warningEntries,
	})
}
