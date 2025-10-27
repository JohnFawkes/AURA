package routes_tempimages

import (
	"aura/internal/api"
	"aura/internal/logging"

	"fmt"
	"net/http"
	"os"
	"path"
	"time"
)

var TempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	TempImageFolder = path.Join(configPath, "temp-images")

}

func Clear(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Check if the temporary folder exists
	Err = api.Util_File_CheckFolderExists(TempImageFolder)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	clearCount, Err := api.Util_File_ClearFilesFromFolder(TempImageFolder, 0)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if clearCount == 0 {
		api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
			Status:  "success",
			Elapsed: api.Util_ElapsedTime(startTime),
			Data:    "No temporary images to clear",
		})
		return
	}

	logging.LOG.Info(fmt.Sprintf("Cleared %d temporary images from %s", clearCount, TempImageFolder))

	// Return a JSON response
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    fmt.Sprintf("Cleared %d temporary images", clearCount),
	})
}
