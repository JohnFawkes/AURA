package routes_sonarr_radarr

import (
	"aura/internal/api"
	"aura/internal/logging"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"time"
)

type SonarrWebHookOnUpgradePayload struct {
	DeletedFiles []SonarrFile    `json:"deletedFiles"`
	EpisodeFile  SonarrFile      `json:"episodeFile"`
	Episodes     []SonarrEpisode `json:"episodes"`
	EventType    string          `json:"eventType"`
	InstanceName string          `json:"instanceName"`
	IsUpgrade    bool            `json:"isUpgrade"`
	Series       SonarrSeries    `json:"series"`
}

type SonarrFile struct {
	Path         string `json:"path"`
	RelativePath string `json:"relativePath"`
}

type SonarrEpisode struct {
	EpisodeNumber int `json:"episodeNumber"`
	SeasonNumber  int `json:"seasonNumber"`
}

type SonarrSeries struct {
	Path   string `json:"path"`
	Title  string `json:"title"`
	TmdbID int    `json:"tmdbId"`
}

func SonarrWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Handle Sonarr Webhook", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Library from the URL params
	library := r.URL.Query().Get("library")
	if library == "" {
		logAction.SetError("Missing library parameter", "The 'library' URL parameter is required", nil)
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Get the webhook payload
	var payload SonarrWebHookOnUpgradePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// If not an upgrade event, ignore
	if !payload.IsUpgrade {
		logging.LOGGER.Debug().Timestamp().Str("Library", library).Str("Title", payload.Series.Title).Msg("Ignoring non-upgrade Sonarr webhook event")
		w.WriteHeader(http.StatusOK)
		return
	} else if payload.EpisodeFile == (SonarrFile{}) || payload.EpisodeFile.Path == "" {
		logging.LOGGER.Debug().Timestamp().Str("Library", library).Str("Title", payload.Series.Title).Msg("No episode file in payload, ignoring")
		w.WriteHeader(http.StatusOK)
		return
	} else if payload.Series == (SonarrSeries{}) || payload.Series.TmdbID == 0 {
		logging.LOGGER.Debug().Timestamp().Str("Library", library).Str("Title", payload.Series.Title).Msg("No series TMDB ID in payload, ignoring")
		w.WriteHeader(http.StatusOK)
		return
	} else if len(payload.Episodes) == 0 {
		logging.LOGGER.Debug().Timestamp().Str("Library", library).Str("Title", payload.Series.Title).Msg("No episodes in payload, ignoring")
		w.WriteHeader(http.StatusOK)
		return
	}

	sonarrEpisode := payload.Episodes[0]

	// Using the TMDB ID and the Library Title from the URL Param, find the corresponding Aura DB Item
	items, _, _, pageErr := api.DB_GetAllItemsWithFilter(
		ctx,
		strconv.Itoa(payload.Series.TmdbID), // searchTMDBID
		library,                             // searchLibrary
		0,                                   // searchYear
		"",                                  // searchTitle
		[]string{},                          // librarySections
		[]string{},                          // filteredTypes
		"all",                               // filterAutoDownload
		false,                               // multisetOnly
		[]string{},                          // filteredUsernames
		5,                                   // itemsPerPage
		1,                                   // pageNumber
		"dateDownloaded",                    // sortOption
		"desc",                              // sortOrder
	)
	if pageErr.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if len(items) == 0 {
		logAction.SetError("No DB item found", "No matching item found in the database for the given TMDB ID and library", map[string]any{
			"tmdb_id": payload.Series.TmdbID,
			"library": library,
		})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	} else if len(items) > 1 {
		logAction.SetError("Multiple DB items found", "Multiple matching items found in the database for the given TMDB ID and library", map[string]any{
			"tmdb_id": payload.Series.TmdbID,
			"library": library,
			"count":   len(items),
		})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	dbItem := items[0]

	// Respond to Sonarr immediately
	w.WriteHeader(http.StatusOK)

	// Run the rest in a goroutine
	go func(dbItem api.DBMediaItemWithPosterSets, sonarrEpisode SonarrEpisode, library string) {
		// Create a new background context and logging data for the background task
		bgCtx := context.Background()
		bgCtx, bgLd := logging.CreateLoggingContext(bgCtx, "Downloading Titlecard for Sonarr Upgrade Webhook")
		logAction := bgLd.AddAction("Sonarr Webhook Background Task", logging.LevelInfo)
		bgCtx = logging.WithCurrentAction(bgCtx, logAction)

		// Sleep for a short time to ensure the media server has processed the upgrade
		sleepAction := logAction.AddSubAction("Sleeping for 15 seconds to give time for media server to update", logging.LevelTrace)
		time.Sleep(15 * time.Second)
		sleepAction.Complete()

		defer bgLd.Log()

		// Get the latest information from the Media Server
		mediaItem, Err := api.CallFetchItemContent(bgCtx, dbItem.MediaItem.RatingKey, library)
		if Err.Message != "" {
			logging.LOGGER.Error().Msgf("Error fetching media item: %s", Err.Message)
			return
		}

		// Loop through poster sets
		for _, posterSet := range dbItem.PosterSets {
			if !slices.Contains(posterSet.SelectedTypes, "titlecard") {
				continue
			}
			var downloadFile api.PosterFile
			for _, episodeFile := range posterSet.PosterSet.TitleCards {
				if episodeFile.Episode.EpisodeNumber != sonarrEpisode.EpisodeNumber || episodeFile.Episode.SeasonNumber != sonarrEpisode.SeasonNumber {
					continue
				}
				downloadFile = episodeFile
			}
			downloadFileName := api.MediaServer_GetFileDownloadName(downloadFile)
			Err = api.CallDownloadAndUpdatePosters(bgCtx, mediaItem, downloadFile)
			if Err.Message != "" {
				logging.LOGGER.Error().Msgf("Error downloading/updating poster: %s", Err.Message)
				return
			}
			api.DeleteTempImageForNextLoad(bgCtx, downloadFile, mediaItem.RatingKey)
			logging.LOGGER.Debug().Timestamp().Str("Library", library).Str("Title", dbItem.MediaItem.Title).Msgf("Redownloaded title card: %s for S%dE%d", downloadFileName, sonarrEpisode.SeasonNumber, sonarrEpisode.EpisodeNumber)
		}
	}(dbItem, sonarrEpisode, library)
}
