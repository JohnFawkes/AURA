package api

type MediuxFollowHideUserInfo struct {
	ID       string           `json:"id"`
	Username string           `json:"username"`
	Avatar   MediuxUserAvatar `json:"avatar"`
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

type MediuxUserInfo struct {
	ID        string `json:"ID"`
	Username  string `json:"Username"`
	Avatar    string `json:"Avatar"`
	Follow    bool   `json:"Follow"`
	Hide      bool   `json:"Hide"`
	TotalSets int    `json:"TotalSets"`
}

type MediuxUserAvatar struct {
	ID string `json:"id"`
}
