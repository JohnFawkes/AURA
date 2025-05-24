export interface MediuxUserFollowHide {
	Follows?: MediuxFollowHideUserInfo[];
	Hides?: MediuxFollowHideUserInfo[];
}

export interface MediuxFollowHideUserInfo {
	ID: string;
	Username: string;
}
