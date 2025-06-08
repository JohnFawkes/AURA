package notifications

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"slices"
	"time"
)

// Valid NotificationProviders is a list of valid notification providers
var ValidNotificationProviders = []string{
	"Discord",
}

func validNotificationProvider() bool {
	return slices.Contains(ValidNotificationProviders, config.Global.Notification.Provider)
}

func SendTestNotification(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	if !validNotificationProvider() {
		logging.LOG.Warn(fmt.Sprintf("Invalid notification provider: %s", config.Global.Notification.Provider))
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("invalid notification provider"),
			Log: logging.Log{
				Message: fmt.Sprintf("Invalid notification provider: %s", config.Global.Notification.Provider),
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	message := "This is a test notification from MediUX AURA"
	title := "Test Notification"
	imageURL := ""
	logErr := SendDiscordNotification(message, imageURL, title)
	if logErr.Err != nil {
		logging.LOG.Warn(logErr.Log.Message)
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}
	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Test notification sent successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success",
	})
}
