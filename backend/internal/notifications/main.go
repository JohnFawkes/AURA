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
	Err := logging.NewStandardError()

	if !validNotificationProvider() {
		logging.LOG.Warn(fmt.Sprintf("Invalid notification provider: %s", config.Global.Notification.Provider))

		Err.Message = fmt.Sprintf("Invalid notification provider: %s", config.Global.Notification.Provider)
		Err.HelpText = "Ensure the notification provider is set to a valid value in the configuration."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	message := "This is a test notification from MediUX AURA"
	title := "Test Notification"
	imageURL := ""
	Err = SendDiscordNotification(message, imageURL, title)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}
	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success",
	})
}
