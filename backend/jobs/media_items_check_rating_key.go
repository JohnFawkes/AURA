package jobs

import (
	"aura/logging"
	"aura/mediaserver"
	"context"
)

func StartCheckForRatingKeyChangesJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if checkForRatingKeyChangesJobID != 0 {
		c.Remove(checkForRatingKeyChangesJobID)
		checkForRatingKeyChangesJobID = 0
	}

	var err error
	spec := "0 */6 * * *"
	checkForRatingKeyChangesJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in scheduled CheckForRatingKeyChangesJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Check for Rating Key Changes", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := mediaserver.CheckForRatingKeyChanges(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(checkForRatingKeyChangesJobID).Next.String()).
				Msg("Error running Check for Rating Key Changes Job")
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(checkForRatingKeyChangesJobID).Next.String()).
				Msg("Check for Rating Key Changes Job Completed")
		}
	})
	if err != nil {
		return err
	}
	jobSpecs[checkForRatingKeyChangesJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Str("interval", "every 6 hours").
		Msg("Check for Rating Key Changes Job Started")
	return nil
}

func RunCheckForRatingKeyChangesJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if checkForRatingKeyChangesJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("Check for Rating Key Changes Job is not scheduled")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in CheckForRatingKeyChangesJob")
			}
		}()
		logging.LOGGER.Info().Timestamp().Msg("Manually triggering Check for Rating Key Changes Job")
		entry := c.Entry(checkForRatingKeyChangesJobID)
		entry.Job.Run()
	}()
}
