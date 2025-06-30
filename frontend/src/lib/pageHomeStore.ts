import { create } from "zustand";
import { persist } from "zustand/middleware";

import { SortOptionsStore } from "@/lib/sortOptionsStore";

interface HomePageStore extends SortOptionsStore {
	filteredLibraries: string[];
	setFilteredLibraries: (libraries: string[]) => void;
	filterOutInDB: boolean;
	setFilterOutInDB: (filter: boolean) => void;
	clear: () => void;
}

export const useHomePageStore = create<HomePageStore>()(
	persist(
		(set) => ({
			filteredLibraries: [],
			setFilteredLibraries: (libraries) => set({ filteredLibraries: libraries }),
			filterOutInDB: false,
			setFilterOutInDB: (filter) => set({ filterOutInDB: filter }),
			sortOption: "date",
			sortOrder: "desc",
			setSortOption: (option) => set({ sortOption: option }),
			setSortOrder: (order) => set({ sortOrder: order }),

			clear: () => set({ sortOption: "date", sortOrder: "desc" }),
		}),
		{
			name: "home-page-storage", // key in localStorage
		}
	)
);
