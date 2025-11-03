package routes_download_queue

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func DeleteFromDownloadQueue(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Delete From Download Queue", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get the DBMediaItemWithPosterSets
	var item api.DBMediaItemWithPosterSets
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &item, "DBMediaItemWithPosterSets")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	queueDir := api.GetDownloadQueueFolderPath(ctx)
	if queueDir == "" {
		logAction.SetError("Download queue folder path not found", "", nil)
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Read all files in the download queue directory
	files, err := os.ReadDir(queueDir)
	if err != nil {
		logAction.SetError("Failed to read download queue folder",
			"Ensure that the download-queue folder exists and is accessible",
			map[string]any{"error": err.Error()})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Build a regex pattern to match files related to the given item
	pattern := fmt.Sprintf(`^(error_|warning_)?%s_%s_\d+\.json$`,
		strings.ReplaceAll(item.LibraryTitle, " ", `_`),
		item.TMDB_ID,
	)
	re := regexp.MustCompile(pattern)

	deleted := 0
	// Iterate over files to find matches
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		logging.LOGGER.Debug().Timestamp().Str("file_name", file.Name()).Str("pattern", pattern).Msg("Checking file for deletion")

		if re.MatchString(file.Name()) {
			err := os.Remove(fmt.Sprintf("%s/%s", queueDir, file.Name()))
			if err != nil {
				logAction.AppendWarning("file_delete_error",
					fmt.Sprintf("Failed to delete file: %s", file.Name()))
			} else {
				deleted++
				logAction.AppendResult("file_deleted", fmt.Sprintf("Deleted file: %s", file.Name()))
			}
		}
	}

	if deleted == 0 {
		logAction.AppendWarning("no_files_deleted", "No matching files found to delete")
	} else {
		logAction.AppendResult("total_files_deleted", fmt.Sprintf("Total files deleted: %d", deleted))
	}

	api.Util_Response_SendJSON(w, ld,
		fmt.Sprintf("Deleted %s (%s) from download queue", item.MediaItem.Title, item.MediaItem.LibraryTitle))
}
