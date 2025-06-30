import { create } from "zustand";
import { persist } from "zustand/middleware";

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
