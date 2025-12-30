package routes_sonarr_radarr

import (
	"aura/internal/api"
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
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

type mediaItemFetchInfo struct {
	FromCache bool
	FetchErr  string
	CacheHit  bool
}

// fetchMediaItemWithCacheFallback tries the media-server fetch first, then falls back to cache by TMDB ID.
// Returns the media item (zero-value if not found), info about the path taken, and ok=false if both fail.
func fetchMediaItemWithCacheFallback(
	ctx context.Context,
	librarySection string,
	ratingKey string,
	tmdbID string,
) (api.MediaItem, mediaItemFetchInfo, bool) {
	mediaItem, Err := api.CallFetchItemContent(ctx, ratingKey, librarySection)
	if Err.Message == "" {
		return mediaItem, mediaItemFetchInfo{FromCache: false, FetchErr: "", CacheHit: false}, true
	}

	ptr, exists := api.Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(librarySection, tmdbID)
	if !exists || ptr == nil {
		return api.MediaItem{}, mediaItemFetchInfo{FromCache: false, FetchErr: Err.Message, CacheHit: false}, false
	}

	return *ptr, mediaItemFetchInfo{FromCache: true, FetchErr: Err.Message, CacheHit: true}, true
}

func SonarrWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Handle Sonarr Webhook", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Library from the URL params
	librarySection := r.URL.Query().Get("library")
	if librarySection == "" {
		logAction.SetError("Missing library parameter", "The 'library' URL parameter is required", nil)
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Decode into typed struct
	var payload SonarrWebHookOnUpgradePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Run Validation on Payload to determine if we could/should proceed
	// We only want to run this when EventType is Download
	if payload.EventType != "Download" {
		logAction.AppendResult("event_type", payload.EventType)
		w.WriteHeader(http.StatusOK)
		return
	} else if payload.Series == (SonarrSeries{}) || payload.Series.TmdbID == 0 {
		logAction.AppendResult("series_info", "missing or invalid")
		w.WriteHeader(http.StatusOK)
		return
	} else if len(payload.Episodes) == 0 {
		logAction.AppendResult("episodes_info", "no episodes in payload")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Now we want to check if this TMDB ID + Library Title exists in the Aura DB
	items, _, _, pageErr := api.DB_GetAllItemsWithFilter(
		ctx,
		strconv.Itoa(payload.Series.TmdbID), // searchTMDBID
		librarySection,                      // searchLibrary
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
		"",                                  // posterSetID
	)
	if pageErr.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	} else if len(items) > 1 {
		logAction.SetError("Multiple DB items found", "Multiple matching items found in the database for the given TMDB ID and library", map[string]any{
			"tmdb_id": payload.Series.TmdbID,
			"library": librarySection,
			"count":   len(items),
		})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// We have a single matching item
	dbItem := items[0]

	// Determine whether any set cares about titlecards/season posters
	hasTitleCardSelected := false
	hasSeasonPosterSelected := false
	for _, posterSet := range dbItem.PosterSets {
		if slices.Contains(posterSet.SelectedTypes, "titlecard") {
			hasTitleCardSelected = true
		}
		if slices.Contains(posterSet.SelectedTypes, "seasonPoster") || slices.Contains(posterSet.SelectedTypes, "specialSeasonPoster") {
			hasSeasonPosterSelected = true
		}
	}
	logAction.AppendResult("has_seasonposter_selected", hasSeasonPosterSelected)
	logAction.AppendResult("has_titlecards_selected", hasTitleCardSelected)

	// Respond to Sonarr immediately
	w.WriteHeader(http.StatusOK)

	if !hasTitleCardSelected && !hasSeasonPosterSelected {
		return
	}

	payloadCopy := payload
	hasTitleCardSelectedCopy := hasTitleCardSelected
	hasSeasonPosterSelectedCopy := hasSeasonPosterSelected

	// Run the rest in a goroutine
	go func(
		dbItem api.DBMediaItemWithPosterSets,
		librarySection string,
		payload SonarrWebHookOnUpgradePayload,
		hasSeasonPosterSelected bool,
		hasTitleCardSelected bool,
	) {
		bgCtx := context.Background()
		bgCtx, bgLd := logging.CreateLoggingContext(bgCtx, "Downloading Titlecard for Sonarr Upgrade Webhook")
		defer bgLd.Log()

		bgAction := bgLd.AddAction("Sonarr Webhook Background Task", logging.LevelInfo)
		bgCtx = logging.WithCurrentAction(bgCtx, bgAction)
		defer bgAction.Complete()

		switch payload.IsUpgrade {
		case true:
			if hasTitleCardSelected {
				SonarrProcessDownload(bgCtx, payload, librarySection, dbItem, hasTitleCardSelected)
			}
		case false:
			SonarrProcessNewDownload(bgCtx, payload, librarySection, dbItem, hasSeasonPosterSelected, hasTitleCardSelected)
		}
	}(dbItem, librarySection, payloadCopy, hasSeasonPosterSelectedCopy, hasTitleCardSelectedCopy)
}

func SonarrProcessNewDownload(ctx context.Context, payload SonarrWebHookOnUpgradePayload, librarySection string, dbItem api.DBMediaItemWithPosterSets, hasSeasonPosterSelected bool, hasTitleCardSelected bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Processing Sonarr New Download for %s", payload.Series.Title), logging.LevelDebug)

	// Determine if this is a new season
	sonarrSeasonNumber := payload.Episodes[0].SeasonNumber
	isNewSeason := true
	for _, season := range dbItem.MediaItem.Series.Seasons {
		if season.SeasonNumber == sonarrSeasonNumber {
			isNewSeason = false
			break
		}
	}
	logAction.AppendResult("is_new_season", isNewSeason)
	logAction.AppendResult("sonarr_season_number", sonarrSeasonNumber)

	// If it's a new season and we have season posters selected, download the season poster
	var mediaItem api.MediaItem
	if isNewSeason && hasSeasonPosterSelected {
		item, info, ok := fetchMediaItemWithCacheFallback(ctx, librarySection, dbItem.MediaItem.RatingKey, dbItem.TMDB_ID)
		if !ok {
			logAction.SetError("Media item fetch failed", "Could not fetch media item from media server or cache", map[string]any{
				"tmdb_id":          dbItem.TMDB_ID,
				"library":          librarySection,
				"fetch_from_cache": info.FromCache,
				"fetch_error":      info.FetchErr,
				"cache_hit":        info.CacheHit,
			})
			return
		}
		mediaItem = item

		if mediaItem.TMDB_ID == "" && mediaItem.RatingKey == "" {
			logAction.SetError("Fetched media item missing IDs", "Fetched media item is missing both TMDB ID and Rating Key", map[string]any{
				"tmdb_id": dbItem.TMDB_ID,
				"library": librarySection,
			})
			return
		}

		// Loop through poster sets
		for _, posterSet := range dbItem.PosterSets {
			selected := posterSet.SelectedTypes

			if !slices.Contains(selected, "seasonPoster") && !slices.Contains(selected, "specialSeasonPoster") {
				logAction.AppendWarning("skip_posterSet_no_seasonposter_"+posterSet.PosterSet.ID, "No seasonPoster or specialSeasonPoster selected")
				continue
			}

			// We need to get the latest version of the MediUX poster set
			latestSet, Err := api.Mediux_FetchShowSetByID(ctx, dbItem.MediaItem.LibraryTitle, dbItem.MediaItem.TMDB_ID, posterSet.PosterSet.ID)
			if Err.Message != "" {
				logAction.AppendWarning("Mediux_FetchShowSetByID_error_"+posterSet.PosterSet.ID, Err.Message)
				continue
			}
			posterSet.PosterSet = latestSet

			var downloadFile api.PosterFile
			for _, seasonFile := range posterSet.PosterSet.SeasonPosters {
				if seasonFile.Season.Number != sonarrSeasonNumber {
					continue
				}
				downloadFile = seasonFile
			}

			if downloadFile.ID == "" {
				logAction.AppendWarning("no_seasonposter_found_"+posterSet.PosterSet.ID, fmt.Sprintf(
					"No season poster found for season %d in posterSet %s",
					sonarrSeasonNumber,
					posterSet.PosterSet.ID,
				))
				continue
			}

			downloadFileName := api.MediaServer_GetFileDownloadName(downloadFile)
			logAction.AppendResult("downloading_seasonposter_"+posterSet.PosterSet.ID, fmt.Sprintf(
				"Downloading season poster posterSet=%s fileId=%s fileName=%s season=%d",
				posterSet.PosterSet.ID,
				downloadFile.ID,
				downloadFileName,
				sonarrSeasonNumber,
			))
			Err = api.CallDownloadAndUpdatePosters(ctx, mediaItem, downloadFile)
			if Err.Message != "" {
				logAction.AppendWarning("CallDownloadAndUpdatePosters_error_"+posterSet.PosterSet.ID, Err.Message)
				return
			}

			api.DeleteTempImageForNextLoad(ctx, downloadFile, mediaItem.RatingKey)

			go func() {
				SendFileDownloadNotification(mediaItem.Title, posterSet.PosterSet.ID, downloadFile, payload.IsUpgrade)
			}()
		}
	} else {
		logAction.AppendResult("skip_seasonposter_download", fmt.Sprintf(
			"Skipping season poster download isNewSeason=%v hasSeasonPosterSelected=%v",
			isNewSeason,
			hasSeasonPosterSelected,
		))
	}

	// Next steps
	if isNewSeason && hasSeasonPosterSelected && hasTitleCardSelected {
		logAction.AppendResult("proceed_titlecards_download", "New season with season posters and titlecards selected -> SonarrProcessDownload")
		SonarrProcessDownload(ctx, payload, librarySection, dbItem, hasTitleCardSelected)
		return
	} else if !isNewSeason && hasTitleCardSelected {
		logAction.AppendResult("existing_season_titlecards", "Existing season with titlecards selected -> SonarrProcessDownload")
		SonarrProcessDownload(ctx, payload, librarySection, dbItem, hasTitleCardSelected)
		return
	} else if isNewSeason && hasSeasonPosterSelected && !hasTitleCardSelected {
		logAction.AppendResult("new_season_seasonposters_only", "New season with season posters only -> update DB item with refreshed mediaItem")
		newDBItem := dbItem
		newDBItem.MediaItem = mediaItem
		newDBItem.MediaItemJSON = ""
		Err := api.DB_InsertAllInfoIntoTables(ctx, newDBItem)
		if Err.Message != "" {
			logAction.AppendResult("db_update_error", Err.Message)
		} else {
			logAction.AppendResult("db_update_success", "DB updated successfully")
		}
	}
}

func SonarrProcessDownload(ctx context.Context, payload SonarrWebHookOnUpgradePayload, librarySection string, dbItem api.DBMediaItemWithPosterSets, hasTitleCardSelected bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Processing Sonarr Download for %s", payload.Series.Title), logging.LevelDebug)

	// Since this is an upgrade, we need to give the Media Server some time to process the download
	sleepAction := logAction.AddSubAction("Sleeping for 15 seconds to give time for media server to update", logging.LevelTrace)
	time.Sleep(15 * time.Second)
	sleepAction.Complete()

	// Get the latest information from the Media Server
	item, info, ok := fetchMediaItemWithCacheFallback(ctx, librarySection, dbItem.MediaItem.RatingKey, dbItem.TMDB_ID)
	if !ok {
		logAction.AppendWarning("media_item_fetch_failed", fmt.Sprintf("fetchErr=%s", info.FetchErr))
		return
	}
	mediaItem := item

	if mediaItem.TMDB_ID == "" && mediaItem.RatingKey == "" {
		logAction.AppendWarning("missing_tmdbid_ratingkey", "Fetched media item missing tmdbId and ratingKey; aborting")
		return
	}

	// Loop through poster sets
	for _, posterSet := range dbItem.PosterSets {
		// We only care about titlecards for the downloaded episodes
		if !slices.Contains(posterSet.SelectedTypes, "titlecard") {
			logAction.AppendResult("skip_posterSet_no_titlecards_"+posterSet.PosterSet.ID, "No titlecards selected")
			continue
		}

		// We need to get the latest version of the MediUX poster set
		latestSet, Err := api.Mediux_FetchShowSetByID(ctx, dbItem.MediaItem.LibraryTitle, dbItem.MediaItem.TMDB_ID, posterSet.PosterSet.ID)
		if Err.Message != "" {
			logAction.AppendWarning("Mediux_FetchShowSetByID_error", Err.Message)
			continue
		}
		posterSet.PosterSet = latestSet

		// For each episode in the payload, find and download the corresponding titlecard
		for _, sonarrEpisode := range payload.Episodes {
			var downloadFile api.PosterFile
			for _, episodeFile := range posterSet.PosterSet.TitleCards {
				if episodeFile.Episode.EpisodeNumber != sonarrEpisode.EpisodeNumber || episodeFile.Episode.SeasonNumber != sonarrEpisode.SeasonNumber {
					continue
				}
				downloadFile = episodeFile
			}

			if downloadFile.ID == "" {
				logAction.AppendWarning("no_titlecard_found", fmt.Sprintf("No titlecard found posterSet=%s for S%dE%d", posterSet.PosterSet.ID, sonarrEpisode.SeasonNumber, sonarrEpisode.EpisodeNumber))
				continue
			}

			downloadFileName := api.MediaServer_GetFileDownloadName(downloadFile)
			logAction.AppendResult("downloading_titlecard_"+posterSet.PosterSet.ID+"_"+downloadFile.ID, fmt.Sprintf("Downloading titlecard fileName=%s S%dE%d", downloadFileName, sonarrEpisode.SeasonNumber, sonarrEpisode.EpisodeNumber))

			Err = api.CallDownloadAndUpdatePosters(ctx, mediaItem, downloadFile)
			if Err.Message != "" {
				logAction.AppendWarning("CallDownloadAndUpdatePosters_error", Err.Message)
				return
			}

			api.DeleteTempImageForNextLoad(ctx, downloadFile, mediaItem.RatingKey)
			go func() {
				SendFileDownloadNotification(mediaItem.Title, posterSet.PosterSet.ID, downloadFile, payload.IsUpgrade)
			}()
		}

		// Update the database with the latest information
		newDBItem := dbItem
		newDBItem.MediaItem = mediaItem
		newDBItem.MediaItemJSON = ""
		newDBItem.PosterSets = make([]api.DBPosterSetDetail, len(dbItem.PosterSets))
		for i, ps := range dbItem.PosterSets {
			if ps.PosterSetID == posterSet.PosterSetID {
				// Update only the set being worked on
				newDBItem.PosterSets[i] = api.DBPosterSetDetail{
					PosterSetID:    latestSet.ID,
					PosterSet:      latestSet,
					AutoDownload:   ps.AutoDownload,
					SelectedTypes:  ps.SelectedTypes,
					LastDownloaded: time.Now().Format("2006-01-02 15:04:05"),
				}
			} else {
				// Keep other sets unchanged
				newDBItem.PosterSets[i] = ps
			}
		}
		Err = api.DB_InsertAllInfoIntoTables(ctx, newDBItem)
		if Err.Message != "" {
			logAction.AppendWarning("DB_InsertAllInfoIntoTables_error", Err.Message)
			logAction.AppendResult("db_update_error", Err.Message)
		} else {
			logAction.AppendResult("db_update_success", fmt.Sprintf("DB updated successfully posterSet=%s", posterSet.PosterSet.ID))
		}
	}
}

func SendFileDownloadNotification(itemTitle, posterSetID string, posterFile api.PosterFile, isUpgrade bool) {
	if len(api.Global_Config.Notifications.Providers) == 0 || api.Global_Config.Notifications.Enabled == false {
		return
	}

	notificationTitle := "Sonarr Episode"
	if isUpgrade {
		notificationTitle += " - Upgrade"
	} else {
		notificationTitle += " - New Download"
	}
	messageBody := fmt.Sprintf(
		"%s (Set: %s) - %s",
		itemTitle,
		posterSetID,
		api.MediaServer_GetFileDownloadName(posterFile),
	)

	imageURL := fmt.Sprintf("%s/%s?v=%s&key=jpg",
		"https://images.mediux.io/assets",
		posterFile.ID,
		posterFile.Modified.Format("20060102150405"),
	)

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send File Download Message")
	logAction := ld.AddAction("Sending File Download Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	// Send a notification to all configured providers
	for _, provider := range api.Global_Config.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				api.Notification_SendDiscordMessage(
					ctx,
					provider.Discord,
					messageBody,
					imageURL,
					notificationTitle,
				)
			case "Pushover":
				api.Notification_SendPushoverMessage(
					ctx,
					provider.Pushover,
					messageBody,
					imageURL,
					notificationTitle,
				)
			case "Gotify":
				api.Notification_SendGotifyMessage(
					ctx,
					provider.Gotify,
					messageBody,
					imageURL,
					notificationTitle,
				)
			case "Webhook":
				api.Notification_SendWebhookMessage(
					ctx,
					provider.Webhook,
					messageBody,
					imageURL,
					notificationTitle,
				)
			}
		}
	}
}
