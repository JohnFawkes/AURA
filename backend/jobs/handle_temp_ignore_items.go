package jobs

import (
	"aura/logging"
	"aura/mediaserver"
	"context"
)

func StartHandleTempIgnoredItemsJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if handleTempIgnoredItemsJobID != 0 {
		c.Remove(handleTempIgnoredItemsJobID)
		handleTempIgnoredItemsJobID = 0
	}

	var err error
	spec := "0 */1 * * *"
	handleTempIgnoredItemsJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("Panic in scheduled HandleTempIgnoredItemsJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Handle Temp Ignored Items", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := mediaserver.HandleTempIgnoredItems(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(handleTempIgnoredItemsJobID).Next.String()).
				Msg("Error running Handle Temp Ignored Items Job")
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(handleTempIgnoredItemsJobID).Next.String()).
				Msg("Handle Temp Ignored Items Job Completed")
		}
	})
	if err != nil {
		return err
	}
	jobSpecs[handleTempIgnoredItemsJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Str("interval", "every 1 hour").
		Msg("Handle Temp Ignored Items Job Started")
	return nil
}

func RunHandleTempIgnoredItemsJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if handleTempIgnoredItemsJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("Handle Temp Ignored Items Job is not scheduled")
		return
	}

	go func() {
		entry := c.Entry(handleTempIgnoredItemsJobID)
		if entry.ID == 0 {
			logging.LOGGER.Error().Timestamp().Msg("Handle Temp Ignored Items Job entry not found")
			return
		}
		entry.Job.Run()
	}()
}
