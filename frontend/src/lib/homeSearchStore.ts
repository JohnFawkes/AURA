import { create } from "zustand";
import { persist } from "zustand/middleware";

interface HomeSearchStore {
	searchQuery: string;
	currentPage: number;
	setSearchQuery: (query: string) => void;
	setCurrentPage: (page: number) => void;
	clear: () => void;
}

export const useHomeSearchStore = create<HomeSearchStore>()(
	persist(
		(set) => ({
			searchQuery: "",
			currentPage: 1,
			setSearchQuery: (query) => set({ searchQuery: query }),
			setCurrentPage: (page) => set({ currentPage: page }),
			clear: () => set({ searchQuery: "", currentPage: 1 }),
		}),
		{
			name: "home-search-storage", // key in localStorage
		}
	)
);
