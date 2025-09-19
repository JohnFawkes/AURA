import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

interface SearchQueryStore {
	searchQuery: string;
	setSearchQuery: (query: string) => void;

	_hasHydrated: boolean;
	hasHydrated: () => boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useSearchQueryStore = create<SearchQueryStore>()(
	persist(
		(set, get) => ({
			searchQuery: "",
			setSearchQuery: (query) => set({ searchQuery: query }),

			_hasHydrated: false,
			hasHydrated: () => get()._hasHydrated,
			hydrate: () => set({ _hasHydrated: true }),

			clear: () => set({ searchQuery: "" }),
		}),
		{
			name: "SearchQuery",
			storage: GlobalStore,
			partialize: (state) => ({
				searchQuery: state.searchQuery,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
