export interface SortOptionsStore {
	sortOption: string;
	setSortOption: (option: string) => void;

	sortOrder: "asc" | "desc";
	setSortOrder: (order: "asc" | "desc") => void;
}
