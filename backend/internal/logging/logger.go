package logging

import (
	"time"
)

// Complete finalizes LogData before logging.
func (ld *LogData) Complete() {
	if ld.Status == "" {
		ld.Status = StatusSuccess
	}
	ld.ElapsedMicroseconds = time.Since(ld.Timestamp).Microseconds()

	// Complete all actions
	for _, action := range ld.Actions {
		action.Complete()
	}
}

// Log writes the structured log using zerolog.
func (ld *LogData) Log() {
	ld.Complete()

	highestLevel := getHighestActionLevel(ld.Actions)

	// Skip logging if below global log level and no route info
	if int(LogLevel) > LogLevelToInt(highestLevel) && ld.Route == nil {
		return
	}

	// Filter actions by log level, unless error status
	var filteredActions []*LogAction
	var hasWarn bool
	if ld.Status == StatusError {
		filteredActions = ld.Actions
	} else {
		for _, action := range ld.Actions {
			if action.Level == LevelWarn {
				hasWarn = true
			}
			if int(LogLevel) <= LogLevelToInt(action.Level) {
				filteredActions = append(filteredActions, action)
			}
		}
	}
	ld.Actions = filteredActions

	// Select zerolog event level
	event := LOGGER.Debug()
	if !hasWarn {
		switch highestLevel {
		case LevelError:
			event = LOGGER.Error()
			ld.Status = StatusError
		case LevelWarn:
			event = LOGGER.Warn()
		case LevelInfo:
			event = LOGGER.Info()
		case LevelDebug:
			event = LOGGER.Debug()
		case LevelTrace:
			event = LOGGER.Trace()
		}
	} else {
		if highestLevel != LevelError {
			event = LOGGER.Warn()
		} else {
			event = LOGGER.Error()
			ld.Status = StatusError
		}
	}

	// Add fields to log event
	event.Timestamp().
		Str("status", ld.Status).
		Int64("elapsed_us", ld.ElapsedMicroseconds)

	if len(ld.Actions) > 0 {
		event.Interface("actions", ld.Actions)
	}
	if ld.Route != nil {
		event.Interface("route", ld.Route)
	}

	event.Msg(ld.Message)
}

// getHighestActionLevel recursively finds the highest log level among actions and sub-actions.
func getHighestActionLevel(actions []*LogAction) string {
	highest := LevelDebug
	for _, action := range actions {
		if LogLevelToInt(action.Level) > LogLevelToInt(highest) {
			highest = action.Level
		}
		if len(action.SubActions) > 0 {
			subHighest := getHighestActionLevel(action.SubActions)
			if LogLevelToInt(subHighest) > LogLevelToInt(highest) {
				highest = subHighest
			}
		}
	}
	return highest
}
