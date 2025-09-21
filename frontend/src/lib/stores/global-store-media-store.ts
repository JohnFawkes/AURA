import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

interface MediaStore {
	mediaItem: MediaItem | null;
	setMediaItem: (mediaItem: MediaItem | null) => void;

	clear: () => void;

	hasHydrated: boolean;
	hydrate: () => void;
}

export const useMediaStore = create<MediaStore>()(
	persist(
		(set) => ({
			mediaItem: null,
			setMediaItem: (mediaItem) => set({ mediaItem }),

			clear: () => set({ mediaItem: null }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),
		}),
		{
			name: "CurrentMedia",
			storage: GlobalStore,
			partialize: (state) => ({
				mediaItem: state.mediaItem,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
