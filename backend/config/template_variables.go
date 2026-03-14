package config

const (
	TemplateTypeAppStartup       = "app_startup"
	TemplateTypeTestNotification = "test_notification"
	TemplateTypeAutodownload     = "autodownload"
)

var BaseTemplateVariables = []string{
	"{{AppName}}",
	"{{AppVersion}}",
	"{{AppPort}}",
	"{{AppAuthor}}",
	"{{AppLicense}}",
	"{{MediaServerName}}",
	"{{MediaServerType}}",
	"{{Timestamp}}",
	"{{NewLine}}",
	"{{Tab}}",
}

var MediaItemVariables = []string{
	"{{MediaItemTitle}}",
	"{{MediaItemYear}}",
	"{{MediaItemTMDBID}}",
	"{{MediaItemLibraryTitle}}",
	"{{MediaItemRatingKey}}",
	"{{MediaItemType}}",
}

var SetItemVariables = []string{
	"{{SetID}}",
	"{{SetTitle}}",
	"{{SetType}}",
	"{{SetCreator}}",
}

var ImageVariables = []string{
	"{{ImageName}}",
	"{{ImageType}}",
}

var DownloadReasonVariables = []string{
	"{{ReasonTitle}}",
	"{{Reason}}",
}

type NotificationTemplateVariableCatalog struct {
	TemplateVariables map[string][]string `json:"template_variables"`
}

func AllowedTemplateVariables(templateType string) []string {
	switch templateType {
	case TemplateTypeAppStartup:
		return mergeTemplateVariableGroups(BaseTemplateVariables)
	case TemplateTypeTestNotification:
		return mergeTemplateVariableGroups(BaseTemplateVariables)
	case TemplateTypeAutodownload:
		return mergeTemplateVariableGroups(
			BaseTemplateVariables,
			MediaItemVariables,
			SetItemVariables,
			ImageVariables,
			DownloadReasonVariables,
		)
	default:
		return []string{}
	}
}

func GetNotificationTemplateVariableCatalog() NotificationTemplateVariableCatalog {
	return NotificationTemplateVariableCatalog{
		TemplateVariables: map[string][]string{
			TemplateTypeAppStartup:       AllowedTemplateVariables(TemplateTypeAppStartup),
			TemplateTypeTestNotification: AllowedTemplateVariables(TemplateTypeTestNotification),
			TemplateTypeAutodownload:     AllowedTemplateVariables(TemplateTypeAutodownload),
		},
	}
}

func mergeTemplateVariableGroups(groups ...[]string) []string {
	seen := make(map[string]struct{})
	merged := make([]string, 0)

	for _, group := range groups {
		for _, variable := range group {
			if _, ok := seen[variable]; ok {
				continue
			}
			seen[variable] = struct{}{}
			merged = append(merged, variable)
		}
	}

	return merged
}

func cloneStringSlice(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}
