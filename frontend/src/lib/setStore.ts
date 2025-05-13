import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";
import { create } from "zustand";
import { persist } from "zustand/middleware";

// Zustand store type
interface PosterMediaStore {
	posterSet: PosterSet | null;
	mediaItem: MediaItem | null;
	setPosterSet: (posterSet: PosterSet) => void;
	setMediaItem: (mediaItem: MediaItem) => void;
	clear: () => void;
}

export const usePosterMediaStore = create<PosterMediaStore>()(
	persist(
		(set) => ({
			posterSet: null,
			mediaItem: null,
			setPosterSet: (posterSet) => set({ posterSet }),
			setMediaItem: (mediaItem) => set({ mediaItem }),
			clear: () => set({ posterSet: null, mediaItem: null }),
		}),
		{
			name: "poster-media-storage", // key in localStorage
		}
	)
);
