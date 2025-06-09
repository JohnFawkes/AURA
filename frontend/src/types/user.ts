// User-related types
import type { AssetFile } from "./files";

// Homepage preferences types
export const SECTION_IDS = {
	TRENDING_MIXED: "trending-mixed",
	TRENDING_MOVIES: "trending-movies",
	TRENDING_SHOWS: "trending-shows",
	FILE_COUNTER: "file-counter",
} as const;

export type SectionId = (typeof SECTION_IDS)[keyof typeof SECTION_IDS];

export interface HomepageSection {
	id: SectionId;
	visible: boolean;
}

export type HomepagePreferences = HomepageSection[];

// Default preferences configuration
export const DEFAULT_HOMEPAGE_PREFERENCES: HomepagePreferences = [
	{ id: SECTION_IDS.TRENDING_MIXED, visible: true },
	// { id: SECTION_IDS.TRENDING_MOVIES, visible: true },
	// { id: SECTION_IDS.TRENDING_SHOWS, visible: true },
	{ id: SECTION_IDS.FILE_COUNTER, visible: true },
];

export interface User {
	id: string;
	username: string;
	email: string;
	joined: string;
	avatar?: AssetFile;
	role: string;
	last_login: string | null;
	status: string;
	title?: string;
	token?: string;
	last_access?: string;
	discord?: string;
	instagram?: string;
	reddit?: string;
	paypal?: string;
	buymeacoffee?: string;
	patreon?: string;
	kofi?: string;
	trakt?: string;
	tagline?: string;
	about?: string;
	backdrop?: AssetFile;
	simkl?: string;
	plex?: string;
	letterboxd?: string;
	homepage_preferences?: HomepagePreferences;
	following?: UserFollow[];
	followers?: UserFollow[];
	file_counts?: FileCounts;
}

export interface UserFollow {
	id: number;
	date_created: string | null;
	follower: User;
	followee: User;
}

// Define a type for the simplified followee information needed in the session
export interface FolloweeInfo {
	id: string;
	username: string;
}

// Define a type for the simplified hidden user information needed in the session
export interface HiddenUserInfo {
	id: string;
	username: string;
}

export type SessionInfo = {
	token: string;
	expiresAt: number; // Timestamp in milliseconds
	userData?: {
		id?: string;
		username?: string;
		email?: string;
		avatar?: string;
		backdrop?: string;
		tfa?: boolean; // Indicates if the user has 2FA enabled
		homepage_preferences?: HomepagePreferences;
		following?: FolloweeInfo[]; // Updated following type
		hiding?: HiddenUserInfo[]; // Added hiding type
	};
} | null;

export type RefreshResponse = {
	success: boolean;
	token?: string;
	expiresAt?: number;
	message?: string;
	userData?: {
		id?: string;
		username?: string;
		email?: string;
		avatar?: string;
		backdrop?: string;
		tfa?: boolean; // Indicates if the user has 2FA enabled
		homepage_preferences?: HomepagePreferences;
		following?: FolloweeInfo[]; // Updated following type
		hiding?: HiddenUserInfo[]; // Added hiding type
	};
};

// File count aggregation types
export interface FileCountGroup {
	file_type: string;
}

export interface FileCount {
	file_type: number;
}

export interface FileCountAggregation {
	group: FileCountGroup;
	count: FileCount;
}

export interface FileCountsResponse {
	data: {
		files_aggregated: FileCountAggregation[];
	};
}

export interface FileCounts {
	album?: number;
	backdrop?: number;
	logo?: number;
	poster?: number;
	titlecard?: number;
}

// Set count aggregation types
export interface SetCount {
	id: number;
}

export interface SetCountAggregation {
	count: SetCount;
}

export interface SetCountsResponse {
	data: {
		movie_sets_aggregated: SetCountAggregation[];
		show_sets_aggregated: SetCountAggregation[];
		collection_sets_aggregated: SetCountAggregation[];
	};
}

export interface SetCounts {
	movies: number;
	shows: number;
	collections: number;
}

// User Activity Aggregation Types
export interface UserActivityGroup {
	uploaded_on_month: number;
	uploaded_on_year: number;
	file_type: string;
}

export interface UserActivityCount {
	id: number; // Represents the upload count
}

export interface UserActivityAggregation {
	group: UserActivityGroup;
	count: UserActivityCount;
}

export interface UserActivityData {
	files_aggregated: UserActivityAggregation[];
}

export interface UserActivityResponse {
	data: UserActivityData;
	errors?: { message: string; [key: string]: unknown }[]; // Use unknown instead of any
}

// User Download Records Aggregation Types
export interface UserDownloadRecordGroup {
	date_created_month: number;
	date_created_year: number;
}

export interface UserDownloadRecordCount {
	id: number; // Represents the download count for the period
}

export interface UserDownloadRecordAggregation {
	group: UserDownloadRecordGroup;
	count: UserDownloadRecordCount;
}

export interface UserDownloadRecordsResponse {
	data: {
		download_records_aggregated: UserDownloadRecordAggregation[];
	};
	errors?: { message: string; [key: string]: unknown }[];
}
