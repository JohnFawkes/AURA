package tempimages

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
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

	// Iterate over the contents of the folder and remove each item
	err := filepath.Walk(TempImageFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root folder itself
		if path == TempImageFolder {
			return nil
		}

		// Remove the file or directory
		if info.IsDir() {
			return os.RemoveAll(path) // Remove subdirectories
		}
		return os.Remove(path) // Remove files
	})
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
	utils.SendJsonResponse(w, http.StatusInternalServerError, utils.JSONResponse{
		Status:  "success",
		Message: "Temporary images folder cleared successfully",
		Elapsed: utils.ElapsedTime(startTime),
	})
}
