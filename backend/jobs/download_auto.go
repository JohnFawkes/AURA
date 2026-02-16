package jobs

import (
	"aura/config"
	autodownload "aura/download/auto"
	"aura/logging"
	"context"
)

func StartAutoDownloadJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if autodownloadJobID != 0 {
		c.Remove(autodownloadJobID)
		autodownloadJobID = 0
	}

	enabled := config.Current.AutoDownload.Enabled
	if !enabled {
		logging.LOGGER.Info().Timestamp().Msg("AutoDownload Job Stopped")
		return nil
	}

	spec := config.Current.AutoDownload.Cron
	if spec == "" {
		spec = "0 0 * * *" // Default to daily at midnight
	}

	var err error

	autodownloadJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("Panic in scheduled AutoDownload Job")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("AutoDownload Check", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := autodownload.CheckAllItems(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(autodownloadJobID).Next.String()).
				Msg("Error running AutoDownload Job")
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(autodownloadJobID).Next.String()).
				Msg("AutoDownload Job Completed")
		}
	})
	if err != nil {
		return err
	}
	jobSpecs[autodownloadJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Msg("AutoDownload Job Started")
	return nil
}

func RunAutoDownloadJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if config.Current.AutoDownload.Enabled == false {
		logging.LOGGER.Warn().Timestamp().Msg("AutoDownload is disabled, cannot run job")
		return
	}

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if autodownloadJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("AutoDownload Job is not scheduled")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("Panic in AutoDownload Job")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Manual Job Run")
		action := ld.AddAction("AutoDownload Check", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := autodownload.CheckAllItems(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(autodownloadJobID).Next.String()).
				Msg("Error running Manual AutoDownload Job")
			return
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(autodownloadJobID).Next.String()).
				Msg("Manual AutoDownload Job Completed")
		}
	}()
}
