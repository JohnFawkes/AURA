package api

import (
	"fmt"
)

func MS_AddedMoreSeasonsOrEpisodes(dbSavedItem, latestMediaItem MediaItem) bool {
	// Check if the latest media item has more seasons or episodes than the saved item
	if latestMediaItem.Series.SeasonCount > dbSavedItem.Series.SeasonCount ||
		latestMediaItem.Series.EpisodeCount > dbSavedItem.Series.EpisodeCount {
		return true
	}

	return false
}

func MS_CheckEpisodePathChanges(dbSavedItem, latestMediaItem MediaItem) bool {
	// Check if any episode paths have changed
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

func CheckSeasonAdded(seasonNumber int, dbSavedItem, latestMediaItem MediaItem, psFileReasons *PosterFileWithReason) {
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
				// Season was added
				psFileReasons.ReasonTitle = "Downloading - New Season Added"
				psFileReasons.ReasonDetails = fmt.Sprintf("Season %s was added", Util_Format_Get2DigitNumber(int64(seasonNumber)))
				return
			}
		}
	}
}

func CheckEpisodeAdded(seasonNumber, episodeNumber int, dbSavedItem, latestMediaItem MediaItem, psFileReasons *PosterFileWithReason) {
	var episodePathInDB, episodePathInLatest string
	episodeExistsInDB := false

	// Find episode in DB
	for _, season := range dbSavedItem.Series.Seasons {
		if season.SeasonNumber == seasonNumber {
			for _, episode := range season.Episodes {
				if episode.EpisodeNumber == episodeNumber {
					episodeExistsInDB = true
					episodePathInDB = episode.File.Path
					break
				}
			}
			break
		}
	}

	// Find episode in latest
	for _, season := range latestMediaItem.Series.Seasons {
		if season.SeasonNumber == seasonNumber {
			for _, episode := range season.Episodes {
				if episode.EpisodeNumber == episodeNumber {
					episodePathInLatest = episode.File.Path
					if !episodeExistsInDB {
						psFileReasons.ReasonTitle = "Downloading - New Episode Added"
						psFileReasons.ReasonDetails = fmt.Sprintf("S%sE%s was added", Util_Format_Get2DigitNumber(int64(seasonNumber)), Util_Format_Get2DigitNumber(int64(episodeNumber)))
						return
					}
					break
				}
			}
			break
		}
	}

	// If episode exists in both, check if path changed
	if episodeExistsInDB && episodePathInDB != "" && episodePathInLatest != "" && episodePathInDB != episodePathInLatest {
		psFileReasons.ReasonTitle = "Redownloading - Episode Path Changed"
		psFileReasons.ReasonDetails = fmt.Sprintf("S%sE%s path changed from\n'%s'\nto\n'%s'", Util_Format_Get2DigitNumber(int64(seasonNumber)), Util_Format_Get2DigitNumber(int64(episodeNumber)), episodePathInDB, episodePathInLatest)
		return
	}
}
