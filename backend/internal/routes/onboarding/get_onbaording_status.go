package routes_onboarding

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"net/http"
	"time"
)

type onboardingStatus struct {
	ConfigLoaded bool          `json:"configLoaded"`
	ConfigValid  bool          `json:"configValid"`
	NeedsSetup   bool          `json:"needsSetup"`
	CurrentSetup modals.Config `json:"currentSetup"`
}

func GetOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logging.LOG.Trace(r.URL.Path)

	config.ValidateConfig(config.Global)

	status := onboardingStatus{
		ConfigLoaded: config.ConfigLoaded,
		ConfigValid:  config.ConfigValid,
		NeedsSetup:   !(config.ConfigLoaded && config.ConfigValid),
		CurrentSetup: utils.SanitizedCopy(),
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(start),
		Data:    status,
	})
}
