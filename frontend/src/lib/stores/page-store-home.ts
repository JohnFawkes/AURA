import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PaginationStore, SortStore } from "@/types/store-interfaces";
import { TYPE_ITEMS_PER_PAGE_OPTIONS, TYPE_SORT_ORDER_OPTIONS } from "@/types/ui-options";
import { TYPE_FILTER_IN_DB_OPTIONS } from "@/types/ui-options";

type Direction = "next" | "previous";

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

	getAdjacentMediaItem: (currentRatingKey: string, direction: Direction) => MediaItem | null;

	// Adjacent Items
	previousMediaItem: MediaItem | null;
	setPreviousMediaItem: (mediaItem: MediaItem | null) => void;
	nextMediaItem: MediaItem | null;
	setNextMediaItem: (mediaItem: MediaItem | null) => void;

	// Hydrate and Clear
	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useHomePageStore = create<Home_PageStore>()(
	persist(
		(set, get) => ({
			sortOption: "dateAdded",
			setSortOption: (option) => set({ sortOption: option }),

			sortOrder: "desc",
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

			/**
			 * Retrieves adjacent media item (wrap-around) from the Home page store's
			 * filteredAndSortedMediaItems array.
			 */
			getAdjacentMediaItem: (currentRatingKey: string, direction: Direction): MediaItem | null => {
				const mediaItems = get().filteredAndSortedMediaItems || [];
				if (!mediaItems.length) return null;

				const currentIndex = mediaItems.findIndex((m) => m.RatingKey === currentRatingKey);
				if (currentIndex === -1) return null;

				const nextIndex =
					direction === "next"
						? (currentIndex + 1) % mediaItems.length
						: (currentIndex - 1 + mediaItems.length) % mediaItems.length;

				return mediaItems[nextIndex] ?? null;
			},

			previousMediaItem: null,
			setPreviousMediaItem: (mediaItem) => set({ previousMediaItem: mediaItem }),

			nextMediaItem: null,
			setNextMediaItem: (mediaItem) => set({ nextMediaItem: mediaItem }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

			clear: () =>
				set({
					sortOption: "dateAdded",
					sortOrder: "desc",
					currentPage: 1,
					itemsPerPage: 20,
					filteredAndSortedMediaItems: [],
					filteredLibraries: [],
					filterInDB: "all",
					previousMediaItem: null,
					nextMediaItem: null,
					hasHydrated: false,
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
				previousMediaItem: state.previousMediaItem,
				nextMediaItem: state.nextMediaItem,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
