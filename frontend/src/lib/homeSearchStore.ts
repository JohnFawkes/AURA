import { create } from "zustand";
import { persist } from "zustand/middleware";

interface HomeSearchStore {
	searchQuery: string;
	currentPage: number;
	itemsPerPage: number;
	filteredLibraries: string[];
	filterOutInDB: boolean;
	setSearchQuery: (query: string) => void;
	setCurrentPage: (page: number) => void;
	setItemsPerPage: (itemsPerPage: number) => void;
	setFilteredLibraries: (libraries: string[]) => void;
	setFilterOutInDB: (filter: boolean) => void;
	clear: () => void;
}

export const useHomeSearchStore = create<HomeSearchStore>()(
	persist(
		(set) => ({
			searchQuery: "",
			currentPage: 1,
			itemsPerPage: 20, // Default items per page
			filteredLibraries: [],
			filterOutInDB: false,
			setSearchQuery: (query) => set({ searchQuery: query }),
			setCurrentPage: (page) => set({ currentPage: page }),
			setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),
			setFilteredLibraries: (libraries) => set({ filteredLibraries: libraries }),
			setFilterOutInDB: (filter) => set({ filterOutInDB: filter }),
			clear: () => set({ searchQuery: "", currentPage: 1 }),
		}),
		{
			name: "home-search-storage", // key in localStorage
		}
	)
);
