package tempimages

import (
	"aura/internal/logging"
	"aura/internal/utils"
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

	clearCount, Err := utils.ClearFilesFromFolder(TempImageFolder, 0)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if clearCount == 0 {
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    "No temporary images to clear",
		})
		return
	}

	logging.LOG.Info(fmt.Sprintf("Cleared %d temporary images from %s", clearCount, TempImageFolder))

	// Return a JSON response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    fmt.Sprintf("Cleared %d temporary images", clearCount),
	})
}
