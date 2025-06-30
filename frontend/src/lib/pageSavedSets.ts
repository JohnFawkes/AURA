import { create } from "zustand";
import { persist } from "zustand/middleware";

interface SavedSetsPageStore extends SortOptionsStore {
	clear: () => void;
}
export const useSavedSetsPageStore = create<SavedSetsPageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			sortOrder: "asc",
			setSortOption: (option) => set({ sortOption: option }),
			setSortOrder: (order) => set({ sortOrder: order }),

			clear: () => set({ sortOption: "date", sortOrder: "asc" }),
		}),
		{
			name: "saved-sets-page-storage", // key in localStorage
		}
	)
);
