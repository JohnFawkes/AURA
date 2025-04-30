import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";
import { create } from "zustand";

// Zustand store type
interface PosterMediaStore {
	posterSet: PosterSet | null;
	mediaItem: MediaItem | null;
	setPosterSet: (posterSet: PosterSet) => void;
	setMediaItem: (mediaItem: MediaItem) => void;
	clear: () => void;
}

export const usePosterMediaStore = create<PosterMediaStore>((set) => ({
	posterSet: null,
	mediaItem: null,
	setPosterSet: (posterSet) => set({ posterSet }),
	setMediaItem: (mediaItem) => set({ mediaItem }),
	clear: () => set({ posterSet: null, mediaItem: null }),
}));
