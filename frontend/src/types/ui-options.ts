export type TYPE_SORT_ORDER_OPTIONS = "asc" | "desc";
export const SORT_ORDER_OPTIONS: TYPE_SORT_ORDER_OPTIONS[] = ["asc", "desc"];

export type TYPE_VIEW_TYPE_OPTIONS = "card" | "table";
export const VIEW_TYPE_OPTIONS: TYPE_VIEW_TYPE_OPTIONS[] = ["card", "table"];

export type TYPE_SET_TYPE_OPTIONS = "show" | "movie" | "collection" | "boxset";
export const SET_TYPE_OPTIONS: TYPE_SET_TYPE_OPTIONS[] = ["show", "movie", "collection", "boxset"];

export type TYPE_DB_SET_TYPE_OPTIONS = "show" | "movie" | "collection";
export const DB_SET_TYPE_OPTIONS: TYPE_DB_SET_TYPE_OPTIONS[] = ["show", "movie", "collection"];

export type TYPE_ITEMS_PER_PAGE_OPTIONS = 10 | 20 | 30 | 50 | 100;
export const ITEMS_PER_PAGE_OPTIONS: TYPE_ITEMS_PER_PAGE_OPTIONS[] = [10, 20, 30, 50, 100];

export type TYPE_FILTER_IN_DB_OPTIONS = "all" | "notInDB" | "inDB";
export const FILTER_IN_DB_OPTIONS: TYPE_FILTER_IN_DB_OPTIONS[] = ["all", "notInDB", "inDB"];

export type TYPE_FILTER_IGNORED_OPTIONS = "none" | "always" | "temp";
export const FILTER_IGNORED_OPTIONS: TYPE_FILTER_IGNORED_OPTIONS[] = ["none", "always", "temp"];

export type TYPE_DOWNLOAD_DEFAULT_OPTIONS =
    | "poster"
    | "backdrop"
    | "season_poster"
    | "special_season_poster"
    | "titlecard";
export const DOWNLOAD_DEFAULT_TYPE_OPTIONS: TYPE_DOWNLOAD_DEFAULT_OPTIONS[] = [
    "poster",
    "backdrop",
    "season_poster",
    "special_season_poster",
    "titlecard",
];

export const DOWNLOAD_DEFAULT_LABELS: Record<TYPE_DOWNLOAD_DEFAULT_OPTIONS, string> = {
    poster: "Poster",
    backdrop: "Backdrop",
    season_poster: "Season Posters",
    special_season_poster: "Special Season Posters",
    titlecard: "Titlecards",
};
