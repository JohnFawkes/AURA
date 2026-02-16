import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import { BaseSetInfo, IncludedItem, SetRef } from "@/types/media-and-posters/sets";
import { TYPE_SET_TYPE_OPTIONS } from "@/types/ui-options";

interface PosterSetsStore {
  setBaseInfo: BaseSetInfo;
  setSetBaseInfo: (setBaseInfo: BaseSetInfo) => void;

  posterSets: SetRef[];
  setPosterSets: (posterSets: SetRef[]) => void;

  includedItems: { [tmdb_id: string]: IncludedItem };
  setIncludedItems: (includedItems: { [tmdb_id: string]: IncludedItem }) => void;

  hasHydrated: boolean;
  hydrate: () => void;
  clear: () => void;
}

export const usePosterSetsStore = create<PosterSetsStore>()(
  persist(
    (set) => ({
      setBaseInfo: {
        id: "",
        title: "",
        type: "set" as TYPE_SET_TYPE_OPTIONS,
        user_created: "",
        date_created: "",
        date_updated: "",
        popularity: 0,
        popularity_global: 0,
      },
      setSetBaseInfo: (setBaseInfo) => set({ setBaseInfo }),

      posterSets: [] as SetRef[],
      setPosterSets: (posterSets) => set({ posterSets }),

      includedItems: {} as { [tmdb_id: string]: IncludedItem },
      setIncludedItems: (includedItems) => set({ includedItems }),

      hasHydrated: false,
      hydrate: () => set({ hasHydrated: true }),

      clear: () =>
        set({
          setBaseInfo: {
            id: "",
            title: "",
            type: "set" as TYPE_SET_TYPE_OPTIONS,
            user_created: "",
            date_created: "",
            date_updated: "",
            popularity: 0,
            popularity_global: 0,
          },
          posterSets: [] as SetRef[],
          includedItems: {} as { [tmdb_id: string]: IncludedItem },
        }),
    }),
    {
      name: "PosterSets",
      storage: GlobalStore,
      partialize: (state) => ({
        setBaseInfo: state.setBaseInfo,
        posterSets: state.posterSets,
        includedItems: state.includedItems,
      }),
      onRehydrateStorage: () => (state) => {
        state?.hydrate();
      },
    }
  )
);
