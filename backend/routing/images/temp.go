package routes_images

import (
	"aura/config"
	"aura/logging"
	"aura/utils"
	"aura/utils/httpx"
	"fmt"
	"net/http"
	"path"
)

func DeleteTempImages(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Delete Temp Images", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	clearCount, Err := utils.ClearAllFilesFromFolder(ctx, path.Join(config.ConfigPath, "temp-images"))
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	if clearCount == 0 {
		httpx.SendResponse(w, ld, "No temporary images to delete")
		return
	} else {
		logAction.AppendResult("cleared_files", clearCount)
	}

	httpx.SendResponse(w, ld, fmt.Sprintf("Cleared %d temporary images", clearCount))
}
