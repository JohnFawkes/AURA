package api

import (
	"aura/internal/logging"
	"context"
)

func Mediux_FetchUserFollowingAndHiding(ctx context.Context) (UserFollowHide, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetch User Following And Hiding", logging.LevelInfo)
	defer logAction.Complete()

	requestBody := Mediux_GenerateUserFollowingAndHidingBody()

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		return UserFollowHide{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxUserFollowHideResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxUserFollowHideResponse")
	if Err.Message != "" {
		return UserFollowHide{}, Err
	}

	data := responseBody.Data
	if data.Follows == nil && data.Hides == nil {
		return UserFollowHide{}, logging.LogErrorInfo{}
	}

	// Create a new UserFollowHide object to store the response data
	userFollowHide := UserFollowHide{
		Follows: make([]MediuxUserFollowHideUserInfo, 0),
		Hides:   make([]MediuxUserFollowHideUserInfo, 0),
	}

	// Populate the UserFollowHide object with the response data
	if data.Follows != nil {
		for _, follow := range data.Follows {
			userFollowHide.Follows = append(userFollowHide.Follows, MediuxUserFollowHideUserInfo{
				ID:       follow.FolloweeID.ID,
				Username: follow.FolloweeID.Username,
			})
		}
	}
	if data.Hides != nil {
		for _, hide := range data.Hides {
			userFollowHide.Hides = append(userFollowHide.Hides, MediuxUserFollowHideUserInfo{
				ID:       hide.HidingID.ID,
				Username: hide.HidingID.Username,
			})
		}
	}

	return userFollowHide, logging.LogErrorInfo{}
}
