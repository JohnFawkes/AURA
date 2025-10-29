package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"os"
)

func DeleteTempImageForNextLoad(ctx context.Context, file PosterFile, ratingKey string) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Delete Temporary Image For Next Load", logging.LevelDebug)
	defer logAction.Complete()

	if file.Type == "poster" || file.Type == "backdrop" {
		var tmpFolder string
		switch Global_Config.MediaServer.Type {
		case "Plex":
			tmpFolder = PlexTempImageFolder
		case "Emby", "Jellyfin":
			tmpFolder = EmbyJellyTempImageFolder
		default:
			logAction.AppendWarning("type", Global_Config.MediaServer.Type)
			logAction.AppendWarning("message", "Unknown Media Server type, cannot determine temp image folder")
			return
		}

		// Delete the poster and backdrop temporary image file
		fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, file.Type)
		filePath := fmt.Sprintf("%s/%s", tmpFolder, fileName)
		exists := Util_File_CheckIfFileExists(filePath)
		if exists {
			err := os.Remove(filePath)
			if err != nil {
				logAction.AppendWarning("file", filePath)
				logAction.AppendWarning("error", err.Error())
				logAction.AppendWarning("message", "Failed to delete temporary image")
			}
			logAction.AppendResult("deletedFile", fileName)
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
			err := os.Remove(otherFilePath)
			if err != nil {
				logAction.AppendWarning("file", otherFilePath)
				logAction.AppendWarning("error", err.Error())
				logAction.AppendWarning("message", "Failed to delete temporary image")
			}
			logAction.AppendResult("deletedFile", otherFileName)
		}
	}

}
