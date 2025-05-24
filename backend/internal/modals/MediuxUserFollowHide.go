package modals

type MediuxFollowHideUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type MediuxUserFollow struct {
	FolloweeID MediuxFollowHideUserInfo `json:"followee_id"`
}

type MediuxUserHide struct {
	HidingID MediuxFollowHideUserInfo `json:"hiding_id"`
}

type MediuxUserFollowHideResponse struct {
	Data struct {
		Follows []MediuxUserFollow `json:"user_follows,omitempty"`
		Hides   []MediuxUserHide   `json:"user_hides,omitempty"`
	} `json:"data"`
}

type MediuxUserFollowHideUserInfo struct {
	ID       string `json:"ID"`
	Username string `json:"Username"`
}

type UserFollowHide struct {
	Follows []MediuxUserFollowHideUserInfo `json:"Follows"`
	Hides   []MediuxUserFollowHideUserInfo `json:"Hides"`
}
