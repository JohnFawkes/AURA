import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import { DOWNLOAD_DEFAULT_TYPE_OPTIONS, TYPE_DOWNLOAD_DEFAULT_OPTIONS } from "@/types/ui-options";

interface UserPreferencesStore {
	downloadDefaults: TYPE_DOWNLOAD_DEFAULT_OPTIONS[];
	setDownloadDefaults: (downloadDefaults: TYPE_DOWNLOAD_DEFAULT_OPTIONS[]) => void;

	showOnlyDownloadDefaults: boolean;
	setShowOnlyDownloadDefaults: (showOnlyDownloadDefaults: boolean) => void;

	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useUserPreferencesStore = create<UserPreferencesStore>()(
	persist(
		(set) => ({
			downloadDefaults: DOWNLOAD_DEFAULT_TYPE_OPTIONS,
			setDownloadDefaults: (downloadDefaults: TYPE_DOWNLOAD_DEFAULT_OPTIONS[]) => set({ downloadDefaults }),

			showOnlyDownloadDefaults: false,
			setShowOnlyDownloadDefaults: (showOnlyDownloadDefaults: boolean) => set({ showOnlyDownloadDefaults }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

			clear: () =>
				set({
					downloadDefaults: DOWNLOAD_DEFAULT_TYPE_OPTIONS,
					showOnlyDownloadDefaults: false,
				}),
		}),
		{
			name: "UserPreferences",
			storage: GlobalStore,
			partialize: (state) => ({
				downloadDefaults: state.downloadDefaults,
				showOnlyDownloadDefaults: state.showOnlyDownloadDefaults,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
