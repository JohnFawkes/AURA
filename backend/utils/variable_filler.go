package utils

import (
	"aura/config"
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
		"MediaServerName": config.MediaServerName,
		"MediaServerType": config.Current.MediaServer.Type,
		"Timestamp":       time.Now().Format("2006-01-02 15:04:05"),
	}
}

func TemplateVars_AppStartup(appName, appVersion string, appPort int) map[string]string {
	return MergeTemplateVars(
		BaseTemplateVars(),
		map[string]string{
			"AppName":    appName,
			"AppVersion": appVersion,
			"AppPort":    fmt.Sprintf("%d", appPort),
		},
	)
}

func TemplateVars_TestNotification() map[string]string {
	return MergeTemplateVars(
		BaseTemplateVars(),
	)
}
