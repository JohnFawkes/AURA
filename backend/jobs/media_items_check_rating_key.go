package jobs

import (
	"aura/logging"
	"aura/mediaserver"
	"context"
)

func StartCheckForMediaItemChangesJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if checkForMediaItemChangesJobID != 0 {
		c.Remove(checkForMediaItemChangesJobID)
		checkForMediaItemChangesJobID = 0
	}

	var err error
	spec := "0 */6 * * *"
	checkForMediaItemChangesJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in scheduled CheckForMediaItemChangesJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Check for Media Item Changes", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := mediaserver.CheckForMediaItemChanges(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(checkForMediaItemChangesJobID).Next.String()).
				Msg("Error running Check for Media Item Changes Job")
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(checkForMediaItemChangesJobID).Next.String()).
				Msg("Check for Media Item Changes Job Completed")
		}
		ld.Log()
	})
	if err != nil {
		return err
	}
	jobSpecs[checkForMediaItemChangesJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Str("interval", "every 6 hours").
		Msg("Check for Media Item Changes Job Started")
	return nil
}

func RunCheckForMediaItemChangesJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if checkForMediaItemChangesJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("Check for Media Item Changes Job is not scheduled")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in CheckForMediaItemChangesJob")
			}
		}()
		logging.LOGGER.Info().Timestamp().Msg("Manually triggering Check for Media Item Changes Job")
		entry := c.Entry(checkForMediaItemChangesJobID)
		entry.Job.Run()
	}()
}
