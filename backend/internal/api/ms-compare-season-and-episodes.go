package api

import (
	"fmt"
)

func MS_AddedMoreSeasonsOrEpisodes(dbSavedItem, latestMediaItem MediaItem) bool {
	// Check if the latest media item has more seasons or episodes than the saved item
	if dbSavedItem.Series == nil || latestMediaItem.Series == nil {
		return false
	}

	if latestMediaItem.Series.SeasonCount > dbSavedItem.Series.SeasonCount ||
		latestMediaItem.Series.EpisodeCount > dbSavedItem.Series.EpisodeCount {
		return true
	}

	return false
}

func MS_CheckEpisodePathChanges(dbSavedItem, latestMediaItem MediaItem) bool {
	// Check if any episode paths have changed
	if dbSavedItem.Series == nil || latestMediaItem.Series == nil {
		return false
	}

	for _, latestSeason := range latestMediaItem.Series.Seasons {
		var dbSeason *MediaItemSeason
		for _, s := range dbSavedItem.Series.Seasons {
			if s.SeasonNumber == latestSeason.SeasonNumber {
				dbSeason = &s
				break
			}
		}
		if dbSeason == nil {
			// Season doesn't exist in DB, so new season added
			return true
		}
		for _, latestEpisode := range latestSeason.Episodes {
			var dbEpisode *MediaItemEpisode
			for _, e := range dbSeason.Episodes {
				if e.EpisodeNumber == latestEpisode.EpisodeNumber {
					dbEpisode = &e
					break
				}
			}
			if dbEpisode == nil {
				// Episode doesn't exist in DB, so new episode added
				return true
			}
			if dbEpisode.File.Path != "" && latestEpisode.File.Path != "" && dbEpisode.File.Path != latestEpisode.File.Path {
				// Episode path has changed
				return true
			}
		}
	}
	return false
}

func CheckSeasonExistsAndAdded(seasonNumber int, dbSavedItem, latestMediaItem MediaItem) (existsInDB, existsInLatest bool) {
	// First check if the season exists in dbSavedItem
	existsInDB = false
	existsInLatest = false

	if dbSavedItem.Series.Seasons != nil {
		for _, season := range dbSavedItem.Series.Seasons {
			if season.SeasonNumber == seasonNumber {
				existsInDB = true
				break
			}
		}
	}

	if !existsInDB {
		for _, season := range latestMediaItem.Series.Seasons {
			if season.SeasonNumber == seasonNumber {
				existsInLatest = true
				return existsInDB, existsInLatest // Season was added, download it
			}
		}
	}

	// Season doesn't exist in DB or latest (skip download)
	return existsInDB, existsInLatest
}

func CheckEpisodeExistsAddedAndPath(seasonNumber, episodeNumber int, dbSavedItem, latestMediaItem MediaItem) (existsInDB, existsInLatest bool, pathChanged string) {
	existsInDB = false
	existsInLatest = false
	pathChanged = ""
	var episodePathInDB, episodePathInLatest string

	// Find episode in DB
	for _, season := range dbSavedItem.Series.Seasons {
		if season.SeasonNumber == seasonNumber {
			for _, episode := range season.Episodes {
				if episode.EpisodeNumber == episodeNumber {
					existsInDB = true
					episodePathInDB = episode.File.Path
					break
				}
			}
			break
		}
	}

	// Find episode in latest
	if latestMediaItem.Series == nil {
		return existsInDB, existsInLatest, pathChanged
	}

	for _, season := range latestMediaItem.Series.Seasons {
		if season.SeasonNumber == seasonNumber {
			for _, episode := range season.Episodes {
				if episode.EpisodeNumber == episodeNumber {
					episodePathInLatest = episode.File.Path
					existsInLatest = true
					break
				}
			}
			break
		}
	}

	// If the episode exists in both, check if path changed
	if existsInDB && existsInLatest {
		// Check if the path has changed
		if episodePathInDB != "" && episodePathInLatest != "" && episodePathInDB != episodePathInLatest {
			pathChanged = fmt.Sprintf("S%sE%s path changed from\n'%s'\nto\n'%s'", Util_Format_Get2DigitNumber(int64(seasonNumber)), Util_Format_Get2DigitNumber(int64(episodeNumber)), episodePathInDB, episodePathInLatest)
		}
	}

	return existsInDB, existsInLatest, pathChanged
}
