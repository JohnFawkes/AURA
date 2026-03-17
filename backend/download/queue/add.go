package downloadqueue

import (
	"aura/logging"
	"aura/models"
	"aura/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

func AddToQueue(ctx context.Context, saveItem models.DBSavedItem) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Add Entry for %s",
			utils.MediaItemInfo(saveItem.MediaItem)),
		logging.LevelDebug)

	Err = logging.LogErrorInfo{}

	// Build a regex pattern to find existing files for the same item in the download queue folder
	pattern := fmt.Sprintf(`^%s_%s_\d+\.json$`,
		strings.ReplaceAll(saveItem.MediaItem.LibraryTitle, " ", `_`),
		saveItem.MediaItem.TMDB_ID,
	)
	re := regexp.MustCompile(pattern)

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
		return Err
	}

	// Loop through each file and check if it matches the item being added
	// If it does, append a warning
	// If no matches, create a new file name
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if re.MatchString(file.Name()) {
			// If a matching file is found, append a warning to the log and return an error
			logAction.AppendWarning("message", "Item already exists in download queue")
			logAction.AppendWarning("file", file.Name())
			logAction.Complete()
			return Err
		}
	}

	// If no matching file is found, create a new file name with the format: LibraryTitle_TMDBID_timestamp.json
	timestamp := time.Now().Unix()
	fileName := path.Join(FolderPath, fmt.Sprintf("%s_%s_%d.json",
		strings.ReplaceAll(saveItem.MediaItem.LibraryTitle, " ", `_`),
		saveItem.MediaItem.TMDB_ID,
		timestamp,
	))

	// Marshal the saveItem to JSON
	jsonData, marshallErr := json.Marshal(saveItem)
	if marshallErr != nil {
		logAction.SetError("Failed to marshal Save Item to JSON",
			"Ensure that the Save Item can be converted to JSON",
			map[string]any{
				"error": marshallErr.Error(),
				"item":  saveItem,
			})
		logAction.Complete()
		return Err
	}

	// Write the JSON data to a file in the download queue folder
	writeErr := os.WriteFile(fileName, jsonData, 0644)
	if writeErr != nil {
		logAction.SetError("Failed to write Save Item to download queue file",
			"Ensure that the application has permission to write to the download queue folder",
			map[string]any{
				"error": writeErr.Error(),
				"file":  fileName,
			})
		logAction.Complete()
		return Err
	}

	logAction.AppendResult("file", fileName)
	logAction.Complete()
	return Err
}
