import { create } from "zustand";
import { persist } from "zustand/middleware";

interface PageStore {
	currentPage: number;
	setCurrentPage: (page: number) => void;
	itemsPerPage: number;
	setItemsPerPage: (itemsPerPage: number) => void;
}

export const usePageStore = create<PageStore>()(
	persist(
		(set) => ({
			currentPage: 1,
			setCurrentPage: (page) => set({ currentPage: page }),
			itemsPerPage: 20, // Default items per page
			setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),
		}),
		{
			name: "page-storage", // key in localStorage
		}
	)
);

interface SearchQueryStore {
	searchQuery: string;
	setSearchQuery: (query: string) => void;
}

export const useSearchQueryStore = create<SearchQueryStore>()(
	persist(
		(set) => ({
			searchQuery: "",
			setSearchQuery: (query) => set({ searchQuery: query }),
		}),
		{
			name: "search-query-storage", // key in localStorage
		}
	)
);

interface SortOptionsStore {
	sortOption: string;
	setSortOption: (option: string) => void;

	sortOrder: "asc" | "desc";
	setSortOrder: (order: "asc" | "desc") => void;
}

interface MediaPageStore extends SortOptionsStore {
	showHiddenUsers: boolean;
	setShowHiddenUsers: (show: boolean) => void;
	showOnlyTitlecardSets: boolean;
	setShowOnlyTitlecardSets: (show: boolean) => void;
	clear: () => void;
}

export const useMediaPageStore = create<MediaPageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			setSortOption: (option) => set({ sortOption: option }),
			sortOrder: "asc",
			setSortOrder: (order) => set({ sortOrder: order }),
			showHiddenUsers: false,
			setShowHiddenUsers: (show) => set({ showHiddenUsers: show }),
			showOnlyTitlecardSets: false,
			setShowOnlyTitlecardSets: (show) => set({ showOnlyTitlecardSets: show }),

			clear: () => set({ sortOption: "date", sortOrder: "asc", showHiddenUsers: false }),
		}),
		{
			name: "media-page-storage", // key in localStorage
		}
	)
);

interface SavedSetsPageStore extends SortOptionsStore {
	clear: () => void;
}
export const useSavedSetsPageStore = create<SavedSetsPageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			sortOrder: "asc",
			setSortOption: (option) => set({ sortOption: option }),
			setSortOrder: (order) => set({ sortOrder: order }),

			clear: () => set({ sortOption: "date", sortOrder: "asc" }),
		}),
		{
			name: "saved-sets-page-storage", // key in localStorage
		}
	)
);

interface HomePageStore extends SortOptionsStore {
	filteredLibraries: string[];
	setFilteredLibraries: (libraries: string[]) => void;
	filterOutInDB: boolean;
	setFilterOutInDB: (filter: boolean) => void;
	clear: () => void;
}

export const useHomePageStore = create<HomePageStore>()(
	persist(
		(set) => ({
			filteredLibraries: [],
			setFilteredLibraries: (libraries) => set({ filteredLibraries: libraries }),
			filterOutInDB: false,
			setFilterOutInDB: (filter) => set({ filterOutInDB: filter }),
			sortOption: "date",
			sortOrder: "asc",
			setSortOption: (option) => set({ sortOption: option }),
			setSortOrder: (order) => set({ sortOrder: order }),

			clear: () => set({ sortOption: "date", sortOrder: "asc" }),
		}),
		{
			name: "home-page-storage", // key in localStorage
		}
	)
);
