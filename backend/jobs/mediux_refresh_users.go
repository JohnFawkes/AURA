package jobs

import (
	"aura/logging"
	"aura/mediux"
	"context"
)

func StartRefreshMediuxUsersJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if refreshMediuxUsersJobID != 0 {
		c.Remove(refreshMediuxUsersJobID)
		refreshMediuxUsersJobID = 0
	}

	var err error
	spec := "*/90 * * * *"
	refreshMediuxUsersJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in scheduled RefreshMediuxUsersJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Refresh Mediux Users", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		_, Err := mediux.GetAllUsers(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(refreshMediuxUsersJobID).Next.String()).
				Msg("Error running Refresh Mediux Users Job")
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(refreshMediuxUsersJobID).Next.String()).
				Msg("Refresh Mediux Users Job Completed")
		}
		ld.Log()
	})
	if err != nil {
		return err
	}
	jobSpecs[refreshMediuxUsersJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Str("interval", "every 90 minutes").
		Msg("Refresh Mediux Users Job Started")
	return nil
}

func RunRefreshMediuxUsersJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if refreshMediuxUsersJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("Refresh Mediux Users Job is not scheduled")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in RefreshMediuxUsersJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Manual Job Run")
		action := ld.AddAction("Refresh Mediux Users", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		mediux.GetAllUsers(ctx)
		logging.LOGGER.Info().Timestamp().
			Str("next_run", c.Entry(refreshMediuxUsersJobID).Next.String()).
			Msg("Manual Refresh Mediux Users Job Completed")
	}()
}
