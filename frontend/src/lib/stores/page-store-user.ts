import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";

import { PaginationStore, SortStore } from "@/types/store-interfaces";
import { TYPE_ITEMS_PER_PAGE_OPTIONS, TYPE_SORT_ORDER_OPTIONS } from "@/types/ui-options";

interface User_PageStore
	extends SortStore<string, TYPE_SORT_ORDER_OPTIONS>,
		PaginationStore<number, TYPE_ITEMS_PER_PAGE_OPTIONS> {
	// Hydration and Clear
	_hasHydrated: boolean;
	hasHydrated: () => boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useUserPageStore = create<User_PageStore>()(
	persist(
		(set, get) => ({
			sortOption: "date",
			setSortOption: (option) => set({ sortOption: option }),

			sortOrder: "desc",
			setSortOrder: (order) => set({ sortOrder: order }),

			currentPage: 1,
			setCurrentPage: (page) => set({ currentPage: page }),

			itemsPerPage: 20,
			setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),

			_hasHydrated: false,
			hasHydrated: () => get()._hasHydrated,
			hydrate: () => set({ _hasHydrated: true }),

			clear: () =>
				set({
					sortOption: "date",
					sortOrder: "desc",
					currentPage: 1,
					itemsPerPage: 20,
				}),
		}),
		{
			name: "UserPage",
			storage: PageStore,
			partialize: (state) => ({
				sortOption: state.sortOption,
				sortOrder: state.sortOrder,
				currentPage: state.currentPage,
				itemsPerPage: state.itemsPerPage,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
