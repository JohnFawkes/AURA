package health

import (
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
	"time"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	safeConfigData := *config.Global

	// Remove by masking sensitive information with "***"
	// Keep the last 4 characters of the data
	safeConfigData.Mediux.Token = "***" + safeConfigData.Mediux.Token[len(safeConfigData.Mediux.Token)-4:]
	safeConfigData.TMDB.ApiKey = "***" + safeConfigData.TMDB.ApiKey[len(safeConfigData.TMDB.ApiKey)-4:]
	safeConfigData.MediaServer.Token = "***" + safeConfigData.MediaServer.Token[len(safeConfigData.MediaServer.Token)-4:]

	safeConfigData.Logging.File = logging.GetTodayLogFile()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Config data retrieved successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}
