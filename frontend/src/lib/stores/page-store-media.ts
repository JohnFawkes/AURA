import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";
import { TYPE_SORT_ORDER_OPTIONS } from "@/types/ui-options";

type MediaType = "movie" | "show";

interface SortState {
  sortOption: string;
  sortOrder: TYPE_SORT_ORDER_OPTIONS;
}

interface Media_PageStore {
  sortStates: Record<MediaType, SortState>;
  setSortOption: (type: MediaType, option: string) => void;
  setSortOrder: (type: MediaType, order: TYPE_SORT_ORDER_OPTIONS) => void;

  // Filters
  showHiddenUsers: boolean;
  setShowHiddenUsers: (show: boolean) => void;
  showOnlyTitlecardSets: boolean;
  setShowOnlyTitlecardSets: (show: boolean) => void;
  showDefaultImagesOnly: boolean;
  setShowDefaultImagesOnly: (show: boolean) => void;

  // Hydrate and Clear
  hasHydrated: boolean;
  hydrate: () => void;
  clear: () => void;
}

export const useMediaPageStore = create<Media_PageStore>()(
  persist(
    (set) => ({
      sortStates: {
        movie: { sortOption: "date", sortOrder: "desc" },
        show: { sortOption: "date", sortOrder: "desc" },
      },
      setSortOption: (type, option) =>
        set((state) => ({
          sortStates: {
            ...state.sortStates,
            [type]: { ...state.sortStates[type], sortOption: option },
          },
        })),
      setSortOrder: (type, order) =>
        set((state) => ({
          sortStates: {
            ...state.sortStates,
            [type]: { ...state.sortStates[type], sortOrder: order },
          },
        })),

      showHiddenUsers: false,
      setShowHiddenUsers: (show) => set({ showHiddenUsers: show }),

      showOnlyTitlecardSets: false,
      setShowOnlyTitlecardSets: (show) => set({ showOnlyTitlecardSets: show }),

	  showDefaultImagesOnly: false,
	  setShowDefaultImagesOnly: (show) => set({ showDefaultImagesOnly: show }),

      hasHydrated: false,
      hydrate: () => set({ hasHydrated: true }),

      clear: () =>
        set({
          sortStates: {
            movie: { sortOption: "date", sortOrder: "desc" },
            show: { sortOption: "date", sortOrder: "desc" },
          },
          showHiddenUsers: false,
          showOnlyTitlecardSets: false,
		  showDefaultImagesOnly: false,
        }),
    }),
    {
      name: "MediaItem",
      storage: PageStore,
      partialize: (state) => ({
        sortStates: state.sortStates,
        showHiddenUsers: state.showHiddenUsers,
        showOnlyTitlecardSets: state.showOnlyTitlecardSets,
		showDefaultImagesOnly: state.showDefaultImagesOnly,
      }),
      onRehydrateStorage: () => (state) => {
        state?.hydrate();
      },
    }
  )
);