package jobs

import (
	"aura/logging"
	"aura/mediaserver"
	"context"
)

func StartRefreshMediaItemsAndCollectionsJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if refreshMediaItemsAndCollectionsJobID != 0 {
		c.Remove(refreshMediaItemsAndCollectionsJobID)
		refreshMediaItemsAndCollectionsJobID = 0
	}

	var err error
	spec := "*/90 * * * *"
	refreshMediaItemsAndCollectionsJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in scheduled RefreshMediaItemsAndCollectionsJob")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Refresh Media Items and Collections", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		mediaserver.GetAllLibrarySectionsAndItems(ctx, true)
	})
	if err != nil {
		return err
	}
	jobSpecs[refreshMediaItemsAndCollectionsJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Str("interval", "every 90 minutes").
		Msg("Refresh Media Items and Collections Job Started")
	return nil
}
