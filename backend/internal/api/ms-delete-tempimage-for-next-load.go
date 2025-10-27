package api

import (
	"aura/internal/logging"
	"fmt"
	"os"
)

func DeleteTempImageForNextLoad(file PosterFile, ratingKey string) {
	if file.Type == "poster" || file.Type == "backdrop" {
		var tmpFolder string
		switch Global_Config.MediaServer.Type {
		case "Plex":
			tmpFolder = PlexTempImageFolder
		case "Emby", "Jellyfin":
			tmpFolder = EmbyJellyTempImageFolder
		default:
			logging.LOG.Error(fmt.Sprintf("Unsupported media server type: %s", Global_Config.MediaServer.Type))
			return
		}

		// Delete the poster and backdrop temporary image file
		fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, file.Type)
		filePath := fmt.Sprintf("%s/%s", tmpFolder, fileName)
		exists := Util_File_CheckIfFileExists(filePath)
		if exists {
			logging.LOG.Trace(fmt.Sprintf("Deleting temporary image %s", fileName))
			err := os.Remove(filePath)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to delete temporary image %s: %s", fileName, err.Error()))
			}
		}

		otherFile := "backdrop"
		if file.Type == "backdrop" {
			otherFile = "poster"
		}
		// Delete the other temporary image file
		otherFileName := fmt.Sprintf("%s_%s.jpg", ratingKey, otherFile)
		otherFilePath := fmt.Sprintf("%s/%s", tmpFolder, otherFileName)
		exists = Util_File_CheckIfFileExists(otherFilePath)
		if exists {
			logging.LOG.Trace(fmt.Sprintf("Deleting temporary image %s", otherFileName))
			err := os.Remove(otherFilePath)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to delete temporary image %s: %s", otherFileName, err.Error()))
			}
		}
	}
}
