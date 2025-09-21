import { create } from "zustand";
import { persist } from "zustand/middleware";

import { log } from "@/lib/logger";
import { GlobalStore } from "@/lib/stores/stores";

import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

type UpdateType = "add" | "update" | "delete";

// Max Cache Duration: 1 Hour
export const MAX_CACHE_DURATION = 60 * 60 * 1000;

interface LibrarySectionsStore {
	sections: Record<string, LibrarySection>;
	setSections: (sections: Record<string, LibrarySection>, timestamp?: number) => void;
	setSection: (sectionTitle: string, record: LibrarySection) => void;

	timestamp?: number;
	setTimestamp: (timestamp: number) => void;

	removeSection: (sectionTitle: string) => void;
	updateMediaItem: (ratingKey: string, sectionTitle: string, updateType: UpdateType) => void;

	getSectionSummaries: () => { title: string; type: string }[];

	clear: () => void;

	hasHydrated: boolean;
	hydrate: () => void;
}

export const useLibrarySectionsStore = create<LibrarySectionsStore>()(
	persist(
		(set, get) => ({
			sections: {},

			setSections: (sections, timestamp) =>
				set({
					sections,
					timestamp: timestamp ?? Date.now(),
				}),

			setSection: (sectionTitle, record) =>
				set((state) => ({
					sections: {
						...state.sections,
						[sectionTitle]: record,
					},
				})),

			timestamp: undefined,
			setTimestamp: (timestamp) => set({ timestamp }),

			removeSection: (sectionTitle) =>
				set((state) => {
					const next = { ...state.sections };
					delete next[sectionTitle];
					return { sections: next };
				}),

			updateMediaItem: (ratingKey, sectionTitle, updateType) => {
				const { sections } = get();
				const section = sections[sectionTitle];
				if (!section || !Array.isArray(section.MediaItems)) {
					log(`librarySections - Section "${sectionTitle}" missing or invalid while updating ${ratingKey}`);
					return;
				}

				const mediaItems = [...section.MediaItems];
				const idx = mediaItems.findIndex((m) => m.RatingKey === ratingKey);
				if (idx === -1) {
					log(`librarySections - Media item ${ratingKey} not found in section "${sectionTitle}"`);
					return;
				}

				if (updateType === "add" || updateType === "update") {
					mediaItems[idx] = { ...mediaItems[idx], ExistInDatabase: true };
				} else if (updateType === "delete") {
					mediaItems[idx] = { ...mediaItems[idx], ExistInDatabase: false };
				}

				set((state) => ({
					sections: {
						...state.sections,
						[sectionTitle]: {
							...section,
							MediaItems: mediaItems,
						},
					},
				}));

				log(
					`librarySections - Updated media item (${updateType}) RatingKey=${ratingKey} in section="${sectionTitle}"`
				);
			},

			getSectionSummaries: () => {
				const { sections } = get();
				return Object.values(sections).map((rec) => ({
					title: rec.Title,
					type: rec.Type,
				}));
			},

			clear: () => {
				set({ sections: {}, timestamp: undefined });
			},

			hasHydrated: false,
			hydrate: () => {
				set({ hasHydrated: true });
			},
		}),
		{
			name: "LibrarySections",
			storage: GlobalStore,
			partialize: (state) => ({
				sections: state.sections,
				timestamp: state.timestamp,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
