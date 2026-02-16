package routes_images

import (
	"aura/cache"
	"aura/logging"
	"aura/mediux"
	"aura/utils"
	"aura/utils/httpx"
	"net/http"
)

func GetMediuxImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Mediux Image", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	assetID := r.URL.Query().Get("asset_id")
	modifiedDate := r.URL.Query().Get("modified_date")
	quality := r.URL.Query().Get("quality") // thumb, optimized, original

	if quality == "" {
		quality = "thumb"
	}
	if quality != "thumb" && quality != "optimized" && quality != "original" {
		logAction.SetError("Invalid quality parameter", "quality must be one of: thumb, optimized, original", map[string]any{
			"quality": quality,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	formatDate := utils.ConvertDateStringToTime(modifiedDate).Format("20060102150405")

	var imageQuality mediux.ImageQuality
	switch quality {
	case "thumb":
		imageQuality = mediux.ImageQualityThumb
	case "optimized":
		imageQuality = mediux.ImageQualityOptimized
	case "original":
		imageQuality = mediux.ImageQualityOriginal
	}

	imageData, imageType, Err := mediux.GetImage(ctx, assetID, formatDate, imageQuality)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func GetMediuxAvatarImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Mediux Avatar Image", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	avatarID := r.URL.Query().Get("avatar_id")
	username := r.URL.Query().Get("username")

	if avatarID == "" && username == "" {
		httpx.SendResponse(w, ld, nil)
		return
	} else if avatarID == "" && username != "" {
		// Lookup avatar ID from username
		cachedUser, found := cache.MediuxUsers.GetMediuxUserByUsername(username)
		if found && cachedUser.Avatar != "" {
			avatarID = cachedUser.Avatar
		} else {
			httpx.SendResponse(w, ld, nil)
			return
		}
	}

	imageData, imageType, Err := mediux.GetAvatarImage(ctx, avatarID)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)

}
