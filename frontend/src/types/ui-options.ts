export type TYPE_SORT_ORDER_OPTIONS = "asc" | "desc";
export const SORT_ORDER_OPTIONS: TYPE_SORT_ORDER_OPTIONS[] = ["asc", "desc"];

export type TYPE_VIEW_TYPE_OPTIONS = "card" | "table";
export const VIEW_TYPE_OPTIONS: TYPE_VIEW_TYPE_OPTIONS[] = ["card", "table"];

export type TYPE_POSTER_SET_TYPE_OPTIONS = "set" | "show" | "movie" | "collection" | "boxset";
export const POSTER_SET_TYPE_OPTIONS: TYPE_POSTER_SET_TYPE_OPTIONS[] = ["set", "show", "movie", "collection", "boxset"];

export type TYPE_ITEMS_PER_PAGE_OPTIONS = 10 | 20 | 30 | 50 | 100;
export const ITEMS_PER_PAGE_OPTIONS: TYPE_ITEMS_PER_PAGE_OPTIONS[] = [10, 20, 30, 50, 100];

export type TYPE_FILTER_IN_DB_OPTIONS = "all" | "notInDB" | "inDB";
export const FILTER_IN_DB_OPTIONS: TYPE_FILTER_IN_DB_OPTIONS[] = ["all", "notInDB", "inDB"];

export type TYPE_DEFAULT_IMAGE_TYPE_OPTIONS =
	| "poster"
	| "backdrop"
	| "seasonPoster"
	| "specialSeasonPoster"
	| "titlecard";
export const DEFAULT_IMAGE_TYPE_OPTIONS: TYPE_DEFAULT_IMAGE_TYPE_OPTIONS[] = [
	"poster",
	"backdrop",
	"seasonPoster",
	"specialSeasonPoster",
	"titlecard",
];
