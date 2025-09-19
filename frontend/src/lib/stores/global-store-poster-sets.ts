import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import { PosterSet } from "@/types/media-and-posters/poster-sets";
import { TYPE_POSTER_SET_TYPE_OPTIONS } from "@/types/ui-options";

interface PosterSetsStore {
	setType: TYPE_POSTER_SET_TYPE_OPTIONS;
	setSetType: (setType: TYPE_POSTER_SET_TYPE_OPTIONS) => void;

	setTitle: string;
	setSetTitle: (setTitle: string) => void;

	setAuthor: string;
	setSetAuthor: (setAuthor: string) => void;

	setID: string;
	setSetID: (setID: string) => void;

	posterSets: PosterSet[];
	setPosterSets: (posterSets: PosterSet[]) => void;

	_hasHydrated: boolean;
	hasHydrated: () => boolean;
	hydrate: () => void;
	clear: () => void;
}

export const usePosterSetsStore = create<PosterSetsStore>()(
	persist(
		(set, get) => ({
			setType: "set",
			setSetType: (setType) => set({ setType }),

			setTitle: "",
			setSetTitle: (setTitle) => set({ setTitle }),

			setAuthor: "",
			setSetAuthor: (setAuthor) => set({ setAuthor }),

			setID: "",
			setSetID: (setID) => set({ setID }),

			posterSets: [],
			setPosterSets: (posterSets) => set({ posterSets }),

			_hasHydrated: false,
			hasHydrated: () => get()._hasHydrated,
			hydrate: () => set({ _hasHydrated: true }),

			clear: () =>
				set({
					setType: "set",
					setTitle: "",
					setAuthor: "",
					setID: "",
					posterSets: [],
				}),
		}),
		{
			name: "PosterSets",
			storage: GlobalStore,
			partialize: (state) => ({
				setType: state.setType,
				setTitle: state.setTitle,
				setAuthor: state.setAuthor,
				setID: state.setID,
				posterSets: state.posterSets,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
