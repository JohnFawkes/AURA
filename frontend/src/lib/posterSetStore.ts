import { create } from "zustand";
import { persist } from "zustand/middleware";

import { PosterSet } from "@/types/posterSets";

interface PosterSetsStore {
	setType: "set" | "show" | "movie" | "collection" | "boxset";
	setTitle: string;
	setAuthor: string;
	setID: string;
	posterSets: PosterSet[];
	setPosterSets: (posterSets: PosterSet[]) => void;
	setSetType: (setType: "show" | "movie" | "collection" | "boxset") => void;
	setSetTitle: (setTitle: string) => void;
	setSetAuthor: (setAuthor: string) => void;
	setSetID: (setID: string) => void;
	clear: () => void;
}
export const usePosterSetsStore = create<PosterSetsStore>()(
	persist(
		(sets) => ({
			setType: "set",
			setTitle: "",
			setAuthor: "",
			setID: "",
			setSetType: (setType) => sets({ setType }),
			setSetTitle: (setTitle) => sets({ setTitle }),
			setSetAuthor: (setAuthor) => sets({ setAuthor }),
			setSetID: (setID) => sets({ setID }),
			posterSets: [],
			setPosterSets: (posterSets) => sets({ posterSets: posterSets }),
			clear: () => sets({ posterSets: [] }),
		}),
		{
			name: "poster-sets-storage", // key in localStorage
		}
	)
);
