package mediaserver

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/server/emby_jellyfin"
	"aura/internal/server/plex"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetImageFromMediaServer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	Err := logging.NewStandardError()

	ratingKey := chi.URLParam(r, "ratingKey")
	imageType := chi.URLParam(r, "imageType")
	if ratingKey == "" || imageType == "" {
		Err.Message = "Missing rating key or image type in URL"
		Err.HelpText = "Ensure the URL contains both rating key and image type parameters."
		Err.Details = fmt.Sprintf("Received ratingKey: %s, imageType: %s", ratingKey, imageType)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	} else if imageType != "poster" && imageType != "backdrop" {
		Err.Message = "Invalid image type"
		Err.HelpText = "Image type must be either 'poster' or 'backdrop'."
		Err.Details = fmt.Sprintf("Received image type: %s", imageType)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	var mediaServer mediaserver_shared.MediaServer
	var tmpFolder string
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
		tmpFolder = plex.PlexTempImageFolder
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
		tmpFolder = emby_jellyfin.EmbyJellyTempImageFolder
	default:
		Err.Message = "Unsupported media server type"
		Err.HelpText = fmt.Sprintf("The media server type '%s' is not supported.", config.Global.MediaServer.Type)
		Err.Details = fmt.Sprintf("Received media server type: %s", config.Global.MediaServer.Type)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Check if the temporary folder has the image
	fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, imageType)
	filePath := path.Join(tmpFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	if exists {
		// Serve the image from the temporary folder
		imagePath := path.Join(tmpFolder, fileName)
		http.ServeFile(w, r, imagePath)
		return
	}

	// If the image does not exist, then get it from the media server
	imageData, Err := mediaServer.FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// If the user has enabled caching, then save the image to the temporary folder
	if config.Global.Images.CacheImages.Enabled {
		Err = utils.CheckFolderExists(tmpFolder)
		if Err.Message != "" {
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
		imagePath := path.Join(tmpFolder, fileName)
		err := os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			Err.Message = "Failed to write image to temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the temporary folder %s is writable.", tmpFolder)
			Err.Details = fmt.Sprintf("Error writing image to %s: %v", imagePath, err)
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	// Write the image data to the response
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
