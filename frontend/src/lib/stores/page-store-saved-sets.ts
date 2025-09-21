import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";

import { PaginationStore, SortStore } from "@/types/store-interfaces";
import { TYPE_ITEMS_PER_PAGE_OPTIONS, TYPE_SORT_ORDER_OPTIONS, TYPE_VIEW_TYPE_OPTIONS } from "@/types/ui-options";

interface SavedSets_PageStore
	extends SortStore<string, TYPE_SORT_ORDER_OPTIONS>,
		PaginationStore<number, TYPE_ITEMS_PER_PAGE_OPTIONS> {
	// View Store
	viewOption: TYPE_VIEW_TYPE_OPTIONS;
	setViewOption: (option: TYPE_VIEW_TYPE_OPTIONS) => void;

	// Library Filter
	filteredLibraries: string[];
	setFilteredLibraries: (libraries: string[]) => void;

	// AutoDownload Filter
	filterAutoDownloadOnly: boolean;
	setFilterAutoDownloadOnly: (value: boolean) => void;

	// User Filter
	filteredUsers: string[];
	setFilteredUsers: (users: string[]) => void;

	// Selected Type Filter
	filteredTypes: string[];
	setFilteredTypes: (types: string[]) => void;

	// MultiSet Filter
	filterMultiSetOnly: boolean;
	setFilterMultiSetOnly: (value: boolean) => void;

	// Hydration and Clear
	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useSavedSetsPageStore = create<SavedSets_PageStore>()(
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

			viewOption: "card",
			setViewOption: (option) => set({ viewOption: option }),

			filteredLibraries: [],
			setFilteredLibraries: (libraries) => set({ filteredLibraries: libraries }),

			filterAutoDownloadOnly: false,
			setFilterAutoDownloadOnly: (value) => set({ filterAutoDownloadOnly: value }),

			filteredUsers: [],
			setFilteredUsers: (users) => set({ filteredUsers: users }),

			filteredTypes: [],
			setFilteredTypes: (types) => set({ filteredTypes: types }),

			filterMultiSetOnly: false,
			setFilterMultiSetOnly: (value) => set({ filterMultiSetOnly: value }),

			// Hydration and Clear
			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

			clear: () =>
				set({
					sortOption: "dateUpdated",
					sortOrder: "asc",
					currentPage: 1,
					itemsPerPage: 20,
					viewOption: "card",
					filteredLibraries: [],
					filterAutoDownloadOnly: false,
					filteredUsers: [],
					filteredTypes: [],
					filterMultiSetOnly: false,
					hasHydrated: false,
				}),
		}),
		{
			name: "SavedSets",
			storage: PageStore,
			partialize: (state) => ({
				sortOption: state.sortOption,
				sortOrder: state.sortOrder,
				currentPage: state.currentPage,
				itemsPerPage: state.itemsPerPage,
				viewOption: state.viewOption,
				filteredLibraries: state.filteredLibraries,
				filterAutoDownloadOnly: state.filterAutoDownloadOnly,
				filteredUsers: state.filteredUsers,
				filteredTypes: state.filteredTypes,
				filterMultiSetOnly: state.filterMultiSetOnly,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
