// Image and file-related types
// Define basic types that don't depend on circular references
import type { Show } from "./content";
import type { Season } from "./content";
import type { Episode } from "./content";
import type { Collection } from "./content";
import type { Movie } from "./content";
import type { CollectionSet, MovieSet, ShowSet } from "./sets";
import type { User } from "./user";

// Forward references to avoid circular dependencies
export interface Person {
	id: string; // varchar(255)
	date_created: string | null; // timestamp NULL
	date_updated: string | null; // timestamp NULL
	name: string | null; // varchar(255) NULL
	biography: string | null; // varchar(255) NULL
	birthday: string | null; // date NULL
	deathday: string | null; // date NULL
	known_for: string | null; // varchar(255) NULL
	profile_pic: string | null; // varchar(255) NULL
}

export interface Edition {
	id: number; // int(10) unsigned Auto Increment
	edition_name: string | null; // varchar(255) NULL
}

export interface Language {
	iso_639_1: string; // varchar(255)
	display_name: string | null; // varchar(255) NULL
	iso_639_2: string | null; // varchar(255) NULL
	native_name: string | null; // varchar(255) NULL
}

// Define the possible file types
export const FILE_TYPES = {
	POSTER: "poster",
	BACKDROP: "backdrop",
	TITLECARD: "titlecard",
	ALBUM: "album",
	LOGO: "logo",
} as const;

export type FileType = (typeof FILE_TYPES)[keyof typeof FILE_TYPES] | null;

// Define the AssetFile type according to the provided schema
export interface AssetFile {
	id: string;
	title?: string | null;
	uploaded_by?: User;
	created_on: string;
	modified_by?: User;
	modified_on: string;
	filesize?: number | null;
	type: string;
	width?: number | null;
	height?: number | null;
	uploaded_on?: string | null;
	file_type: FileType;
	language?: Language;
	edition?: Edition;
	show?: Show;
	season?: Season;
	episode?: Episode;
	collection?: Collection;
	movie?: Movie;
	person?: Person;
	collection_set?: CollectionSet;
	movie_set?: MovieSet;
	show_set?: ShowSet;
	filename_disk?: string | null;
	filename_download: string;
	downloads?:
		| {
				id: number;
				date_created: string | null;
				file: AssetFile;
		  }[]
		| null;
	downloads_func?: number | null;
}

// Interface for file uploads
export interface AssetFileUpload {
	// Core file data
	file: File;

	// File metadata
	file_type: FileType;
	title?: string;
	description?: string;

	// Content relations
	show?: string;
	season?: string;
	episode?: string;
	collection?: string;
	movie?: string;
	person?: string;

	// Set relations
	collection_set?: number;
	movie_set?: number;
	show_set?: number;
	set_id?: number;

	// Additional metadata
	language?: string;
	edition?: number;
	category?: string;
}

// Export AssetImage type which is used by the image components
export type AssetImage = AssetFile;

export interface FileCount {
	group: {
		file_type: string;
	};
	count: {
		file_type: number;
	};
}

export interface FileCountsResponse {
	data: {
		files_aggregated: FileCount[];
	};
}

export type FileUploadType = "logo" | "backdrop" | "album" | "titlecard" | "poster";

interface DimensionLimits {
	minWidth: number;
	maxWidth: number;
	minHeight: number;
	maxHeight: number;
	aspectRatio?: {
		min: number;
		max: number;
	};
}

interface FileLimits {
	maxSizeMB: number;
	dimensions: DimensionLimits;
	allowedTypes: string[];
}

export const FILE_LIMITS: Record<FileUploadType, FileLimits> = {
	poster: {
		maxSizeMB: 10,
		dimensions: {
			minWidth: 1000,
			maxWidth: 2000,
			minHeight: 1500,
			maxHeight: 3000,
			aspectRatio: {
				min: 0.6666666666666666, // Exact 2:3
				max: 0.6666666666666666, // Exact 2:3
			},
		},
		allowedTypes: [".png", ".jpg", ".jpeg", ".webp"],
	},
	backdrop: {
		maxSizeMB: 10,
		dimensions: {
			minWidth: 1920,
			maxWidth: 3840,
			minHeight: 1080,
			maxHeight: 2160,
			aspectRatio: {
				min: 1.7777777777777777, // Exact 16:9
				max: 1.7777777777777777, // Exact 16:9
			},
		},
		allowedTypes: [".png", ".jpg", ".jpeg", ".webp"],
	},
	titlecard: {
		maxSizeMB: 3,
		dimensions: {
			minWidth: 1280,
			maxWidth: 2560,
			minHeight: 720,
			maxHeight: 1440,
			aspectRatio: {
				min: 1.7777777777777777, // Exact 16:9
				max: 1.7777777777777777, // Exact 16:9
			},
		},
		allowedTypes: [".png", ".jpg", ".jpeg", ".webp"],
	},
	album: {
		maxSizeMB: 3,
		dimensions: {
			minWidth: 720,
			maxWidth: 3000,
			minHeight: 720,
			maxHeight: 3000,
			aspectRatio: {
				min: 1, // Exact 1:1
				max: 1, // Exact 1:1
			},
		},
		allowedTypes: [".png", ".jpg", ".jpeg", ".webp"],
	},
	logo: {
		maxSizeMB: 5,
		dimensions: {
			minWidth: 0,
			maxWidth: Infinity,
			minHeight: 0,
			maxHeight: Infinity,
		},
		allowedTypes: [".png", ".svg"],
	},
};

export const ASPECT_RATIOS: Record<FileUploadType, string> = {
	poster: "aspect-[2/3]",
	backdrop: "aspect-video",
	titlecard: "aspect-video",
	album: "aspect-square",
	logo: "", // No aspect ratio requirement for logos
};

export const FILE_TYPE_ERRORS: Record<FileUploadType, string> = {
	poster: "Only PNG, JPG, and WEBP files are allowed with 2:3 aspect ratio (1000x1500px - 2000x3000px)",
	backdrop:
		"Only PNG, JPG, and WEBP files are allowed with 16:9 aspect ratio (1920x1080px - 3840x2160px)",
	titlecard:
		"Only PNG, JPG, and WEBP files are allowed with 16:9 aspect ratio (1280x720px - 2560x1440px)",
	album: "Only PNG, JPG, and WEBP files are allowed with 1:1 aspect ratio (720x720px - 3000x3000px)",
	logo: "Only SVG and PNG files are allowed (up to 5MB, any dimensions)",
};

// Helper function to convert MB to bytes
export const mbToBytes = (mb: number) => mb * 1024 * 1024;
