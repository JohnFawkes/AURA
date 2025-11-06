package api

import (
	"aura/internal/logging"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func (p *PlexServer) DownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo {
	return Plex_DownloadAndUpdatePosters(ctx, mediaItem, file)
}

func (e *EmbyJellyServer) DownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo {
	return EJ_DownloadAndUpdatePosters(ctx, mediaItem, file)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallDownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return Err
	}

	return mediaServer.DownloadAndUpdatePosters(ctx, mediaItem, file)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_DownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo {
	if !Global_Config.Images.SaveImagesLocally.Enabled {
		Err := Plex_UpdateImageViaMediuxURL(ctx, mediaItem, file)
		if Err.Message != "" {
			return Err
		}
		return logging.LogErrorInfo{}
	}

	var posterImageType string
	switch file.Type {
	case "poster":
		posterImageType = "Poster"
	case "backdrop":
		posterImageType = "Backdrop"
	case "seasonPoster", "specialSeasonPoster":
		posterImageType = fmt.Sprintf("Season %d Poster", file.Season.Number)
	case "titlecard":
		posterImageType = fmt.Sprintf("S%dE%d Titlecard", file.Episode.SeasonNumber, file.Episode.EpisodeNumber)
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading and Updating %s in Plex", posterImageType), logging.LevelInfo)
	defer logAction.Complete()

	// Get the Image from MediUX
	// Mediux_GetImage will handle checking the temp folder and caching based on config
	formatDate := file.Modified.Format("20060102150405")
	imageData, _, Err := Mediux_GetImage(ctx, file.ID, formatDate, MediuxImageQualityOriginal)
	if Err.Message != "" {
		return Err
	}

	// Check the Plex Item type
	// If it is a movie or show, handle the poster/backdrop/seasonPoster/titlecard accordingly
	newFilePath := ""
	newFileName := ""

	getFilePathAction := logAction.AddSubAction(fmt.Sprintf("Determining File Path for %s (%s)", posterImageType, mediaItem.Title), logging.LevelDebug)
	switch mediaItem.Type {
	case "movie":
		// If item.Movie is nil, get the movie from the library
		if mediaItem.Movie == nil {
			mediaItem, Err = Plex_FetchItemContent(ctx, mediaItem.RatingKey)
			if Err.Message != "" {
				return Err
			}
		}

		// Handle Movie Specific Logic
		newFilePath = path.Dir(mediaItem.Movie.File.Path)
		switch file.Type {
		case "poster":
			newFileName = "poster.jpg"
		case "backdrop":
			newFileName = "backdrop.jpg"
		}
	case "show":
		// Handle show-specific logic
		newFilePath = mediaItem.Series.Location
		switch file.Type {
		case "poster":
			newFileName = "poster.jpg"
		case "backdrop":
			newFileName = "backdrop.jpg"
		case "seasonPoster", "specialSeasonPoster":
			seasonNumberConvention := Global_Config.Images.SaveImagesLocally.SeasonNamingConvention
			var seasonNumber string
			if seasonNumberConvention == "1" {
				seasonNumber = fmt.Sprintf("%d", file.Season.Number)
			} else {
				seasonNumber = Util_Format_Get2DigitNumber(int64(file.Season.Number))
			}
			// Try and get the season folder from the first episode
			seasonPath := ""
		foundSeasonPath:
			for _, season := range mediaItem.Series.Seasons {
				if season.SeasonNumber == file.Season.Number {
					if len(season.Episodes) > 0 {
						for _, episode := range season.Episodes {
							if episode.File.Path != "" {
								episodeFilePath := episode.File.Path
								seasonPath = path.Dir(episodeFilePath)
								getFilePathAction.AppendResult("season_path", fmt.Sprintf("found season path from S%d E%d", episode.SeasonNumber, episode.EpisodeNumber))
								break foundSeasonPath
							}
						}
					}
				}
			}
			if seasonPath == "" {
				seasonPath = path.Join(newFilePath, fmt.Sprintf("Season %s", seasonNumber))
				getFilePathAction.AppendResult("season_path", "built season path from series path and season number")
			}
			newFilePath = seasonPath
			newFileName = fmt.Sprintf("Season%s.jpg", seasonNumber)
		case "titlecard":
			episodeNamingConvention := Global_Config.Images.SaveImagesLocally.EpisodeNamingConvention
			// For titlecards, get the file path from Plex
			episodePath := Plex_GetEpisodePathFromPlex(mediaItem, file)
			if episodePath != "" {
				newFilePath = path.Dir(episodePath)
				switch episodeNamingConvention {
				case "match":
					newFileName = path.Base(episodePath)
					newFileName = newFileName[:len(newFileName)-len(path.Ext(newFileName))] + ".jpg"
				case "static":
					var seasonNumber string
					var episodeNumber string
					if Global_Config.Images.SaveImagesLocally.SeasonNamingConvention == "1" {
						seasonNumber = fmt.Sprintf("%d", file.Episode.SeasonNumber)
						episodeNumber = fmt.Sprintf("%d", file.Episode.EpisodeNumber)
					} else {
						seasonNumber = Util_Format_Get2DigitNumber(int64(file.Episode.SeasonNumber))
						episodeNumber = Util_Format_Get2DigitNumber(int64(file.Episode.EpisodeNumber))
					}
					newFileName = fmt.Sprintf("S%sE%s.jpg", seasonNumber, episodeNumber)
				default:
					getFilePathAction.SetError("Invalid Episode Naming Convention",
						"EpisodeNamingConvention must be either 'match' or 'static'",
						map[string]any{
							"EpisodeNamingConvention": episodeNamingConvention,
							"SeasonNamingConvention":  Global_Config.Images.SaveImagesLocally.SeasonNamingConvention,
						})
					return *getFilePathAction.Error
				}
			} else {
				getFilePathAction.SetError("Failed to determine file path for titlecard",
					"Could not find episode path in Plex data",
					map[string]any{
						"rating_key": mediaItem.RatingKey,
					})
				return *getFilePathAction.Error
			}
		}
	default:
		getFilePathAction.SetError("Unsupported Media Item Type for Poster Update",
			"Only 'movie' and 'show' types are supported for poster updates",
			map[string]any{
				"item_type": mediaItem.Type,
			})
		return *getFilePathAction.Error
	}
	getFilePathAction.AppendResult("newFilePath", newFilePath)
	getFilePathAction.Complete()

	isCustomLocalPath := false
	if Global_Config.Images.SaveImagesLocally.Enabled && Global_Config.Images.SaveImagesLocally.Path != "" {
		isCustomLocalPath = true
		newPathAction := logAction.AddSubAction("Building New File Path for Local Image Save", logging.LevelDebug)
		// Build newFilePath based on library, content, and config path
		libraryRoot := ""
		libSection, exists := Global_Cache_LibraryStore.GetSectionByTitle(mediaItem.LibraryTitle)
		newPathAction.AppendResult("library_title", mediaItem.LibraryTitle)
		if exists && libSection.Path != "" {
			newPathAction.AppendResult("library_found_in_cache", true)
			// Library exists in cache (e.g. /data/media/movies or /data/media/shows)
			libraryRoot = libSection.Path
			newPathAction.AppendResult("library_root", libraryRoot)

			// Get last part of library root (e.g. "movies" or "shows")
			libraryPath := path.Base(libraryRoot)
			newPathAction.AppendResult("library_path", libraryPath)

			// Get path before library name (e.g. /data/media/)
			remainingLibraryPath := strings.TrimSuffix(libraryRoot, libraryPath)
			newPathAction.AppendResult("remaining_library_path", remainingLibraryPath)

			// Get relative path from newFilePath (e.g. movies/Inception (2020), shows/Breaking Bad/Season 01)
			relativePath := strings.TrimPrefix(newFilePath, remainingLibraryPath)
			relativePath = strings.TrimLeft(relativePath, string(os.PathSeparator))
			newPathAction.AppendResult("relative_path", relativePath)

			// Final path: /local/images/movies/Inception (2020), etc.
			newFilePath = path.Join(Global_Config.Images.SaveImagesLocally.Path, relativePath)
			newPathAction.AppendResult("final_path", newFilePath)
		} else {
			newPathAction.AppendResult("library_found_in_cache", false)
			// Fallback: build path from Plex info
			libraryPath := ""
			contentPath := ""
			seasonPath := ""

			if file.Type == "poster" || file.Type == "backdrop" || mediaItem.Type == "movie" {
				// For movies or posters/backdrops
				contentPath = path.Base(newFilePath)
				newPathAction.AppendResult("content_path", contentPath)

				libraryPath = path.Base(path.Dir(newFilePath))
				newPathAction.AppendResult("library_path", libraryPath)

				// Final path: /local/images/movies/Inception (2020)
				newFilePath = path.Join(Global_Config.Images.SaveImagesLocally.Path, libraryPath, contentPath)
				newPathAction.AppendResult("final_path", newFilePath)
			} else if mediaItem.Type == "show" && (file.Type == "seasonPoster" || file.Type == "specialSeasonPoster" || file.Type == "titlecard") {
				// For shows with seasonPoster/specialSeasonPoster/titlecard
				seasonPath = path.Base(newFilePath)
				newPathAction.AppendResult("season_path", seasonPath)

				contentPath = path.Base(path.Dir(newFilePath))
				newPathAction.AppendResult("content_path", contentPath)

				libraryPath = path.Base(path.Dir(path.Dir(newFilePath)))
				newPathAction.AppendResult("library_path", libraryPath)

				// Final path: /local/images/shows/Breaking Bad/Season 01
				newFilePath = path.Join(Global_Config.Images.SaveImagesLocally.Path, libraryPath, contentPath, seasonPath)
				newPathAction.AppendResult("final_path", newFilePath)
			} else {
				// Error: unable to determine path
				newPathAction.SetError("Failed to determine library path", "Ensure the library exists in Plex and has a valid path",
					map[string]any{
						"title": mediaItem.Title,
						"type":  mediaItem.Type,
						"file":  file.Type,
					})
				return *newPathAction.Error
			}
		}
		newPathAction.Complete()
	}

	createFileAction := logAction.AddSubAction("Saving Image to New File Path", logging.LevelDebug)

	// Ensure the directory exists
	err := os.MkdirAll(newFilePath, os.ModePerm)
	if err != nil {
		createFileAction.SetError("Failed to create directory", "Ensure the directory can be created",
			map[string]any{
				"error": err.Error(),
				"path":  newFilePath,
			})
		return *createFileAction.Error
	}

	// Create the new file
	newFile, err := os.Create(path.Join(newFilePath, newFileName))
	if err != nil {
		createFileAction.SetError("Failed to create file", "Ensure the file can be created",
			map[string]any{
				"error": err.Error(),
				"path":  newFilePath,
			})
		return *createFileAction.Error
	}
	defer newFile.Close()

	// Write the image data to the new file
	_, err = newFile.Write(imageData)
	if err != nil {
		createFileAction.SetError("Failed to write image data to file", "Ensure the file is writable",
			map[string]any{
				"error": err.Error(),
				"path":  newFilePath,
			})
		return *createFileAction.Error
	}
	createFileAction.AppendResult("saved_file_path", path.Join(newFilePath, newFileName))
	createFileAction.Complete()

	// Determine the itemRatingKey
	getRatingKeyAction := logAction.AddSubAction("Determining Item Rating Key for Poster Update", logging.LevelDebug)
	itemRatingKey := Plex_GetItemRatingKey(mediaItem, file)
	if itemRatingKey == "" {
		getRatingKeyAction.SetError("Failed to determine item rating key", "Ensure the media item and file data are correct",
			map[string]any{
				"media_item": mediaItem,
				"file":       file,
			})
		return *getRatingKeyAction.Error
	}
	getRatingKeyAction.AppendResult("item_rating_key", itemRatingKey)
	getRatingKeyAction.Complete()
	// If Save Image Next to Content is enabled and the Path is set, set the poster in Plex via the Mediux URL
	// When the Path is set, the image is saved in a different location than Plex expects it to be.
	// So we need to upload the image to Plex via the Mediux URL.
	if isCustomLocalPath {
		Err := Plex_UpdateImageViaMediuxURL(ctx, mediaItem, file)
		if Err.Message != "" {
			return Err
		}
		return logging.LogErrorInfo{}
	}

	// Refresh the Plex item
	Plex_RefreshItem(ctx, itemRatingKey)

	// Get the Plex Poster Key
	posterKey, Err := Plex_GetPoster(ctx, itemRatingKey, file.Type)
	failedToGetPosterKey := false
	if Err.Message != "" {
		failedToGetPosterKey = true
	} else {
		logAction.AppendResult("poster_key", posterKey)
	}

	// If failedOnGetPosters is true, use the Mediux URL to set the poster
	if failedToGetPosterKey {
		Err := Plex_UpdateImageViaMediuxURL(ctx, mediaItem, file)
		if Err.Message != "" {
			return Err
		}
		return logging.LogErrorInfo{}
	}

	Err = Plex_SetPoster(ctx, itemRatingKey, posterKey, file.Type)
	if Err.Message != "" {
		return Err
	}

	return logging.LogErrorInfo{}
}

func EJ_DownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo {

	var posterImageType string
	switch file.Type {
	case "poster":
		posterImageType = "Poster"
	case "backdrop":
		posterImageType = "Backdrop"
	case "seasonPoster", "specialSeasonPoster":
		posterImageType = fmt.Sprintf("Season %d Poster", file.Season.Number)
	case "titlecard":
		posterImageType = fmt.Sprintf("S%dE%d Titlecard", file.Episode.SeasonNumber, file.Episode.EpisodeNumber)
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading and Updating %s in %s", posterImageType, Global_Config.MediaServer.Type), logging.LevelInfo)
	defer logAction.Complete()

	// Get the Item Rating Key from Emby/Jellyfin
	itemRatingKey := EJ_GetItemRatingKey(mediaItem, file)
	if itemRatingKey == "" {
		logAction.SetError("Failed to determine item rating key", "Ensure the media item and file data are correct",
			map[string]any{
				"media_item": mediaItem,
				"file":       file,
			})
		return *logAction.Error
	}

	// Get the Image from MediUX
	// Mediux_GetImage will handle checking the temp folder and caching based on config
	formatDate := file.Modified.Format("20060102150405")
	imageData, _, Err := Mediux_GetImage(ctx, file.ID, formatDate, MediuxImageQualityOriginal)
	if Err.Message != "" {
		return Err
	}

	var posterType string
	if file.Type == "backdrop" {
		posterType = "Backdrop"
	} else {
		posterType = "Primary"
	}

	// Make the URL
	uploadAction := logAction.AddSubAction(fmt.Sprintf("Uploading %s Image to %s", posterType, Global_Config.MediaServer.Type), logging.LevelDebug)
	defer uploadAction.Complete()

	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		uploadAction.SetError("Failed to parse Media Server URL",
			"Ensure the Media Server URL is valid",
			map[string]any{
				"error": err.Error(),
			})
		return *uploadAction.Error
	}
	u.Path = path.Join(u.Path, "Items", itemRatingKey, "Images", posterType)
	URL := u.String()

	// Encode the image data as Base64
	base64ImageData := base64.StdEncoding.EncodeToString(imageData)

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Add Content-Type Header
	headers["Content-Type"] = "image/jpeg"

	// Make a POST request to upload the image
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodPost, headers, 60, []byte(base64ImageData), Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		uploadAction.SetError("Failed to upload image to Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return *uploadAction.Error
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 && httpResp.StatusCode != 204 {
		uploadAction.SetError("Failed to upload image to Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return *uploadAction.Error
	}

	uploadAction.AppendResult("status_code", httpResp.StatusCode)
	uploadAction.AppendResult("message", fmt.Sprintf("%s uploaded to %s for item '%s'", strings.ToTitle(file.Type), Global_Config.MediaServer.Type, mediaItem.Title))
	uploadAction.Complete()
	return logging.LogErrorInfo{}
}
