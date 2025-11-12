package api

import (
	"aura/internal/logging"
	"context"
)

func Mediux_FetchUserFollowingAndHiding(ctx context.Context) ([]MediuxUserInfo, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetch User Following And Hiding", logging.LevelInfo)
	defer logAction.Complete()

	requestBody := Mediux_GenerateUserFollowingAndHidingBody()

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		return nil, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxUserFollowHideResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxUserFollowHideResponse")
	if Err.Message != "" {
		return nil, Err
	}

	data := responseBody.Data
	if data.Follows == nil && data.Hides == nil {
		return nil, logging.LogErrorInfo{}
	}

	var userFollowHide []MediuxUserInfo

	// Populate the UserFollowHide object with the response data
	if data.Follows != nil {
		for _, follow := range data.Follows {
			userFollowHide = append(userFollowHide, MediuxUserInfo{
				ID:       follow.FolloweeID.ID,
				Username: follow.FolloweeID.Username,
				Avatar:   follow.FolloweeID.Avatar.ID,
				Follow:   true,
			})
		}
	}
	if data.Hides != nil {
		for _, hide := range data.Hides {
			userFollowHide = append(userFollowHide, MediuxUserInfo{
				ID:       hide.HidingID.ID,
				Username: hide.HidingID.Username,
				Avatar:   hide.HidingID.Avatar.ID,
				Hide:     true,
			})
		}
	}

	return userFollowHide, logging.LogErrorInfo{}
}
