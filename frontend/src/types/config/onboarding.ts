import { AppConfig } from "@/types/config/config-app";

export interface AppOnboardingStatus {
	configLoaded: boolean;
	configValid: boolean;
	needsSetup: boolean;
	currentSetup: AppConfig;
}
