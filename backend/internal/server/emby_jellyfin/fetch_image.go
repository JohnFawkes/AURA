package emby_jellyfin

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
)

var EmbyJellyTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	EmbyJellyTempImageFolder = path.Join(configPath, "temp-images", "emby-jellyfin")
}

// https://emby.mooseboxx.com/emby/Items/1396/Images/Backdrop/0?tag=e317e1fc568744cd70cdd45ef99903ca&maxWidth=1920&quality=70
func FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting %s for rating key: %s", imageType, ratingKey))

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("Items/%s/Images/%s", ratingKey, imageType), config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return nil, logErr
	}

	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return nil, logErr
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	// Check if the response is empty
	if len(body) == 0 {
		return nil, logging.ErrorLog{Err: fmt.Errorf("received empty response from %s server", config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received empty response from %s server", config.Global.MediaServer.Type)}}
	}

	// Return the image data
	return body, logging.ErrorLog{}
}
