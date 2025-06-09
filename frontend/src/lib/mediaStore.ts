import { create } from "zustand";
import { persist } from "zustand/middleware";

import { MediaItem } from "@/types/mediaItem";

interface MediaStore {
	mediaItem: MediaItem | null;
	setMediaItem: (mediaItem: MediaItem) => void;
	clear: () => void;
	hasHydrated: boolean;
	setHasHydrated: (hydrated: boolean) => void;
}

export const useMediaStore = create<MediaStore>()(
	persist(
		(set) => ({
			mediaItem: null,
			setMediaItem: (mediaItem) => set({ mediaItem }),
			clear: () => set({ mediaItem: null }),
			hasHydrated: false,
			setHasHydrated: (hydrated) => set({ hasHydrated: hydrated }),
		}),
		{
			name: "media-item-storage",
			onRehydrateStorage: () => (state) => {
				state?.setHasHydrated(true);
			},
		}
	)
);
