package routes_logging

import (
	"aura/internal/api"
	"aura/internal/logging"
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type structActionLabelSection struct {
	Label   string `json:"label"`
	Section string `json:"section"`
}

var possible_actions_paths = map[string]structActionLabelSection{
	// Login & Auth
	"/api/login": {
		Label:   "User Login",
		Section: "AUTH",
	},
	// Config Routes
	"/api/config": {
		Label:   "Get Config",
		Section: "CONFIG",
	},
	"/api/config/status": {
		Label:   "Get Config Status",
		Section: "CONFIG",
	},
	"/api/config/reload": {
		Label:   "Reload Config",
		Section: "CONFIG",
	},
	"/api/config/update": {
		Label:   "Update Config",
		Section: "CONFIG",
	},
	"/api/config/validate/mediux": {
		Label:   "Validate Mediux Token",
		Section: "CONFIG",
	},
	"/api/config/validate/mediaserver": {
		Label:   "Validate Media Server Info",
		Section: "CONFIG",
	},
	"/api/config/validate/sonarr": {
		Label:   "Validate Sonarr Connection",
		Section: "CONFIG",
	},
	"/api/config/validate/radarr": {
		Label:   "Validate Radarr Connection",
		Section: "CONFIG",
	},
	"/api/config/validate/notification": {
		Label:   "Test Notifications",
		Section: "CONFIG",
	},
	// Logging Routes
	"/api/log": {
		Label:   "Get Logs",
		Section: "LOGS",
	},
	"/api/log/clear": {
		Label:   "Clear Logs",
		Section: "LOGS",
	},
	// Temp Images Routes
	"/api/temp-images/clear": {
		Label:   "Clear Temp Images",
		Section: "TEMP IMAGES",
	},
	// Media Server Routes
	"/api/mediaserver/status": {
		Label:   "Get Media Server Status",
		Section: "MEDIA",
	},
	"/api/mediaserver/type": {
		Label:   "Get Media Server Type",
		Section: "MEDIA",
	},
	"/api/mediaserver/library-options": {
		Label:   "Get Media Server Library Options",
		Section: "MEDIA",
	},
	"/api/mediaserver/sections": {
		Label:   "Get Media Server Library Sections",
		Section: "MEDIA",
	},
	"/api/mediaserver/sections/items": {
		Label:   "Get Media Server Sections & Items",
		Section: "MEDIA",
	},
	"/api/mediaserver/item": {
		Label:   "Get Item Content",
		Section: "MEDIA",
	},
	"/api/mediaserver/download": {
		Label:   "Download and Update",
		Section: "MEDIA",
	},
	"/api/mediaserver/add-to-queue": {
		Label:   "Add Item to Download Queue",
		Section: "MEDIA",
	},
	// MediUX Routes
	"/api/mediux/sets": {
		Label:   "Get All Sets",
		Section: "MEDIUX",
	},
	"/api/mediux/sets-by-user": {
		Label:   "Get Sets From User",
		Section: "MEDIUX",
	},
	"/api/mediux/set-by-id": {
		Label:   "Get Set by ID",
		Section: "MEDIUX",
	},
	"/api/mediux/image": {
		Label:   "Get Images From Set",
		Section: "MEDIUX",
	},
	"/api/mediux/user-follow-hiding": {
		Label:   "Get User Following/Hiding Sets",
		Section: "MEDIUX",
	},
	"/api/mediux/check-link": {
		Label:   "Check Mediux Link",
		Section: "MEDIUX",
	},
	// Database Routes
	"/api/db/get-all": {
		Label:   "Get All Items",
		Section: "DATABASE",
	},
	"/api/db/delete": {
		Label:   "Delete Item",
		Section: "DATABASE",
	},
	"/api/db/update": {
		Label:   "Update Item",
		Section: "DATABASE",
	},
	"/api/db/add": {
		Label:   "Add Item",
		Section: "DATABASE",
	},
	"/api/db/force-recheck": {
		Label:   "Force Recheck on Item",
		Section: "DATABASE",
	},
}

func GetLogContents(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Log Contents", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	ctx, logAction = logging.AddSubActionToContext(ctx, "Read Log File Contents", logging.LevelDebug)
	file, err := os.Open(logging.LogFilePath)
	if err != nil {
		logAction.SetError("Failed to read log file", "Make sure the log file exists and is readable",
			map[string]any{
				"error": err.Error(),
				"path":  logging.LogFilePath,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	defer file.Close()

	// Query Param - Log Level Filter
	filteredLogLevelsStr := r.URL.Query().Get("filteredLogLevels")
	var filteredLogLevels []string
	if filteredLogLevelsStr != "" {
		filteredLogLevels = strings.Split(filteredLogLevelsStr, ",")
	}

	// Query Param - Status Filter
	filteredStatusesStr := r.URL.Query().Get("filteredStatuses")
	var filteredStatuses []string
	if filteredStatusesStr != "" {
		filteredStatuses = strings.Split(filteredStatusesStr, ",")
	}

	// Query Param - Route/Action Filter
	filteredActionsStr := r.URL.Query().Get("filteredActions")
	var filteredActions []string
	if filteredActionsStr != "" {
		filteredActions = strings.Split(filteredActionsStr, ",")
	}

	// Query Param - Pagination
	itemsPerPage := 20
	pageNumber := 1
	ippStr := r.URL.Query().Get("itemsPerPage")
	if ippStr != "" {
		if val, err := strconv.Atoi(ippStr); err == nil {
			itemsPerPage = val
		}
	}
	pnStr := r.URL.Query().Get("pageNumber")
	if pnStr != "" {
		if val, err := strconv.Atoi(pnStr); err == nil {
			pageNumber = val
		}
	}

	var logEntries []*logging.LogData
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			logAction.SetError("Failed to read log file line", err.Error(), nil)
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
		var entry logging.LogData
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			// If there is no Route info and No Actions, skip this log entry
			if entry.Route == nil && len(entry.Actions) == 0 {
				continue
			}
			// If there is Route info and Actions are present, skip "Route Not Found" entries
			if entry.Route != nil && entry.Route.Path != "" && len(entry.Actions) != 0 && entry.Actions[0].Name == "Route Not Found" {
				continue
			}
			if entry.Route == nil && len(entry.Actions) != 0 {
				// If there is no Route info but Actions are present, append Action Name to possible_actions_paths
				actionName := entry.Message
				if _, exists := possible_actions_paths[actionName]; !exists {
					possible_actions_paths[actionName] = structActionLabelSection{
						Label:   actionName,
						Section: "AURA BACKGROUND TASK",
					}
				}
			}
			logEntries = append(logEntries, &entry)
		} else {
			continue
		}
	}

	// Apply Log Level Filter
	if len(filteredLogLevels) > 0 {
		filteredEntries := make([]*logging.LogData, 0, len(logEntries))
		for _, entry := range logEntries {
			// If the level is error, always keep the entry
			if strings.EqualFold(entry.Level, "error") {
				filteredEntries = append(filteredEntries, entry)
				continue
			}
			filteredActions := make([]*logging.LogAction, 0, len(entry.Actions))
			for _, action := range entry.Actions {
				filtered := filterLogActionByLevels(action, filteredLogLevels)
				if filtered != nil {
					filteredActions = append(filteredActions, filtered)
				}
			}
			entry.Actions = filteredActions
			if len(entry.Actions) > 0 {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		logEntries = filteredEntries
	}

	// Apply Status Filter
	if len(filteredStatuses) > 0 {
		filteredEntries := make([]*logging.LogData, 0, len(logEntries))
		for _, entry := range logEntries {
			// If the status is error, always keep the entry
			if strings.EqualFold(entry.Status, "error") {
				filteredEntries = append(filteredEntries, entry)
				continue
			}
			filteredActions := make([]*logging.LogAction, 0, len(entry.Actions))
			for _, action := range entry.Actions {
				filtered := filterLogActionByStatuses(action, filteredStatuses)
				if filtered != nil {
					filteredActions = append(filteredActions, filtered)
				}
			}
			entry.Actions = filteredActions
			if len(entry.Actions) > 0 {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		logEntries = filteredEntries
	}

	// Apply Route/Action Filter
	if len(filteredActions) > 0 {
		filteredEntries := make([]*logging.LogData, 0, len(logEntries))
		for _, entry := range logEntries {
			entryMatches := false
			// Check Route Path
			if entry.Route != nil && entry.Route.Path != "" {
				for _, actionFilter := range filteredActions {
					if strings.EqualFold(entry.Route.Path, actionFilter) {
						entryMatches = true
						break
					}
				}
			} else {
				// Check Background Task Name
				if entry.Message != "" {
					for _, actionFilter := range filteredActions {
						if strings.EqualFold(entry.Message, actionFilter) {
							entryMatches = true
							break
						}
					}
				}
			}
			if entryMatches {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		logEntries = filteredEntries
	}

	// Sort log entries by timestamp descending
	for i, j := 0, len(logEntries)-1; i < j; i, j = i+1, j-1 {
		logEntries[i], logEntries[j] = logEntries[j], logEntries[i]
	}

	// Get the total number of log entries before pagination
	totalNumberOfLogEntries := len(logEntries)

	// Apply Pagination
	startIndex := (pageNumber - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage
	if startIndex > len(logEntries) {
		logEntries = []*logging.LogData{}
	} else if endIndex > len(logEntries) {
		logEntries = logEntries[startIndex:]
	} else {
		logEntries = logEntries[startIndex:endIndex]
	}

	logging.LOGGER.Debug().Timestamp().Msgf("Retrieved %d-%d of %d log entries after filtering and pagination",
		startIndex+1, startIndex+len(logEntries), totalNumberOfLogEntries)
	logAction.AppendResult("log_entries_total", totalNumberOfLogEntries)
	logAction.AppendResult("log_entries_returned", len(logEntries))
	logAction.AppendResult("log_entries_filtered", totalNumberOfLogEntries-len(logEntries))

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"total_log_entries":      totalNumberOfLogEntries,
		"possible_actions_paths": possible_actions_paths,
		"log_entries":            logEntries,
	})
}

// Recursively filter sub-actions by log level
func filterLogActionByLevels(action *logging.LogAction, filteredLogLevels []string) *logging.LogAction {
	// If the level is error, always keep it
	if strings.EqualFold(action.Level, "error") {
		return action
	}

	// Filter sub-actions recursively
	filteredSubActions := make([]*logging.LogAction, 0, len(action.SubActions))
	for _, sub := range action.SubActions {
		filtered := filterLogActionByLevels(sub, filteredLogLevels)
		if filtered != nil {
			filteredSubActions = append(filteredSubActions, filtered)
		}
	}
	action.SubActions = filteredSubActions

	// Check if this action matches any log level
	actionMatches := false
	for _, lvl := range filteredLogLevels {
		if strings.EqualFold(action.Level, lvl) {
			actionMatches = true
			break
		}
	}

	// Keep this action if it matches or has any sub-actions left
	if actionMatches || len(action.SubActions) > 0 {
		return action
	}
	return nil
}

// Recursively filter log actions by status
func filterLogActionByStatuses(action *logging.LogAction, filteredStatuses []string) *logging.LogAction {
	// If the status is error, always keep it
	if strings.EqualFold(action.Status, "error") {
		return action
	}

	// Filter sub-actions recursively
	filteredSubActions := make([]*logging.LogAction, 0, len(action.SubActions))
	for _, sub := range action.SubActions {
		filtered := filterLogActionByStatuses(sub, filteredStatuses)
		if filtered != nil {
			filteredSubActions = append(filteredSubActions, filtered)
		}
	}
	action.SubActions = filteredSubActions

	// Check if this action matches any status
	actionMatches := false
	for _, status := range filteredStatuses {
		if strings.EqualFold(action.Status, status) {
			actionMatches = true
			break
		}
	}

	// Keep this action if it matches or has any sub-actions left
	if actionMatches || len(action.SubActions) > 0 {
		return action
	}
	return nil
}
