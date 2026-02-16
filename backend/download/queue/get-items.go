package downloadqueue

import (
	"aura/logging"
	"aura/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

func GetQueueItems(ctx context.Context) (inProgressItems []models.DBSavedItem, warningItems []models.DBSavedItem, errorItems []models.DBSavedItem, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Get Download Queue Items", logging.LevelInfo)
	defer logAction.Complete()

	Err = logging.LogErrorInfo{}
	inProgressItems = []models.DBSavedItem{}
	warningItems = []models.DBSavedItem{}
	errorItems = []models.DBSavedItem{}

	// Read all files in the download queue folder
	files, readErr := os.ReadDir(FolderPath)
	if readErr != nil {
		logAction.SetError("Failed to read download queue folder",
			"Ensure that the application has permission to read the download queue folder",
			map[string]any{
				"error": readErr.Error(),
				"path":  FolderPath,
			})
		logAction.Complete()
		return inProgressItems, warningItems, errorItems, Err
	}

	if len(files) == 0 {
		logAction.AppendResult("message", "No items found in the download queue")
		logAction.Complete()
		return inProgressItems, warningItems, errorItems, Err
	}

	// Loop through each file and categorize it based on its prefix
	for _, file := range files {
		if file.IsDir() || !(strings.HasSuffix(file.Name(), ".json")) {
			continue
		}

		// Create a sub-action for processing this file
		_, subAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Processing file: %s", file.Name()), logging.LevelDebug)

		filePath := path.Join(FolderPath, file.Name())
		subAction.AppendResult("file_path", filePath)

		// Read the file content
		content, readFileErr := os.ReadFile(filePath)
		if readFileErr != nil {
			subAction.SetError("Failed to read file",
				"Ensure that the application has permission to read files from the download queue folder",
				map[string]any{
					"error": readFileErr.Error(),
					"file":  filePath,
				})
			subAction.Complete()
			continue
		}

		var item models.DBSavedItem
		decodeErr := json.Unmarshal(content, &item)
		if decodeErr != nil {
			subAction.SetError("Failed to decode JSON content",
				"Ensure that the files in the download queue folder contain valid JSON with the correct structure",
				map[string]any{
					"error": decodeErr.Error(),
					"file":  filePath,
				})
			subAction.Complete()
			continue
		}

		if strings.HasPrefix(file.Name(), "error_") {
			errorItems = append(errorItems, item)
		} else if strings.HasPrefix(file.Name(), "warning_") {
			warningItems = append(warningItems, item)
		} else {
			inProgressItems = append(inProgressItems, item)
		}

		subAction.Complete()
	}

	logAction.AppendResult("in_progress_count", len(inProgressItems))
	logAction.AppendResult("warning_count", len(warningItems))
	logAction.AppendResult("error_count", len(errorItems))
	logAction.Complete()
	return inProgressItems, warningItems, errorItems, Err
}
