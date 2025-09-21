import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import { DEFAULT_IMAGE_TYPE_OPTIONS, TYPE_DEFAULT_IMAGE_TYPE_OPTIONS } from "@/types/ui-options";

interface UserPreferencesStore {
	defaultImageTypes: TYPE_DEFAULT_IMAGE_TYPE_OPTIONS[];
	setDefaultImageTypes: (defaultImageTypes: TYPE_DEFAULT_IMAGE_TYPE_OPTIONS[]) => void;

	hasHydrated: boolean;
	hydrate: () => void;
	clear: () => void;
}

export const useUserPreferencesStore = create<UserPreferencesStore>()(
	persist(
		(set) => ({
			defaultImageTypes: DEFAULT_IMAGE_TYPE_OPTIONS,
			setDefaultImageTypes: (defaultImageTypes: TYPE_DEFAULT_IMAGE_TYPE_OPTIONS[]) => set({ defaultImageTypes }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),

			clear: () =>
				set({
					defaultImageTypes: DEFAULT_IMAGE_TYPE_OPTIONS,
				}),
		}),
		{
			name: "UserPreferences",
			storage: GlobalStore,
			partialize: (state) => ({
				defaultImageTypes: state.defaultImageTypes,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
