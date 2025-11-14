package api

import (
	"aura/internal/logging"
	"context"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
)

func Mediux_SearchUsers(ctx context.Context, query string) (mediux_usernames []MediuxUserInfo, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Get All MediUX Users", logging.LevelInfo)
	defer logAction.Complete()

	mediuxUsers, Err := Mediux_GetAllUsers(ctx)
	if Err.Message != "" {
		return mediux_usernames, Err
	}

	for _, user := range mediuxUsers {
		if strings.Contains(strings.ToLower(user.Username), strings.ToLower(query)) {
			mediux_usernames = append(mediux_usernames, user)
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

	var (
		added         = make(map[string]bool)
		followedUsers []MediuxUserInfo
		normalUsers   []MediuxUserInfo
		hiddenUsers   []MediuxUserInfo
	)

	// First add Followed users
	for _, info := range userFollowHide {
		if info.Follow {
			for _, user := range mediux_usernames {
				if strings.EqualFold(user.Username, info.Username) && !added[strings.ToLower(user.Username)] {
					followedUsers = append(followedUsers, info)
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
			normalUsers = append(normalUsers, user)
			added[strings.ToLower(user.Username)] = true
		}
	}

	// Finally add Hidden users
	for _, info := range userFollowHide {
		if info.Hide {
			for _, user := range mediux_usernames {
				if strings.EqualFold(user.Username, info.Username) && !added[strings.ToLower(user.Username)] {
					hiddenUsers = append(hiddenUsers, info)
					added[strings.ToLower(user.Username)] = true
					break
				}
			}
		}
	}

	// Sort each section by TotalSets descending
	sort.SliceStable(followedUsers, func(i, j int) bool {
		return followedUsers[i].TotalSets > followedUsers[j].TotalSets
	})
	sort.SliceStable(normalUsers, func(i, j int) bool {
		return normalUsers[i].TotalSets > normalUsers[j].TotalSets
	})
	sort.SliceStable(hiddenUsers, func(i, j int) bool {
		return hiddenUsers[i].TotalSets > hiddenUsers[j].TotalSets
	})

	// Combine the sections
	sortedUsernames := append(followedUsers, normalUsers...)
	sortedUsernames = append(sortedUsernames, hiddenUsers...)
	mediux_usernames = sortedUsernames

	logAction.AppendResult("users_found_with_query", len(mediux_usernames))
	return mediux_usernames, logging.LogErrorInfo{}
}

func Mediux_GetAllUsers(ctx context.Context) (mediux_usernames []MediuxUserInfo, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting All Users from MediUX", logging.LevelDebug)
	defer logAction.Complete()

	type MediuxUserInfoResponse struct {
		ID             string `json:"id"`
		DateUpdated    string `json:"date_updated"`
		ShowSets       int    `json:"show_sets"`
		MovieSets      int    `json:"movie_sets"`
		CollectionSets int    `json:"collection_sets"`
		Username       string `json:"username"`
		Avatar         string `json:"avatar"`
		TotalSets      string `json:"total_sets"`
	}

	type MediuxUsersResponse struct {
		Data []MediuxUserInfoResponse `json:"data"`
	}

	// Construct the MediUX URL
	u, err := url.Parse(MediuxBaseURL)
	if err != nil {
		logAction.SetError("Failed to parse MediUX base URL", err.Error(), nil)
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "items", "contributors")
	q := u.Query()
	q.Set("limit", "-1")
	u.RawQuery = q.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("Authorization", Global_Config.Mediux.Token)

	// Make the API request to MediUX
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "MediUX")
	if logErr.Message != "" {
		return mediux_usernames, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response JSON
	var mediuxUsers MediuxUsersResponse
	Err = DecodeJSONBody(ctx, respBody, &mediuxUsers, "MediUX Get All Users Response")
	if Err.Message != "" {
		return mediux_usernames, Err
	}

	// Convert to MediuxUserInfo slice
	for _, user := range mediuxUsers.Data {
		if user.Username == "" || (user.TotalSets == "0" && user.ShowSets == 0 && user.MovieSets == 0 && user.CollectionSets == 0) {
			continue
		}
		mediux_usernames = append(mediux_usernames, MediuxUserInfo{
			ID:       user.ID,
			Username: user.Username,
			Avatar:   user.Avatar,
			TotalSets: func() int {
				if user.TotalSets != "" {
					val, err := strconv.Atoi(user.TotalSets)
					if err == nil {
						return val
					}
				}
				return user.ShowSets + user.MovieSets + user.CollectionSets
			}(),
		})
	}

	// Sort the users by Total Sets descending
	sort.SliceStable(mediux_usernames, func(i, j int) bool {
		return mediux_usernames[i].TotalSets > mediux_usernames[j].TotalSets
	})

	// Load the list of users into the cache for faster access later
	Global_Cache_MediuxUsers.StoreMediuxUsers(mediux_usernames)

	logAction.AppendResult("users_found", len(mediux_usernames))
	return mediux_usernames, logging.LogErrorInfo{}
}

func Mediux_PreloadAllUsersInCache() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Setting Up - Preload All MediUX Users in Cache")
	defer ld.Log()

	action := ld.AddAction("Preloading MediUX Users", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, action)
	defer action.Complete()

	_, Err := Mediux_GetAllUsers(ctx)
	if Err.Message != "" {
		return
	}
}
