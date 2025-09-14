package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"aura/internal/utils"
	"fmt"
	"os"
	"path"
)

func DownloadAndUpdatePosters(plex modals.MediaItem, file modals.PosterFile) logging.StandardError {

	if !config.Global.Images.SaveImageNextToContent.Enabled {
		Err := UpdateSetOnly(plex, file)
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
	filePath := path.Join(mediux.MediuxFullTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	var imageData []byte
	if !exists {
		// Check if the temporary folder exists
		Err := utils.CheckFolderExists(mediux.MediuxFullTempImageFolder)
		if Err.Message != "" {
			return Err
		}
		// Download the image from Mediux
		imageData, _, Err = mediux.FetchImage(file.ID, formatDate, "original")
		if Err.Message != "" {
			return Err
		}
		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {

			Err.Message = "Failed to save image to temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the temporary folder %s is writable.", mediux.MediuxFullTempImageFolder)
			Err.Details = fmt.Sprintf("Error saving image to %s: %v", filePath, err)
			return Err
		}
		logging.LOG.Trace(fmt.Sprintf("Image %s downloaded and saved to temporary folder", file.ID))
	} else {
		// Read the contents of the file
		var err error
		imageData, err = os.ReadFile(filePath)
		if err != nil {

			Err.Message = "Failed to read image from temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the temporary folder %s is accessible.", mediux.MediuxFullTempImageFolder)
			Err.Details = fmt.Sprintf("Error reading image from %s: %v", filePath, err)
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
			plex, Err = FetchItemContent(plex.RatingKey)
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
			seasonNumberConvention := config.Global.MediaServer.SeasonNamingConvention
			var seasonNumber string
			if seasonNumberConvention == "1" {
				seasonNumber = fmt.Sprintf("%d", file.Season.Number)
			} else {
				seasonNumber = utils.Get2DigitNumber(int64(file.Season.Number))
			}
			newFilePath = path.Join(newFilePath, fmt.Sprintf("Season %s", seasonNumber))
			newFileName = fmt.Sprintf("Season%s.jpg", seasonNumber)
		case "titlecard":
			// For titlecards, get the file path from Plex
			episodePath := getEpisodePathFromPlex(plex, file)
			if episodePath != "" {
				newFilePath = path.Dir(episodePath)
				newFileName = path.Base(episodePath)
				newFileName = newFileName[:len(newFileName)-len(path.Ext(newFileName))] + ".jpg"
			} else {

				Err.Message = "Episode path is empty for titlecard"
				Err.HelpText = "Ensure the episode path is correctly set in Plex."
				Err.Details = fmt.Sprintf("No episode path found for titlecard in show '%s'", plex.Title)
				return Err
			}
		}
	}

	if config.Global.Dev.Enabled {
		newFilePath = path.Join(config.Global.Dev.LocalPath, newFilePath)
	}

	// Ensure the directory exists
	err := os.MkdirAll(newFilePath, os.ModePerm)
	if err != nil {

		Err.Message = "Failed to create directory for new file"
		Err.HelpText = fmt.Sprintf("Ensure the directory %s is writable.", newFilePath)
		Err.Details = fmt.Sprintf("Error creating directory %s: %v", newFilePath, err)
		return Err
	}

	// Create the new file
	newFile, err := os.Create(path.Join(newFilePath, newFileName))
	if err != nil {

		Err.Message = "Failed to create new file for image"
		Err.HelpText = fmt.Sprintf("Ensure the directory %s is writable.", newFilePath)
		Err.Details = fmt.Sprintf("Error creating file %s: %v", newFileName, err)
		return Err
	}
	defer newFile.Close()

	// Write the image data to the file
	_, err = newFile.Write(imageData)
	if err != nil {

		Err.Message = "Failed to write image data to new file"
		Err.HelpText = fmt.Sprintf("Ensure the directory %s is writable.", newFilePath)
		Err.Details = fmt.Sprintf("Error writing image data to file %s: %v", newFileName, err)
		return Err
	}

	// If cacheImages is False, delete the image from the temporary folder
	// 		This is to prevent the temporary folder from getting too large
	if !config.Global.Images.CacheImages.Enabled {
		logging.LOG.Trace("Deleting image from temporary folder")
		err := os.Remove(filePath)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to delete image from temporary folder: %v", err))
		}
	}

	// Determine the itemRatingKey
	itemRatingKey := getItemRatingKey(plex, file)
	if itemRatingKey == "" {
		logging.LOG.Error(fmt.Sprintf("Item rating key is empty for '%s' not found", plex.Title))

		Err.Message = "Item rating key is empty"
		Err.HelpText = "Ensure the item exists in Plex and has a valid rating key."
		Err.Details = fmt.Sprintf("No rating key found for item '%s'", plex.Title)
		return Err
	}

	refreshPlexItem(itemRatingKey)
	posterKey, Err := getPosters(itemRatingKey)
	if Err.Message != "" {
		var posterName string
		if file.Type == "poster" {
			posterName = "Poster"
		} else if file.Type == "backdrop" {
			posterName = "Backdrop"
		} else if file.Type == "seasonPoster" || file.Type == "specialSeasonPoster" {
			posterName = fmt.Sprintf("Season %s Poster", utils.Get2DigitNumber(int64(file.Season.Number)))
		} else if file.Type == "titlecard" {
			posterName = fmt.Sprintf("S%sE%s Titlecard", utils.Get2DigitNumber(int64(file.Episode.SeasonNumber)), utils.Get2DigitNumber(int64(file.Episode.EpisodeNumber)))
		}

		Err.Message = fmt.Sprintf("Failed to get %s key for item", posterName)
		Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists in Plex and has a valid %s.", itemRatingKey, posterName)
		Err.Details = fmt.Sprintf("Error getting %s key for item with rating key %s: %v", posterName, itemRatingKey, Err)
		return Err
	}
	setPoster(itemRatingKey, posterKey, file.Type)

	// If config.Global.Kometa.RemoveLabels is true, remove the labels specified in the config
	if config.Global.Kometa.RemoveLabels {
		for _, label := range config.Global.Kometa.Labels {
			Err := removeLabel(itemRatingKey, label)
			if Err.Message != "" {
				logging.LOG.Warn(fmt.Sprintf("Failed to remove label '%s': %v", label, Err.Message))
			}
		}
	}

	return logging.StandardError{}
}
