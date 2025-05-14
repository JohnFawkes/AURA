import { MediaItem } from "@/types/mediaItem";

import { create } from "zustand";
import { persist } from "zustand/middleware";

interface MediaStore {
	mediaItem: MediaItem | null;
	setMediaItem: (mediaItem: MediaItem) => void;
	clear: () => void;
}

export const useMediaStore = create<MediaStore>()(
	persist(
		(set) => ({
			mediaItem: null,
			setMediaItem: (mediaItem) => set({ mediaItem }),
			clear: () => set({ mediaItem: null }),
		}),
		{
			name: "media-item-storage", // key in localStorage
		}
	)
);
