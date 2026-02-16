package sonarr_radarr

import (
	"aura/config"
	"aura/logging"
	"context"
	"fmt"
	"net/url"
	"path"
	"time"
)

func (s *SonarrApp) TestConnection(ctx context.Context, app config.Config_SonarrRadarrApp) (valid bool, Err logging.LogErrorInfo) {
	return srTestConnection(ctx, app)
}

func (r *RadarrApp) TestConnection(ctx context.Context, app config.Config_SonarrRadarrApp) (valid bool, Err logging.LogErrorInfo) {
	return srTestConnection(ctx, app)
}

func TestConnection(ctx context.Context, app config.Config_SonarrRadarrApp) (valid bool, Err logging.LogErrorInfo) {
	interfaceSR, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return false, Err
	}
	return interfaceSR.TestConnection(ctx, app)
}

func srTestConnection(ctx context.Context, app config.Config_SonarrRadarrApp) (valid bool, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Testing Connection %s | %s", app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	valid = false
	Err = logging.LogErrorInfo{}

	// Construct the URL
	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError(fmt.Sprintf("Invalid %s URL", app.Type),
			"Make sure that the URL is properly formatted",
			map[string]any{
				"url":   app.URL,
				"error": err.Error(),
			})
		return valid, *logAction.Error
	}
	u.Path = path.Join(u.Path, "api", "v3", "system", "status")
	URL := u.String()

	var lastErrorMsg string
	var lastErrorDetail map[string]any

	// Try up to 3 times with a 5 second delay between attempts
	for attempt := 1; attempt <= 3; attempt++ {
		attemptAction := logAction.AddSubAction(fmt.Sprintf("Attempt %d to connect to %s", attempt, app.Type), logging.LevelTrace)
		resp, _, reqErr := makeRequest(ctx, app, URL, "GET", nil)
		if reqErr.Message != "" {
			attemptAction.AppendWarning(fmt.Sprintf("attempt_%d", attempt), map[string]any{"error": reqErr.Message})
			lastErrorMsg = reqErr.Message
			lastErrorDetail = reqErr.Detail
			if attempt < 3 {
				time.Sleep(5 * time.Second)
			}
			continue
		}
		defer resp.Body.Close()

		// Successful connection
		attemptAction.Complete()
		valid = true
		return valid, Err
	}

	// If we reach here, all attempts failed
	logAction.SetError(fmt.Sprintf("%s Connection Test Failed", app.Type),
		"Make sure the Sonarr/Radarr server is reachable",
		map[string]any{
			"error":  lastErrorMsg,
			"detail": lastErrorDetail,
		})
	return valid, *logAction.Error
}
