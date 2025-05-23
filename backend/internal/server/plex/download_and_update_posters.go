package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"aura/internal/utils"
	"errors"
	"fmt"
	"os"
	"path"
)

func DownloadAndUpdatePosters(plex modals.MediaItem, file modals.PosterFile) logging.ErrorLog {

	if !config.Global.SaveImageNextToContent {
		logErr := UpdateSetOnly(plex, file)
		if logErr.Err != nil {
			return logErr
		}
		return logging.ErrorLog{}
	}

	// Check if the temporary folder has the image
	// If it does, we don't need to download it again
	// If it doesn't, we need to download it
	// The image is saved in the temp-images/mediux/full folder with the file ID as the name
	formatDate := file.Modified.Format("20060102")
	fileName := fmt.Sprintf("%s_%s.jpg", file.ID, formatDate)
	filePath := path.Join(mediux.MediuxFullTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	var imageData []byte
	if !exists {
		// Check if the temporary folder exists
		logErr := utils.CheckFolderExists(mediux.MediuxFullTempImageFolder)
		if logErr.Err != nil {
			return logErr
		}
		// Download the image from Mediux
		imageData, _, logErr = mediux.FetchImage(file.ID, formatDate, true)
		if logErr.Err != nil {
			return logErr
		}
		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to write image to %s: %v", filePath, err)}}
		}
		logging.LOG.Trace(fmt.Sprintf("Image %s downloaded and saved to temporary folder", file.ID))
	} else {
		// Read the contents of the file
		var err error
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to read image from %s: %v", filePath, err)}}
		}
	}

	// Check the Plex Item type
	// If it is a movie or show, handle the poster/backdrop/seasonPoster/titlecard accordingly

	newFilePath := ""
	newFileName := ""

	if plex.Type == "movie" {

		// If item.Movie is nil, get the movie from the library
		if plex.Movie == nil {
			var logErr logging.ErrorLog
			plex, logErr = FetchItemContent(plex.RatingKey)
			if logErr.Err != nil {
				return logErr
			}
		}

		// Handle movie-specific logic
		newFilePath = path.Dir(plex.Movie.File.Path)
		if file.Type == "poster" {
			newFileName = "poster.jpg"
		} else if file.Type == "backdrop" {
			newFileName = "backdrop.jpg"
		}
	} else if plex.Type == "show" {
		// Handle show-specific logic
		newFilePath = plex.Series.Location
		if file.Type == "poster" {
			newFileName = "poster.jpg"
		} else if file.Type == "backdrop" {
			newFileName = "backdrop.jpg"
		} else if file.Type == "seasonPoster" || file.Type == "specialSeasonPoster" {
			seasonNumberConvention := config.Global.MediaServer.SeasonNamingConvention
			var seasonNumber string
			if seasonNumberConvention == "1" {
				seasonNumber = fmt.Sprintf("%d", file.Season.Number)
			} else {
				seasonNumber = utils.Get2DigitNumber(int64(file.Season.Number))
			}
			newFilePath = path.Join(newFilePath, fmt.Sprintf("Season %s", seasonNumber))
			newFileName = fmt.Sprintf("Season%s.jpg", seasonNumber)
		} else if file.Type == "titlecard" {
			// For titlecards, get the file path from Plex
			episodePath := getEpisodePathFromPlex(plex, file)
			if episodePath != "" {
				newFilePath = path.Dir(episodePath)
				newFileName = path.Base(episodePath)
				newFileName = newFileName[:len(newFileName)-len(path.Ext(newFileName))] + ".jpg"
			} else {
				return logging.ErrorLog{Err: fmt.Errorf("episode path is empty"), Log: logging.Log{Message: "No episode path found for titlecard"}}
			}
		}
	}

	if config.Global.Dev.Enable {
		newFilePath = path.Join(config.Global.Dev.LocalPath, newFilePath)
	}

	// Ensure the directory exists
	err := os.MkdirAll(newFilePath, os.ModePerm)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to create directory %s", newFilePath)}}
	}

	// Create the new file
	newFile, err := os.Create(path.Join(newFilePath, newFileName))
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to create file %s", newFileName)}}
	}
	defer newFile.Close()

	// Write the image data to the file
	_, err = newFile.Write(imageData)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to write image data to file %s", newFileName)}}
	}

	// If cacheImages is False, delete the image from the temporary folder
	// 		This is to prevent the temporary folder from getting too large
	if !config.Global.CacheImages {
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
		return logging.ErrorLog{Err: errors.New("media not found"),
			Log: logging.Log{Message: "Media Item not found"}}
	}

	refreshPlexItem(itemRatingKey)
	posterKey, logErr := getPosters(itemRatingKey)
	if logErr.Err != nil {
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

		return logging.ErrorLog{
			Err: logErr.Err,
			Log: logging.Log{Message: fmt.Sprintf("Failed to find poster for item '%s'", posterName)},
		}
	}
	setPoster(itemRatingKey, posterKey, file.Type)

	// If config.Global.Kometa.RemoveLabels is true, remove the labels specified in the config
	if config.Global.Kometa.RemoveLabels {
		for _, label := range config.Global.Kometa.Labels {
			logErr := removeLabel(itemRatingKey, label)
			if logErr.Err != nil {
				logging.LOG.Warn(fmt.Sprintf("Failed to remove label '%s': %v", label, logErr.Err))
			}
		}
	}

	return logging.ErrorLog{}
}
