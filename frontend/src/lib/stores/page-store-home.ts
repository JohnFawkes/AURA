import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PaginationStore, SortStore } from "@/types/store-interfaces";
import { TYPE_ITEMS_PER_PAGE_OPTIONS, TYPE_SORT_ORDER_OPTIONS } from "@/types/ui-options";
import { TYPE_FILTER_IN_DB_OPTIONS } from "@/types/ui-options";

interface Home_PageStore
	extends SortStore<string, TYPE_SORT_ORDER_OPTIONS>,
		PaginationStore<number, TYPE_ITEMS_PER_PAGE_OPTIONS> {
	// Filters
	filteredAndSortedMediaItems: MediaItem[];
	setFilteredAndSortedMediaItems: (items: MediaItem[]) => void;
	filteredLibraries: string[];
	setFilteredLibraries: (libraries: string[]) => void;
	filterInDB: TYPE_FILTER_IN_DB_OPTIONS;
	setFilterInDB: (filter: TYPE_FILTER_IN_DB_OPTIONS) => void;

	// Hydrate and Clear
	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useHomePageStore = create<Home_PageStore>()(
	persist(
		(set) => ({
			sortOption: "dateUpdated",
			setSortOption: (option) => set({ sortOption: option }),

			sortOrder: "asc",
			setSortOrder: (order) => set({ sortOrder: order }),

			currentPage: 1,
			setCurrentPage: (page) => set({ currentPage: page }),

			itemsPerPage: 20,
			setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),

			filteredAndSortedMediaItems: [],
			setFilteredAndSortedMediaItems: (items) => set({ filteredAndSortedMediaItems: items }),

			filteredLibraries: [],
			setFilteredLibraries: (libraries) => set({ filteredLibraries: libraries }),

			filterInDB: "all",
			setFilterInDB: (filter) => set({ filterInDB: filter }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

			clear: () =>
				set({
					sortOption: "dateUpdated",
					sortOrder: "asc",
					currentPage: 1,
					itemsPerPage: 20,
					filteredAndSortedMediaItems: [],
					filteredLibraries: [],
					filterInDB: "all",
				}),
		}),
		{
			name: "Home",
			storage: PageStore,
			partialize: (state) => ({
				sortOption: state.sortOption,
				sortOrder: state.sortOrder,
				currentPage: state.currentPage,
				itemsPerPage: state.itemsPerPage,
				filteredAndSortedMediaItems: state.filteredAndSortedMediaItems,
				filteredLibraries: state.filteredLibraries,
				filterInDB: state.filterInDB,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
