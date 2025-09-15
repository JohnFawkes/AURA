import { create } from "zustand";
import { persist } from "zustand/middleware";

import { SortOptionsStore } from "@/lib/sortOptionsStore";
import { ViewOptionsStore } from "@/lib/viewOptions";

interface SavedSetsPageStore extends SortOptionsStore, ViewOptionsStore {
	clear: () => void;
}
export const useSavedSetsPageStore = create<SavedSetsPageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			sortOrder: "asc",
			viewOption: "card",
			setSortOption: (option) => set({ sortOption: option }),
			setSortOrder: (order) => set({ sortOrder: order }),
			setViewOption: (option) => set({ viewOption: option }),

			clear: () => set({ sortOption: "date", sortOrder: "asc", viewOption: "card" }),
		}),
		{
			name: "saved-sets-page-storage", // key in localStorage
		}
	)
);
