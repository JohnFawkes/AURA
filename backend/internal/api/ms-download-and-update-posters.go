package api

import (
	"aura/internal/logging"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"
)

func (p *PlexServer) DownloadAndUpdatePosters(mediaItem MediaItem, file PosterFile) logging.StandardError {
	// Download and update the item on Plex
	Err := Plex_DownloadAndUpdatePosters(mediaItem, file)
	if Err.Message != "" {
		return Err
	}
	return logging.StandardError{}
}

func (e *EmbyJellyServer) DownloadAndUpdatePosters(mediaItem MediaItem, file PosterFile) logging.StandardError {
	// Download and update the item on Emby/Jellyfin
	Err := EJ_DownloadAndUpdatePosters(mediaItem, file)
	if Err.Message != "" {
		return Err
	}
	return logging.StandardError{}
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func CallDownloadAndUpdatePosters(mediaItem MediaItem, file PosterFile) logging.StandardError {
	mediaServer, Err := GetMediaServerInterface(Config_MediaServer{})
	if Err.Message != "" {
		return Err
	}
	return mediaServer.DownloadAndUpdatePosters(mediaItem, file)
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func Plex_DownloadAndUpdatePosters(plex MediaItem, file PosterFile) logging.StandardError {

	if !Global_Config.Images.SaveImagesLocally.Enabled {
		Err := Plex_UpdatePosterViaMediuxURL(plex, file)
		if Err.Message != "" {
			return Err
		}
		return logging.StandardError{}
	}
	Err := logging.NewStandardError()

	// Check if the temporary folder has the image
	// If it does, we don't need to download it again
	// If it doesn't, we need to download it
	// The image is saved in the temp-images/mediux/full folder with the file ID as the name
	formatDate := file.Modified.Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.jpg", file.ID, formatDate)
	filePath := path.Join(MediuxFullTempImageFolder, fileName)
	exists := Util_File_CheckIfFileExists(filePath)
	var imageData []byte
	if !exists {
		// Check if the temporary folder exists
		Err := Util_File_CheckFolderExists(MediuxFullTempImageFolder)
		if Err.Message != "" {
			return Err
		}
		// Download the image from Mediux
		imageData, _, Err = Mediux_GetImage(file.ID, formatDate, "original")
		if Err.Message != "" {
			return Err
		}
		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			Err.Message = "Failed to save image to temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the temporary folder %s is writable.", MediuxFullTempImageFolder)
			Err.Details = map[string]any{
				"error":    err.Error(),
				"filePath": filePath,
			}
			return Err
		}
		logging.LOG.Trace(fmt.Sprintf("Image %s downloaded and saved to temporary folder", file.ID))
	} else {
		// Read the contents of the file
		var err error
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			Err.Message = "Failed to read image from temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the temporary folder %s is accessible.", MediuxFullTempImageFolder)
			Err.Details = map[string]any{
				"error":    err.Error(),
				"filePath": filePath,
			}
			return Err
		}
	}

	// Check the Plex Item type
	// If it is a movie or show, handle the poster/backdrop/seasonPoster/titlecard accordingly

	newFilePath := ""
	newFileName := ""

	switch plex.Type {
	case "movie":
		// If item.Movie is nil, get the movie from the library
		if plex.Movie == nil {
			logging.LOG.Debug(fmt.Sprintf("Fetching full movie details for '%s'", plex.Title))
			plex, Err = Plex_FetchItemContent(plex.RatingKey)
			if Err.Message != "" {
				return Err
			}
		}

		// Handle movie-specific logic
		newFilePath = path.Dir(plex.Movie.File.Path)
		switch file.Type {
		case "poster":
			newFileName = "poster.jpg"
		case "backdrop":
			newFileName = "backdrop.jpg"
		}
	case "show":
		// Handle show-specific logic
		newFilePath = plex.Series.Location
		switch file.Type {
		case "poster":
			newFileName = "poster.jpg"
		case "backdrop":
			newFileName = "backdrop.jpg"
		case "seasonPoster", "specialSeasonPoster":
			seasonNumberConvention := Global_Config.MediaServer.SeasonNamingConvention
			var seasonNumber string
			if seasonNumberConvention == "1" {
				seasonNumber = fmt.Sprintf("%d", file.Season.Number)
			} else {
				seasonNumber = Util_Format_Get2DigitNumber(int64(file.Season.Number))
			}
			newFilePath = path.Join(newFilePath, fmt.Sprintf("Season %s", seasonNumber))
			newFileName = fmt.Sprintf("Season%s.jpg", seasonNumber)
		case "titlecard":
			// For titlecards, get the file path from Plex
			episodePath := Plex_GetEpisodePathFromPlex(plex, file)
			if episodePath != "" {
				newFilePath = path.Dir(episodePath)
				newFileName = path.Base(episodePath)
				newFileName = newFileName[:len(newFileName)-len(path.Ext(newFileName))] + ".jpg"
			} else {
				Err.Message = "Episode path is empty for titlecard"
				Err.HelpText = "Ensure the episode path is correctly set in Plex."
				Err.Details = map[string]any{
					"episode": file.Episode,
					"season":  file.Season,
					"show":    plex.Title,
				}
				return Err
			}
		}
	}

	logging.LOG.Debug(fmt.Sprintf("Plex File Path: %s", newFilePath))

	if Global_Config.Images.SaveImagesLocally.Enabled && Global_Config.Images.SaveImagesLocally.Path != "" {
		// Build newFilePath based on library, content, and config path
		libraryRoot := ""
		libSection, exists := Global_Cache_LibraryStore.GetSectionByTitle(plex.LibraryTitle)
		logging.LOG.Trace(fmt.Sprintf("Library section '%s' exists: %t", plex.LibraryTitle, exists))

		if exists && libSection.Path != "" {
			// Library exists in cache (e.g. /data/media/movies or /data/media/shows)
			libraryRoot = libSection.Path
			logging.LOG.Trace(fmt.Sprintf("Library Root: %s", libraryRoot))

			// Get last part of library root (e.g. "movies" or "shows")
			libraryPath := path.Base(libraryRoot)
			logging.LOG.Trace(fmt.Sprintf("Library Path: %s", libraryPath))

			// Get path before library name (e.g. /data/media/)
			remainingLibraryPath := strings.TrimSuffix(libraryRoot, libraryPath)
			logging.LOG.Trace(fmt.Sprintf("Remaining Library Path: %s", remainingLibraryPath))

			// Get relative path from newFilePath (e.g. movies/Inception (2020), shows/Breaking Bad/Season 01)
			relativePath := strings.TrimPrefix(newFilePath, remainingLibraryPath)
			relativePath = strings.TrimLeft(relativePath, string(os.PathSeparator))
			logging.LOG.Debug(fmt.Sprintf("Relative path: %s", relativePath))

			// Final path: /local/images/movies/Inception (2020), etc.
			newFilePath = path.Join(Global_Config.Images.SaveImagesLocally.Path, relativePath)
		} else {
			logging.LOG.Warn(fmt.Sprintf("Library '%s' not found in cache, using Plex paths", plex.LibraryTitle))

			// Fallback: build path from Plex info
			libraryPath := ""
			contentPath := ""
			seasonPath := ""

			if file.Type == "poster" || file.Type == "backdrop" || plex.Type == "movie" {
				// For movies or posters/backdrops
				contentPath = path.Base(newFilePath)
				logging.LOG.Trace(fmt.Sprintf("Content Path: %s", contentPath))

				libraryPath = path.Base(path.Dir(newFilePath))
				logging.LOG.Trace(fmt.Sprintf("Library Path: %s", libraryPath))

				// Final path: /local/images/movies/Inception (2020)
				newFilePath = path.Join(Global_Config.Images.SaveImagesLocally.Path, libraryPath, contentPath)
			} else if plex.Type == "show" && (file.Type == "seasonPoster" || file.Type == "specialSeasonPoster" || file.Type == "titlecard") {
				// For shows with seasonPoster/specialSeasonPoster/titlecard
				seasonPath = path.Base(newFilePath)
				logging.LOG.Trace(fmt.Sprintf("Season Path: %s", seasonPath))

				contentPath = path.Base(path.Dir(newFilePath))
				logging.LOG.Trace(fmt.Sprintf("Content Path: %s", contentPath))

				libraryPath = path.Base(path.Dir(path.Dir(newFilePath)))
				logging.LOG.Trace(fmt.Sprintf("Library Path: %s", libraryPath))

				// Final path: /local/images/shows/Breaking Bad/Season 01
				newFilePath = path.Join(Global_Config.Images.SaveImagesLocally.Path, libraryPath, contentPath, seasonPath)
			} else {
				// Error: unable to determine path
				Err.Message = "Failed to determine library path"
				Err.HelpText = "Ensure the library exists in Plex and has a valid path."
				Err.Details = map[string]any{
					"library": plex.Title,
					"type":    plex.Type,
					"file":    file.Type,
				}
				return Err
			}
		}
	}

	// Ensure the directory exists
	err := os.MkdirAll(newFilePath, os.ModePerm)
	if err != nil {
		Err.Message = "Failed to create directory for new file"
		Err.HelpText = fmt.Sprintf("Ensure the directory %s is writable.", newFilePath)
		Err.Details = map[string]any{
			"error":    err.Error(),
			"filePath": newFilePath,
		}
		return Err
	}

	// Create the new file
	newFile, err := os.Create(path.Join(newFilePath, newFileName))
	if err != nil {
		Err.Message = "Failed to create new file for image"
		Err.HelpText = fmt.Sprintf("Ensure the directory %s is writable.", newFilePath)
		Err.Details = map[string]any{
			"error":    err.Error(),
			"filePath": newFilePath,
		}
		return Err
	}
	defer newFile.Close()

	// Write the image data to the file
	_, err = newFile.Write(imageData)
	if err != nil {
		Err.Message = "Failed to write image data to new file"
		Err.HelpText = fmt.Sprintf("Ensure the directory %s is writable.", newFilePath)
		Err.Details = map[string]any{
			"error":    err.Error(),
			"filePath": newFilePath,
		}
		return Err
	}
	logging.LOG.Info(fmt.Sprintf("Image saved to %s", path.Join(newFilePath, newFileName)))

	// If cacheImages is False, delete the image from the temporary folder
	// 		This is to prevent the temporary folder from getting too large
	if !Global_Config.Images.CacheImages.Enabled {
		logging.LOG.Trace("Deleting image from temporary folder")
		err := os.Remove(filePath)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to delete image from temporary folder: %v", err))
		}
	}

	// Determine the itemRatingKey
	itemRatingKey := Plex_GetItemRatingKey(plex, file)
	if itemRatingKey == "" {
		logging.LOG.Error(fmt.Sprintf("Item rating key is empty for '%s' not found", plex.Title))
		Err.Message = "Item rating key is empty"
		Err.HelpText = "Ensure the item exists in Plex and has a valid rating key."
		Err.Details = map[string]any{
			"item":      plex.Title,
			"type":      plex.Type,
			"file":      file.Type,
			"id":        file.ID,
			"ratingKey": plex.RatingKey,
		}
		return Err
	}

	// If Save Image Next to Content is enabled and the Path is set, set the poster in Plex via the Mediux URL
	// When the Path is set, the image is saved in a different location than Plex expects it to be.
	// So we need to upload the image to Plex via the Mediux URL.
	if Global_Config.Images.SaveImagesLocally.Enabled && Global_Config.Images.SaveImagesLocally.Path != "" {
		logging.LOG.Debug("Setting image via Mediux URL since Save Images Locally is enabled and Path is set")
		Err := Plex_UpdatePosterViaMediuxURL(plex, file)
		if Err.Message != "" {
			return Err
		}
		return logging.StandardError{}
	}

	Plex_RefreshItem(itemRatingKey)
	posterKey, Err := Plex_GetPoster(itemRatingKey)
	failedOnGetPosters := false
	if Err.Message != "" {
		failedOnGetPosters = true
		var posterName string
		switch file.Type {
		case "poster":
			posterName = "Poster"
		case "backdrop":
			posterName = "Backdrop"
		case "seasonPoster", "specialSeasonPoster":
			posterName = fmt.Sprintf("Season %s Poster", Util_Format_Get2DigitNumber(int64(file.Season.Number)))
		case "titlecard":
			posterName = fmt.Sprintf("S%sE%s Titlecard", Util_Format_Get2DigitNumber(int64(file.Episode.SeasonNumber)), Util_Format_Get2DigitNumber(int64(file.Episode.EpisodeNumber)))
		}
		Err.Message = fmt.Sprintf("Failed to get %s key for item", posterName)
		Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists in Plex and has a valid %s.", itemRatingKey, posterName)
		Err.Details = map[string]any{
			"error":     Err.Details,
			"item":      plex.Title,
			"type":      plex.Type,
			"file":      file.Type,
			"id":        file.ID,
			"ratingKey": itemRatingKey,
		}
	}

	// If failedOnGetPosters is true, just upload the image MediUX URL
	if failedOnGetPosters {
		logging.LOG.Debug("Setting image via Mediux URL since failed to get poster key from Plex")
		Err := Plex_UpdatePosterViaMediuxURL(plex, file)
		if Err.Message != "" {
			return Err
		}
		return logging.StandardError{}
	}

	logging.LOG.Debug("Setting image via Plex upload since poster key was found")
	Plex_SetPoster(itemRatingKey, posterKey, file.Type)
	// Handle Labels and Tags in a separate go routine for quicker finish
	go func() {
		Err = Plex_HandleLabels(plex)
		if Err.Message != "" {
			logging.LOG.Warn(Err.Message)
		}
		SR_CallHandleTags(plex)
	}()

	return logging.StandardError{}
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func EJ_DownloadAndUpdatePosters(item MediaItem, file PosterFile) logging.StandardError {
	Err := logging.NewStandardError()

	itemRatingKey := EJ_GetItemRatingKey(item, file)
	if itemRatingKey == "" {
		Err.Message = "Media not found"
		Err.HelpText = "Ensure the item exists in the Emby/Jellyfin library."
		Err.Details = map[string]any{
			"error": fmt.Sprintf("No matching media item found for file ID: %s", file.ID),
		}
		return Err
	}

	// Check if the temporary folder has the image
	// If it does, we don't need to download it again
	// If it doesn't, we need to download it
	// The image is saved in the temp-images/mediux/full folder with the file ID as the name
	formatDate := file.Modified.Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.jpg", file.ID, formatDate)
	filePath := path.Join(MediuxFullTempImageFolder, fileName)
	exists := Util_File_CheckIfFileExists(filePath)
	var imageData []byte
	if !exists {
		// Check if the temporary folder exists
		Err := Util_File_CheckFolderExists(MediuxFullTempImageFolder)
		if Err.Message != "" {
			return Err
		}
		// Download the image from Mediux
		imageData, _, Err = Mediux_GetImage(file.ID, formatDate, "original")
		if Err.Message != "" {
			return Err
		}
		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			Err.Message = "Failed to write image to temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and writable.", MediuxFullTempImageFolder)
			Err.Details = map[string]any{
				"error":    fmt.Sprintf("Error writing image: %v", err),
				"filePath": filePath,
			}
			return Err
		}
		logging.LOG.Trace(fmt.Sprintf("Image %s downloaded and saved to temporary folder", file.ID))
	} else {
		// Read the contents of the file
		var err error
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			Err.Message = "Failed to read image from temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and readable.", MediuxFullTempImageFolder)
			Err.Details = map[string]any{
				"error":    fmt.Sprintf("Error reading image: %v", err),
				"filePath": filePath,
			}
			return Err
		}
	}

	var posterType string
	if file.Type == "backdrop" {
		posterType = "Backdrop"
	} else {
		posterType = "Primary"
	}

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Items/%s/Images/%s", itemRatingKey, posterType), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return Err
	}

	// Encode the image data as Base64
	base64ImageData := base64.StdEncoding.EncodeToString(imageData)

	// Make a POST request to the Emby/Jellyfin server
	headers := map[string]string{
		"Content-Type": "image/jpeg",
	}
	response, _, Err := MakeHTTPRequest(baseURL.String(), "POST", headers, 60, []byte(base64ImageData), "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 && response.StatusCode != 204 {
		Err.Message = "Failed to upload image to Emby/Jellyfin"
		Err.HelpText = "Ensure the Emby/Jellyfin server is reachable and the API key is valid."
		Err.Details = map[string]any{
			"status_code": response.StatusCode,
			"status":      response.Status,
			"body":        response.Body,
			"item":        item.Title,
			"type":        item.Type,
			"file":        file.Type,
			"id":          file.ID,
			"ratingKey":   itemRatingKey,
		}
		return Err
	}

	logging.LOG.Info(fmt.Sprintf("%s uploaded to %s for item '%s'", strings.ToTitle(file.Type), Global_Config.MediaServer.Type, item.Title))

	// Handle Labels in a separate go routine for quicker finish
	go func() {
		SR_CallHandleTags(item)
	}()

	return logging.StandardError{}
}
