package tempimages

import (
	"aura/internal/logging"
	"aura/internal/utils"
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

func ClearTempImages(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Check if the temporary folder exists
	Err = utils.CheckFolderExists(TempImageFolder)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	err := os.RemoveAll(TempImageFolder)
	if err != nil {

		Err.Message = "Failed to clear temporary images folder"
		Err.HelpText = "Ensure the temporary images folder is accessible and not in use by any process."
		Err.Details = err.Error()
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Return a JSON response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "Temporary images folder cleared successfully",
	})
}
