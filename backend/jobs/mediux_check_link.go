package jobs

import (
	"aura/logging"
	"aura/mediux"
	"context"
)

func StartCheckMediuxSiteLinkJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if checkMediuxSiteLinkJobID != 0 {
		c.Remove(checkMediuxSiteLinkJobID)
		checkMediuxSiteLinkJobID = 0
	}

	var err error
	spec := "*/60 * * * *"
	checkMediuxSiteLinkJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in scheduled CheckMediuxSiteLinkJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Check Mediux Site Link Availability", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		mediux.CheckSiteLinkAvailability()
	})
	if err != nil {
		return err
	}
	jobSpecs[checkMediuxSiteLinkJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Str("interval", "every 60 minutes").
		Msg("Check Mediux Site Link Availability Job Started")
	return nil
}

func RunCheckMediuxSiteLinkJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if checkMediuxSiteLinkJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("Check Mediux Site Link Availability Job is not scheduled")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in CheckMediuxSiteLinkJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Manual Job Run")
		action := ld.AddAction("Check Mediux Site Link Availability", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		mediux.CheckSiteLinkAvailability()
	}()
}
