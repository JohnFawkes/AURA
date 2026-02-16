package mediux

import (
	"aura/cache"
	"aura/logging"
	"aura/models"
	"context"
	"sort"
	"strings"
)

func SearchUsersByUsername(ctx context.Context, query string) (matchedUsers []models.MediuxUserInfo, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Searching MediUX Users by Username", logging.LevelDebug)
	defer logAction.Complete()

	matchedUsers = []models.MediuxUserInfo{}
	Err = logging.LogErrorInfo{}

	// Get all users from the cache
	allUsers := cache.MediuxUsers.GetMediuxUsers()
	if len(allUsers) == 0 {
		logAction.SetError("No MediUX users found in cache", "Ensure that MediUX users have been loaded", nil)
		return matchedUsers, *logAction.Error
	}

	// Search for users whose usernames contain the query string (case-insensitive)
	for _, user := range allUsers {
		if strings.Contains(strings.ToLower(user.Username), strings.ToLower(query)) {
			matchedUsers = append(matchedUsers, user)
			if len(matchedUsers) >= 10 {
				break
			}
		}
	}

	// Get the users Follow/Hide information to sort the results
	userFollowHide, Err := GetUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		return matchedUsers, Err
	}

	var (
		added         = make(map[string]bool)
		followedUsers []models.MediuxUserInfo
		normalUsers   []models.MediuxUserInfo
		hiddenUsers   []models.MediuxUserInfo
	)

	for _, user := range matchedUsers {
		found := false
		isFollowed := false
		isHidden := false

		for _, info := range userFollowHide {
			if strings.EqualFold(user.Username, info.Username) {
				found = true
				isFollowed = info.Follow
				isHidden = info.Hide
				break
			}
		}

		if isFollowed {
			user.Follow = true
			followedUsers = append(followedUsers, user)
		} else if isHidden {
			user.Hide = true
			hiddenUsers = append(hiddenUsers, user)
		} else if !found {
			normalUsers = append(normalUsers, user)
		}
		added[strings.ToLower(user.Username)] = true
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
	matchedUsers = sortedUsernames

	logAction.AppendResult("users_found_with_query", len(matchedUsers))
	return matchedUsers, logging.LogErrorInfo{}
}
