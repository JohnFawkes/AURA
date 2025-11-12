package api

import (
	"aura/internal/logging"
	"context"
	"strings"
)

func Mediux_GetAllUsers(ctx context.Context, query string) (mediux_usernames []MediuxUserInfo, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Get All Mediux Users", logging.LevelInfo)
	defer logAction.Complete()

	var allUserNames = []string{
		"aloha_alona", "benno", "brikim", "canadianmovieguy", "casewicked", "cenodude", "cinemalounge_digital_posters",
		"cmdrriker", "codonzero22", "crackerjack", "cundallini", "danmark", "darkwesley", "davidlettice", "defluo",
		"drwhodalek", "edmondantes32", "etiennemza", "evergreen", "fireninja", "fwlolx", "fxckingham", "gemris",
		"hikari", "identitytheft", "ikonok", "jetram", "jezzfreeman", "jjwags23", "jmxd", "josephdoraniii", "jrsly",
		"kealoha_king", "kit4nos", "koltom", "kosherkale", "kwickflix", "ludacris", "mediuxking79", "minizaki",
		"mrmonkey", "oldmankestis", "olivier_286", "paliking", "parrillized", "pejamas", "plexstreaming", "praythea",
		"r3draid3r04", "randalpink87", "randombell", "recker_man", "rogue", "ruvilev", "sarrius", "senex", "simionski",
		"simmsuma", "slm", "stayk", "stoube26", "swaffelsmurf", "tallinex", "tarantula212", "thatja", "the_carlos4",
		"thebigg", "thelibrarian", "themedeusa", "valexv", "whiskeytime", "wholock", "wikid82", "willtong93", "wiwer",
		"wwalkerrrrr",
	}

	for _, username := range allUserNames {
		if strings.Contains(strings.ToLower(username), strings.ToLower(query)) {
			mediux_usernames = append(mediux_usernames, MediuxUserInfo{Username: username, Avatar: ""})
			if len(mediux_usernames) >= 10 {
				break
			}
		}
	}

	// Get the users Follow/Hide information to sort the results
	userFollowHide, Err := Mediux_FetchUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		return mediux_usernames, Err
	}

	var added = make(map[string]bool)
	var sortedUsernames []MediuxUserInfo

	// First add Followed users
	for _, info := range userFollowHide {
		if info.Follow {
			for _, user := range mediux_usernames {
				if strings.EqualFold(user.Username, info.Username) && !added[strings.ToLower(user.Username)] {
					sortedUsernames = append(sortedUsernames, info)
					added[strings.ToLower(user.Username)] = true
					break
				}
			}
		}
	}

	// Then add non-followed and non-hidden users
	for _, user := range mediux_usernames {
		found := false
		for _, info := range userFollowHide {
			if strings.EqualFold(user.Username, info.Username) {
				found = true
				break
			}
		}
		if !found && !added[strings.ToLower(user.Username)] {
			sortedUsernames = append(sortedUsernames, user)
			added[strings.ToLower(user.Username)] = true
		}
	}

	// Finally add Hidden users
	for _, info := range userFollowHide {
		if info.Hide {
			for _, user := range mediux_usernames {
				if strings.EqualFold(user.Username, info.Username) && !added[strings.ToLower(user.Username)] {
					sortedUsernames = append(sortedUsernames, info)
					added[strings.ToLower(user.Username)] = true
					break
				}
			}
		}
	}

	mediux_usernames = sortedUsernames
	return mediux_usernames, logging.LogErrorInfo{}
}
