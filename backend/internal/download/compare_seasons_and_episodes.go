package download

import "aura/internal/modals"

func AddedMoreSeasonsOrEpisodes(dbSavedItem, latestMediaItem modals.MediaItem) bool {
	// Check if the latest media item has more seasons or episodes than the saved item
	if latestMediaItem.Series.SeasonCount > dbSavedItem.Series.SeasonCount ||
		latestMediaItem.Series.EpisodeCount > dbSavedItem.Series.EpisodeCount {
		return true
	}
	return false
}

func CheckSeasonAdded(seasonNumber int, dbSavedItem, latestMediaItem modals.MediaItem) bool {
	// First check if the season exists in dbSavedItem
	seasonExistsInDB := false
	if dbSavedItem.Series.Seasons != nil {
		for _, season := range dbSavedItem.Series.Seasons {
			if season.SeasonNumber == seasonNumber {
				seasonExistsInDB = true
				break
			}
		}
	}

	// If season doesn't exist in DB, check if it exists in latest
	if !seasonExistsInDB {
		for _, season := range latestMediaItem.Series.Seasons {
			if season.SeasonNumber == seasonNumber {
				return true
			}
		}
	}

	return false
}

func CheckEpisodeAdded(seasonNumber, episodeNumber int, dbSavedItem, latestMediaItem modals.MediaItem) bool {
	// First check if episode exists in dbSavedItem
	episodeExistsInDB := false
	if dbSavedItem.Series.Seasons != nil {
		for _, season := range dbSavedItem.Series.Seasons {
			if season.SeasonNumber == seasonNumber {
				for _, episode := range season.Episodes {
					if episode.EpisodeNumber == episodeNumber {
						episodeExistsInDB = true
						break
					}
				}
				break // Found the season, no need to keep looking
			}
		}
	}

	// If episode doesn't exist in DB, check if it exists in latest
	if !episodeExistsInDB {
		for _, season := range latestMediaItem.Series.Seasons {
			if season.SeasonNumber == seasonNumber {
				for _, episode := range season.Episodes {
					if episode.EpisodeNumber == episodeNumber {
						return true // Episode exists in latest but not in DB
					}
				}
				break // Found the season, no need to keep looking
			}
		}
	}

	return false
}
