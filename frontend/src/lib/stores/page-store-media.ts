import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";

import { SortStore } from "@/types/store-interfaces";
import { TYPE_SORT_ORDER_OPTIONS } from "@/types/ui-options";

interface Media_PageStore extends SortStore<string, TYPE_SORT_ORDER_OPTIONS> {
	// Filters
	showHiddenUsers: boolean;
	setShowHiddenUsers: (show: boolean) => void;
	showOnlyTitlecardSets: boolean;
	setShowOnlyTitlecardSets: (show: boolean) => void;

	// Hydrate and Clear
	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useMediaPageStore = create<Media_PageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			setSortOption: (option) => set({ sortOption: option }),

			sortOrder: "desc",
			setSortOrder: (order) => set({ sortOrder: order }),

			showHiddenUsers: false,
			setShowHiddenUsers: (show) => set({ showHiddenUsers: show }),

			showOnlyTitlecardSets: false,
			setShowOnlyTitlecardSets: (show) => set({ showOnlyTitlecardSets: show }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

			clear: () =>
				set({
					sortOption: "date",
					sortOrder: "desc",
					showHiddenUsers: false,
					showOnlyTitlecardSets: false,
				}),
		}),
		{
			name: "MediaItem",
			storage: PageStore,
			partialize: (state) => ({
				sortOption: state.sortOption,
				sortOrder: state.sortOrder,
				showHiddenUsers: state.showHiddenUsers,
				showOnlyTitlecardSets: state.showOnlyTitlecardSets,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
