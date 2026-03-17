package models

type MediuxUserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Follow    bool   `json:"follow"`
	Hide      bool   `json:"hide"`
	TotalSets int    `json:"total_sets"`
}
