import { PosterSet } from "@/types/posterSets";
import { create } from "zustand";
import { persist } from "zustand/middleware";

interface PosterSetStore {
	posterSet: PosterSet | null;
	setPosterSet: (posterSet: PosterSet) => void;
	clear: () => void;
}
export const usePosterSetStore = create<PosterSetStore>()(
	persist(
		(set) => ({
			posterSet: null,
			setPosterSet: (posterSet) => set({ posterSet }),
			clear: () => set({ posterSet: null }),
		}),
		{
			name: "poster-set-storage", // key in localStorage
		}
	)
);
