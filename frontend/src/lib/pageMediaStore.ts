import { create } from "zustand";
import { persist } from "zustand/middleware";

interface MediaPageStore extends SortOptionsStore {
	showHiddenUsers: boolean;
	setShowHiddenUsers: (show: boolean) => void;
	showOnlyTitlecardSets: boolean;
	setShowOnlyTitlecardSets: (show: boolean) => void;
	clear: () => void;
}

export const useMediaPageStore = create<MediaPageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			setSortOption: (option) => set({ sortOption: option }),
			sortOrder: "asc",
			setSortOrder: (order) => set({ sortOrder: order }),
			showHiddenUsers: false,
			setShowHiddenUsers: (show) => set({ showHiddenUsers: show }),
			showOnlyTitlecardSets: false,
			setShowOnlyTitlecardSets: (show) => set({ showOnlyTitlecardSets: show }),

			clear: () => set({ sortOption: "date", sortOrder: "asc", showHiddenUsers: false }),
		}),
		{
			name: "media-page-storage", // key in localStorage
		}
	)
);
