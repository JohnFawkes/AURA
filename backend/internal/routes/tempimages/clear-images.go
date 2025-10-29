package routes_tempimages

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"os"
	"path"
)

var TempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	TempImageFolder = path.Join(configPath, "temp-images")
}

func ClearTempImages(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Clear Temp Images", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Check if the temporary folder exists
	Err := api.Util_File_CheckFolderExists(ctx, TempImageFolder)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Remove all files in the temporary folder
	clearCount, Err := api.Util_File_ClearFilesFromFolder(ctx, TempImageFolder, 0)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if clearCount == 0 {
		api.Util_Response_SendJSON(w, ld, "No temporary images to clear")
		return
	} else {
		logAction.AppendResult("cleared_temp_image_count", clearCount)
	}

	api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Cleared %d temporary images", clearCount))
}
