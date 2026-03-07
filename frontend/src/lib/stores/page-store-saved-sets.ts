import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PageStore } from "@/lib/stores/stores";

import type { PaginationStore, SortStore } from "@/types/store-interfaces";
import type {
  TYPE_FILTER_AUTO_DOWNLOAD_OPTIONS,
  TYPE_ITEMS_PER_PAGE_OPTIONS,
  TYPE_SAVED_SET_VIEW_TYPE_OPTIONS,
  TYPE_SORT_ORDER_OPTIONS,
} from "@/types/ui-options";

interface SavedSets_PageStore
  extends SortStore<string, TYPE_SORT_ORDER_OPTIONS>, PaginationStore<number, TYPE_ITEMS_PER_PAGE_OPTIONS> {
  // View Store
  viewOption: TYPE_SAVED_SET_VIEW_TYPE_OPTIONS;
  setViewOption: (option: TYPE_SAVED_SET_VIEW_TYPE_OPTIONS) => void;

  // Library Filter
  filteredLibraries: string[];
  setFilteredLibraries: (libraries: string[]) => void;

  // AutoDownload Filter
  filterAutoDownload: TYPE_FILTER_AUTO_DOWNLOAD_OPTIONS;
  setFilterAutoDownload: (val: TYPE_FILTER_AUTO_DOWNLOAD_OPTIONS) => void;

  // User Filter
  filteredUsers: string[];
  setFilteredUsers: (users: string[]) => void;

  // Selected Type Filter
  filteredTypes: string[];
  setFilteredTypes: (types: string[]) => void;

  // MultiSet Filter
  filterMultiSetOnly: boolean;
  setFilterMultiSetOnly: (value: boolean) => void;

  // Hydration and Clear
  hasHydrated: boolean;
  hydrate: () => void;
  clear: () => void;
}

export const useSavedSetsPageStore = create<SavedSets_PageStore>()(
  persist(
    (set) => ({
      sortOption: "date_downloaded",
      setSortOption: (option) => set({ sortOption: option }),

      sortOrder: "desc",
      setSortOrder: (order) => set({ sortOrder: order }),

      currentPage: 1,
      setCurrentPage: (page) => set({ currentPage: page }),

      itemsPerPage: 20,
      setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),

      viewOption: "card",
      setViewOption: (option) => set({ viewOption: option }),

      filteredLibraries: [],
      setFilteredLibraries: (libraries) => set({ filteredLibraries: libraries }),

      filterAutoDownload: "",
      setFilterAutoDownload: (value) => set({ filterAutoDownload: value }),

      filteredUsers: [],
      setFilteredUsers: (users) => set({ filteredUsers: users }),

      filteredTypes: [],
      setFilteredTypes: (types) => set({ filteredTypes: types }),

      filterMultiSetOnly: false,
      setFilterMultiSetOnly: (value) => set({ filterMultiSetOnly: value }),

      // Hydration and Clear
      hasHydrated: false,
      hydrate: () => set({ hasHydrated: true }),

      clear: () =>
        set({
          sortOption: "date_downloaded",
          sortOrder: "desc",
          currentPage: 1,
          itemsPerPage: 20,
          viewOption: "card",
          filteredLibraries: [],
          filterAutoDownload: "",
          filteredUsers: [],
          filteredTypes: [],
          filterMultiSetOnly: false,
          hasHydrated: false,
        }),
    }),
    {
      name: "SavedSets",
      storage: PageStore,
      partialize: (state) => ({
        sortOption: state.sortOption,
        sortOrder: state.sortOrder,
        currentPage: state.currentPage,
        itemsPerPage: state.itemsPerPage,
        viewOption: state.viewOption,
        filteredLibraries: state.filteredLibraries,
        filterAutoDownload: state.filterAutoDownload,
        filteredUsers: state.filteredUsers,
        filteredTypes: state.filteredTypes,
        filterMultiSetOnly: state.filterMultiSetOnly,
      }),
      onRehydrateStorage: () => (state) => {
        state?.hydrate();
      },
    }
  )
);
