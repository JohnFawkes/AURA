package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

func AddToDownloadQueue(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Add to Download Queue", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get the DBMediaItemWithPosterSets
	var saveItem api.DBMediaItemWithPosterSets
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &saveItem, "DBMediaItemWithPosterSets")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate the JSON structure
	validateAction := logAction.AddSubAction("Validate Save Item", logging.LevelDebug)
	// Make sure it contains a TMDB_ID and LibraryTitle
	if saveItem.TMDB_ID == "" || saveItem.LibraryTitle == "" {
		validateAction.SetError("Missing Required Fields", "TMDB_ID or LibraryTitle is empty",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate that there is a MediaItem
	if saveItem.MediaItem.TMDB_ID == "" {
		validateAction.SetError("Missing Media Item Field", "MediaItem.TMDB_ID is empty",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate that there is at least one PosterSet
	if len(saveItem.PosterSets) == 0 {
		validateAction.SetError("Missing Poster Set", "At least one PosterSet is required",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate each PosterSetDetail
	for _, ps := range saveItem.PosterSets {
		if ps.PosterSetID == "" {
			validateAction.SetError("Missing PosterSetDetail Field", "PosterSetDetail.PosterSetID is empty",
				map[string]any{
					"body": saveItem,
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	}
	validateAction.Complete()

	// Add to download queue
	// We do this by saving the item to a json file in the /config/download-queue/ directory
	Err = CreateJSONFileInDownloadQueueDir(ctx, saveItem)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, saveItem)
}

func CreateJSONFileInDownloadQueueDir(ctx context.Context, saveItem api.DBMediaItemWithPosterSets) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Create Download Queue JSON File", logging.LevelInfo)
	defer logAction.Complete()

	queueFolderPath := api.GetDownloadQueueFolderPath(ctx)
	if queueFolderPath == "" {
		logAction.SetError("Failed to get download queue folder path",
			"Ensure that the download-queue folder exists and is accessible",
			nil)
		return logging.LogErrorInfo{}
	}

	// Check if the file already exists
	fileName := path.Join(queueFolderPath, fmt.Sprintf("%s_%s.json", strings.ReplaceAll(saveItem.LibraryTitle, " ", "_"), saveItem.TMDB_ID))
	exists := api.Util_File_CheckIfFileExists(fileName)
	if exists {
		logAction.AppendWarning("message", "Download queue file already exists, skipping creation")
		logAction.AppendResult("fileName", fileName)
		return logging.LogErrorInfo{}
	}

	// Marshal the saveItem to JSON
	jsonData, err := json.Marshal(saveItem)
	if err != nil {
		logAction.SetError("Failed to marshal saveItem to JSON",
			"Ensure that the saveItem can be converted to JSON",
			map[string]any{
				"error":    err.Error(),
				"saveItem": saveItem,
			})
		return logging.LogErrorInfo{}
	}

	// Write the JSON data to the file
	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		logAction.SetError("Failed to write download queue JSON file",
			"Ensure that the file can be created and written to",
			map[string]any{
				"error":    err.Error(),
				"fileName": fileName,
			})
		return logging.LogErrorInfo{}
	}

	logAction.AppendResult("fileName", fileName)
	return logging.LogErrorInfo{}

}
