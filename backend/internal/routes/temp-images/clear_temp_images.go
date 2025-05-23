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

	// Check if the temporary folder exists
	logErr := utils.CheckFolderExists(TempImageFolder)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	err := os.RemoveAll(TempImageFolder)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: err,
			Log: logging.Log{
				Message: "Failed to clear temporary images folder",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Return a JSON response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Temporary images folder cleared successfully",
		Elapsed: utils.ElapsedTime(startTime),
	})
}
