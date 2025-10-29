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
		Err := Plex_UpdatePosterViaMediuxURL(ctx, mediaItem, file)
		if Err.Message != "" {
			return Err
		}
		return logging.LogErrorInfo{}
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, "Downloading and Updating Poster in Plex", logging.LevelInfo)
	defer logAction.Complete()

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
		downloadImageAction := logAction.AddSubAction(fmt.Sprintf("Downloading Poster Image '%s' from Mediux", file.ID), logging.LevelDebug)
		// Check if the temporary folder exists
		Err := Util_File_CheckFolderExists(ctx, MediuxFullTempImageFolder)
		if Err.Message != "" {
			return Err
		}

		// Download the image from Mediux
		imageData, _, Err = Mediux_GetImage(ctx, file.ID, formatDate, "original")
		if Err.Message != "" {
			return Err
		}

		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			downloadImageAction.SetError("Failed to save image to temporary folder",
				fmt.Sprintf("Ensure the path %s is accessible.", filePath),
				map[string]any{
					"error":   fmt.Sprintf("Error saving image: %v", err),
					"request": filePath,
				})
			return *downloadImageAction.Error
		}
		downloadImageAction.Complete()
	} else {
		readImageAction := logAction.AddSubAction(fmt.Sprintf("Reading Poster Image '%s' from Temporary Folder", file.ID), logging.LevelDebug)
		var err error
		// Read the image from the temporary folder
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			readImageAction.SetError("Failed to read image from temporary folder",
				fmt.Sprintf("Ensure the path %s is accessible.", filePath),
				map[string]any{
					"error":   fmt.Sprintf("Error reading image: %v", err),
					"request": filePath,
				})
			return *readImageAction.Error
		}
		readImageAction.Complete()
	}

	// Check the Plex Item type
	// If it is a movie or show, handle the poster/backdrop/seasonPoster/titlecard accordingly

	newFilePath := ""
	newFileName := ""

	getFilePathAction := logAction.AddSubAction(fmt.Sprintf("Determining File Path for %s (%s)", file.Type, mediaItem.Title), logging.LevelDebug)
	var Err logging.LogErrorInfo

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
			episodePath := Plex_GetEpisodePathFromPlex(mediaItem, file)
			if episodePath != "" {
				newFilePath = path.Dir(episodePath)
				newFileName = path.Base(episodePath)
				newFileName = newFileName[:len(newFileName)-len(path.Ext(newFileName))] + ".jpg"
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

	// If cacheImages is False, delete the image from the temporary folder
	// 		This is to prevent the temporary folder from getting too large
	if !Global_Config.Images.CacheImages.Enabled {
		deleteTempImageAction := logAction.AddSubAction("Deleting Temporary Image File", logging.LevelDebug)
		err := os.Remove(filePath)
		if err != nil {
			deleteTempImageAction.SetError("Failed to delete temporary image file", "Ensure the file can be deleted",
				map[string]any{
					"error": err.Error(),
					"path":  filePath,
				})
			return *deleteTempImageAction.Error
		}
		deleteTempImageAction.AppendResult("message", fmt.Sprintf("Temporary image file %s deleted", filePath))
		deleteTempImageAction.Complete()
	}

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
		Err := Plex_UpdatePosterViaMediuxURL(ctx, mediaItem, file)
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
		Err := Plex_UpdatePosterViaMediuxURL(ctx, mediaItem, file)
		if Err.Message != "" {
			return Err
		}
		return logging.LogErrorInfo{}
	}

	Err = Plex_SetPoster(ctx, itemRatingKey, posterKey, file.Type)
	if Err.Message != "" {
		return Err
	}

	// Handle Labels in a separate go routine for quicker finish
	go func() {
		Plex_HandleLabels(mediaItem)
	}()

	return logging.LogErrorInfo{}
}

func EJ_DownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading and Updating Poster in %s", Global_Config.MediaServer.Type), logging.LevelInfo)
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
		downloadImageAction := logAction.AddSubAction(fmt.Sprintf("Downloading Poster Image '%s' from Mediux", file.ID), logging.LevelDebug)
		// Check if the temporary folder exists
		Err := Util_File_CheckFolderExists(ctx, MediuxFullTempImageFolder)
		if Err.Message != "" {
			return Err
		}

		// Download the image from Mediux
		imageData, _, Err = Mediux_GetImage(ctx, file.ID, formatDate, "original")
		if Err.Message != "" {
			return Err
		}

		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			downloadImageAction.SetError("Failed to save image to temporary folder",
				fmt.Sprintf("Ensure the path %s is accessible.", filePath),
				map[string]any{
					"error":   fmt.Sprintf("Error saving image: %v", err),
					"request": filePath,
				})
			return *downloadImageAction.Error
		}
		downloadImageAction.Complete()
	} else {
		readImageAction := logAction.AddSubAction(fmt.Sprintf("Reading Poster Image '%s' from Temporary Folder", file.ID), logging.LevelDebug)
		var err error
		// Read the image from the temporary folder
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			readImageAction.SetError("Failed to read image from temporary folder",
				fmt.Sprintf("Ensure the path %s is accessible.", filePath),
				map[string]any{
					"error":   fmt.Sprintf("Error reading image: %v", err),
					"request": filePath,
				})
			return *readImageAction.Error
		}
		readImageAction.Complete()
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

	// Make a POST request to the Emby/Jellyfin server
	headers := map[string]string{
		"Content-Type": "image/jpeg",
	}

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
