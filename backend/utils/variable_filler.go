package utils

import (
	"aura/config"
	"aura/models"
	"fmt"
	maps0 "maps"
	"regexp"
	"strings"
	"time"
)

var templateVarRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

// RenderTemplate replaces {{Variable}} tokens with values from vars.
// Unknown variables are left unchanged.
func RenderTemplate(input string, vars map[string]string) string {
	if input == "" {
		return input
	}

	return templateVarRegex.ReplaceAllStringFunc(input, func(token string) string {
		m := templateVarRegex.FindStringSubmatch(token)
		if len(m) < 2 {
			return token
		}
		key := strings.TrimSpace(m[1])
		if val, ok := vars[key]; ok {
			return val
		}
		return token
	})
}

func MergeTemplateVars(maps ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, m := range maps {
		maps0.Copy(out, m)
	}
	return out
}

func BaseTemplateVars() map[string]string {
	return map[string]string{
		"AppName":         config.AppName,
		"AppVersion":      config.AppVersion,
		"AppPort":         fmt.Sprintf("%d", config.AppPort),
		"AppAuthor":       config.AppAuthor,
		"AppLicense":      config.AppLicense,
		"MediaServerName": config.MediaServerName,
		"MediaServerType": config.Current.MediaServer.Type,
		"Timestamp":       time.Now().Format("2006-01-02 15:04:05"),
		"NewLine":         "\n",
		"Tab":             "\t",
	}
}

func TemplateVars_AppStartup(appName, appVersion string, appPort int) map[string]string {
	return MergeTemplateVars(
		BaseTemplateVars(),
	)
}

func TemplateVars_TestNotification() map[string]string {
	return MergeTemplateVars(
		BaseTemplateVars(),
	)
}

func TemplateVars_Autodownload(mediaItem models.MediaItem, setItem models.DBPosterSetDetail, image models.ImageFile, reasonTitle string, reason string) map[string]string {
	return MergeTemplateVars(
		BaseTemplateVars(),
		map[string]string{
			"MediaItemTitle":        mediaItem.Title,
			"MediaItemYear":         fmt.Sprintf("%d", mediaItem.Year),
			"MediaItemTMDBID":       mediaItem.TMDB_ID,
			"MediaItemLibraryTitle": mediaItem.LibraryTitle,
			"MediaItemRatingKey":    mediaItem.RatingKey,
			"MediaItemType":         mediaItem.Type,
			"SetID":                 setItem.ID,
			"SetTitle":              setItem.Title,
			"SetType":               setItem.Type,
			"SetCreator":            setItem.UserCreated,
			"ImageName":             GetFileDownloadName(mediaItem.Title, image),
			"ImageType":             image.Type,
			"ReasonTitle":           reasonTitle,
			"Reason":                reason,
		},
	)
}
