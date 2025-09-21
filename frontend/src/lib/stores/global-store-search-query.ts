import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

interface SearchQueryStore {
	searchQuery: string;
	setSearchQuery: (query: string) => void;

	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useSearchQueryStore = create<SearchQueryStore>()(
	persist(
		(set) => ({
			searchQuery: "",
			setSearchQuery: (query) => set({ searchQuery: query }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

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
