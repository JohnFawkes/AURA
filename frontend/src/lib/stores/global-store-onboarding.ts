import { fetchOnboardingStatus } from "@/services/settings-onboarding/api-onboarding-fetch-status";
import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import { APIResponse } from "@/types/api/api-response";
import { AppOnboardingStatus } from "@/types/config/onboarding";

interface OnboardingStore {
	status: AppOnboardingStatus | null;
	loading: boolean;
	error: string | null;

	setStatus: (status: AppOnboardingStatus | null) => void;
	setLoading: (loading: boolean) => void;
	setError: (error: string | null) => void;

	fetchStatus: () => Promise<void>;
	clear: () => void;

	hasHydrated: boolean;
	hydrate: () => void;
}

export const useOnboardingStore = create<OnboardingStore>()(
	persist(
		(set) => ({
			status: null,
			loading: false,
			error: null,

			setStatus: (status) => set({ status }),
			setLoading: (loading) => set({ loading }),
			setError: (error) => set({ error }),

			fetchStatus: async () => {
				set({ loading: true, error: null });
				try {
					const response: APIResponse<AppOnboardingStatus> = await fetchOnboardingStatus();
					set({ status: response.data, loading: false });
				} catch (err: unknown) {
					const message = err instanceof Error ? err.message : "Failed to fetch onboarding status";
					set({ error: message, loading: false });
				}
			},

			clear: () => set({ status: null, loading: false, error: null }),

			hasHydrated: false,
			hydrate: () => set({ hasHydrated: true }),
		}),
		{
			name: "Onboarding",
			storage: GlobalStore,
			partialize: (state) => ({
				status: state.status,
				loading: state.loading,
				error: state.error,
			}),
			onRehydrateStorage: () => (state) => {
				state?.hydrate();
			},
		}
	)
);
