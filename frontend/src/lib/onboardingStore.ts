import { AppOnboardingStatus, fetchOnboardingStatus } from "@/services/api.onboarding";
import { create } from "zustand";
import { persist } from "zustand/middleware";

import { APIResponse } from "@/types/apiResponse";
import { AppConfig } from "@/types/config";

const ONBOARD_TIMEOUT = 30000;

interface OnboardingStore {
	// Raw API data
	data: AppOnboardingStatus | null;

	// Derived flags
	needsSetup: boolean | null; // null = unknown
	checked: boolean;

	// Ephemeral state
	loading: boolean;
	error: APIResponse<unknown> | null;

	lastCheckedAt: number | null;
	staleMs: number;

	// Actions
	check: (force?: boolean) => Promise<void>;
	reset: () => void;

	// Persistence hydration tracking
	hasHydrated: boolean;
	setHasHydrated: (hydrated: boolean) => void;

	complete: (partial?: Partial<AppOnboardingStatus>) => void;
	forceNeedsSetup: () => void;
}

export const useOnboardingStore = create<OnboardingStore>()(
	persist(
		(set, get) => ({
			data: null,
			needsSetup: null,
			checked: false,
			loading: false,
			error: null,
			lastCheckedAt: null,
			staleMs: ONBOARD_TIMEOUT, // 30 seconds

			async check(force = false) {
				const { loading, lastCheckedAt, staleMs } = get();
				if (loading) return;
				const fresh = lastCheckedAt && Date.now() - lastCheckedAt < staleMs;
				if (fresh && !force) return;

				set({ loading: true, error: null });
				try {
					const res = await fetchOnboardingStatus();
					if (res.status === "success") {
						const needs = !!res.data?.needsSetup;
						set({
							data: res.data || null,
							needsSetup: needs,
							checked: true,
							loading: false,
							error: null,
							lastCheckedAt: Date.now(),
						});
					} else {
						set({
							data: null,
							needsSetup: true,
							checked: true,
							loading: false,
							error: res,
							lastCheckedAt: Date.now(),
						});
					}
				} catch {
					set({
						data: null,
						needsSetup: true,
						checked: true,
						loading: false,
						error: null,
						lastCheckedAt: Date.now(),
					});
				}
			},

			reset() {
				set({
					data: null,
					needsSetup: null,
					checked: false,
					loading: false,
					error: null,
					lastCheckedAt: null,
					staleMs: ONBOARD_TIMEOUT,
				});
			},

			hasHydrated: false,
			setHasHydrated: (hydrated) => set({ hasHydrated: hydrated }),

			complete(partial) {
				// Merge with existing data (or create a shell) and enforce completion flags
				const prev = get().data || {
					configLoaded: false,
					configValid: false,
					needsSetup: true,
					currentSetup: {} as AppConfig, // Ensure type is AppConfig
				};
				const merged = {
					...prev,
					...partial,
					configLoaded: true,
					configValid: true,
					needsSetup: false,
					currentSetup: (partial?.currentSetup ?? prev.currentSetup) as AppConfig, // Ensure not null
				};
				set({
					data: merged,
					needsSetup: false,
					checked: true,
				});
			},
			forceNeedsSetup() {
				set({
					needsSetup: true,
					checked: true,
					data: {
						...(get().data || {}),
						needsSetup: true,
					} as AppOnboardingStatus,
				});
			},
		}),
		{
			name: "onboarding-storage",
			// Persist only stable fields
			partialize: (state) => ({
				data: state.data,
				needsSetup: state.needsSetup,
				checked: state.checked,
			}),
			onRehydrateStorage: () => (state) => {
				state?.setHasHydrated(true);
			},
		}
	)
);
