"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchConfig } from "@/services/settings-onboarding/api-config-fetch";
import { updateConfig } from "@/services/settings-onboarding/api-config-update";
import { toast } from "sonner";

import { useEffect, useRef, useState } from "react";

import { useRouter } from "next/navigation";

import { ConfigSectionAuth } from "@/components/settings-onboarding/ConfigSectionAuth";
import { ConfigSectionAutoDownload } from "@/components/settings-onboarding/ConfigSectionAutoDownload";
import { ConfigSectionImages } from "@/components/settings-onboarding/ConfigSectionImages";
import { ConfigSectionLabelsAndTags } from "@/components/settings-onboarding/ConfigSectionLabelsAndTags";
import { ConfigSectionLogging } from "@/components/settings-onboarding/ConfigSectionLogging";
import { ConfigSectionMediaServer } from "@/components/settings-onboarding/ConfigSectionMediaServer";
import { ConfigSectionMediux } from "@/components/settings-onboarding/ConfigSectionMediux";
import { ConfigSectionNotifications } from "@/components/settings-onboarding/ConfigSectionNotifications";
import { ConfigSectionTMDB } from "@/components/settings-onboarding/ConfigSectionTMDB";
import { UserPreferencesCard } from "@/components/settings-onboarding/UserPreferences";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { Button } from "@/components/ui/button";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

import { ClearAllStores } from "@/lib/stores/clear-all-stores";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config-app";
import { defaultAppConfig } from "@/types/config/config-default-app";

type ObjectSectionKeys = {
	[K in keyof AppConfig]-?: NonNullable<AppConfig[K]> extends object ? K : never;
}[keyof AppConfig];

type SectionDirty<S extends ObjectSectionKeys = ObjectSectionKeys> = Partial<
	Record<keyof NonNullable<AppConfig[S]>, boolean>
>;

interface ImagesDirty {
	CacheImages?: { Enabled?: boolean };
	SaveImageNextToContent?: { Enabled?: boolean };
}

type DirtyState = {
	[K in ObjectSectionKeys]?: K extends "Images" ? ImagesDirty : SectionDirty<K>;
};

type ValidationErrors = {
	[K in ObjectSectionKeys]?: Record<string, string>; // field -> message
};

const SettingsPage: React.FC = () => {
	const router = useRouter();
	const isMounted = useRef(false);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	const [editing, setEditing] = useState(false);
	const [saving, setSaving] = useState(false);

	const [initialConfig, setInitialConfig] = useState<AppConfig>(() => defaultAppConfig());
	const [newConfig, setNewConfig] = useState<AppConfig>(() => defaultAppConfig());
	const [dirty, setDirty] = useState<DirtyState>({});

	const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});

	const preferencesRef = useRef<HTMLDivElement>(null);

	// State - Debug Mode
	const [debugEnabled, setDebugEnabled] = useState(false);

	useEffect(() => {
		const savedMode = localStorage.getItem("debugMode") === "true";
		setDebugEnabled(savedMode);
	}, []);

	useEffect(() => {
		const scrollToPreferences = () => {
			if (loading) return;
			if (window.location.hash === "#preferences-section") {
				preferencesRef.current?.scrollIntoView({ behavior: "smooth" });
			}
		};
		window.addEventListener("hashchange", scrollToPreferences);
		// Run on mount
		scrollToPreferences();
		return () => window.removeEventListener("hashchange", scrollToPreferences);
	}, [loading]);

	const toggleDebugMode = (checked: boolean) => {
		setDebugEnabled(checked);
		localStorage.setItem("debugMode", checked.toString());
	};

	// Fetch configuration data
	useEffect(() => {
		if (typeof window !== "undefined") {
			document.title = "aura | Settings";
		}
		if (isMounted.current) return;
		isMounted.current = true;

		const fetchConfigFromAPI = async () => {
			try {
				setLoading(true);
				const response = await fetchConfig();
				if (response.status === "error") {
					setError(response);
					setInitialConfig(defaultAppConfig());
					setNewConfig(defaultAppConfig());
					return;
				}

				const cfg = response.data ?? defaultAppConfig();
				setInitialConfig(cfg);
				setNewConfig(cfg);
				setError(null);
			} catch (error) {
				setError(ReturnErrorMessage<AppConfig>(error));
				setInitialConfig(defaultAppConfig());
				setNewConfig(defaultAppConfig());
			} finally {
				setLoading(false);
			}
		};

		fetchConfigFromAPI();
	}, []);

	const handleCancel = () => {
		setEditing(false);
		setSaving(false);
		setDirty({});
		setNewConfig(initialConfig); // <- reset edits
	};

	const handleSaveAll = async () => {
		if (!newConfig) return;

		setSaving(true);

		setEditing(false);

		try {
			const resp = await updateConfig(newConfig);

			if (resp.status === "error") {
				setError(resp);
			} else if (resp.status === "warn") {
				toast.warning("No changes detected.");
			} else {
				toast.success("Configuration updated successfully.");
				window.location.reload();
			}
		} catch (error) {
			toast.error(
				typeof error === "object" && error !== null && "message" in error
					? (error as { message?: string }).message || "Unknown error"
					: "Unknown error"
			);
		}
		setSaving(false);
		setDirty({});
	};

	const structuralEqual = (a: unknown, b: unknown): boolean => {
		if (a === b) return true;
		if (Array.isArray(a) && Array.isArray(b)) {
			if (a.length !== b.length) return false;
			for (let i = 0; i < a.length; i++) {
				if (!structuralEqual(a[i], b[i])) return false;
			}
			return true;
		}
		if (a && b && typeof a === "object" && typeof b === "object") {
			const keysA = Object.keys(a as object);
			const keysB = Object.keys(b as object);
			if (keysA.length !== keysB.length) return false;
			for (const k of keysA) {
				if (!structuralEqual((a as Record<string, unknown>)[k], (b as Record<string, unknown>)[k]))
					return false;
			}
			return true;
		}
		return false;
	};

	const updateConfigField = <S extends ObjectSectionKeys, F extends keyof NonNullable<AppConfig[S]>>(
		section: S,
		field: F,
		value: NonNullable<AppConfig[S]>[F]
	) => {
		setNewConfig((prev) => {
			const prevSection = (prev[section] ?? {}) as NonNullable<AppConfig[S]>;
			return {
				...prev,
				[section]: {
					...prevSection,
					[field]: value,
				},
			};
		});

		setDirty((prev) => {
			const originalValue = initialConfig[section]?.[field];

			// Choose comparison strategy
			let reverted: boolean;
			if (section === "MediaServer" && field === "Libraries") {
				reverted = structuralEqual(originalValue, value); // (order-sensitive alternative)
			} else {
				reverted = structuralEqual(originalValue, value);
			}

			const prevSectionDirty = (prev[section] ?? {}) as SectionDirty<S>;
			let nextSectionDirty = prevSectionDirty;

			if (reverted) {
				if (prevSectionDirty[field]) {
					const clone = { ...prevSectionDirty };
					delete clone[field];
					nextSectionDirty = clone;
				}
			} else if (!prevSectionDirty[field]) {
				nextSectionDirty = { ...prevSectionDirty, [field]: true };
			}

			const nextState: DirtyState = { ...prev };
			if (Object.keys(nextSectionDirty).length === 0) {
				delete nextState[section];
			} else {
				nextState[section] = nextSectionDirty;
			}
			return nextState;
		});
	};

	const updateImagesField = <G extends keyof AppConfig["Images"], F extends keyof AppConfig["Images"][G]>(
		group: G,
		field: F,
		value: AppConfig["Images"][G][F]
	) => {
		setNewConfig((prev) => {
			const nextGroup = {
				...prev.Images[group],
				[field]: value,
			};
			return {
				...prev,
				Images: {
					...prev.Images,
					[group]: nextGroup,
				},
			};
		});

		setDirty((prev) => {
			const originalVal = initialConfig.Images[group][field];
			const reverted = originalVal === value;

			const prevImagesDirty = (prev.Images ?? {}) as ImagesDirty;
			const prevGroupDirty = (prevImagesDirty[group] ?? {}) as { [k in F]?: boolean };

			let nextGroupDirty = prevGroupDirty;

			if (reverted) {
				if (prevGroupDirty[field]) {
					const clone = { ...prevGroupDirty };
					delete clone[field];
					nextGroupDirty = clone;
				}
			} else if (!prevGroupDirty[field]) {
				nextGroupDirty = { ...prevGroupDirty, [field]: true };
			}

			const nextImagesDirty: ImagesDirty = { ...prevImagesDirty };
			if (Object.keys(nextGroupDirty).length === 0) {
				delete nextImagesDirty[group];
			} else {
				nextImagesDirty[group] = nextGroupDirty;
			}

			const nextState: DirtyState = { ...prev };
			if (Object.keys(nextImagesDirty).length === 0) {
				delete nextState.Images;
			} else {
				nextState.Images = nextImagesDirty;
			}
			return nextState;
		});
	};

	const anyDirty =
		Object.values(dirty).some((section) => section && Object.values(section).some(Boolean)) ||
		JSON.stringify(initialConfig) !== JSON.stringify(newConfig);

	const updateSectionErrors = <S extends ObjectSectionKeys>(
		section: S,
		errs: Record<string, string> | Partial<Record<string, string>>
	) => {
		setValidationErrors((prev) => {
			if (!errs || Object.keys(errs).length === 0) {
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
				const { [section]: _, ...rest } = prev;
				return rest;
			}
			return { ...prev, [section]: errs as Record<string, string> };
		});
	};

	const hasValidationErrors = Object.keys(validationErrors).length > 0;

	return (
		<div className="container mx-auto p-6">
			{loading ? (
				<Loader message="Loading configuration..." />
			) : error ? (
				<ErrorMessage error={error} />
			) : (
				<>
					<div className="flex items-center justify-between mb-4">
						<div>
							<h1 className="text-3xl font-bold">Settings</h1>
							<p className="text-gray-600 dark:text-gray-400">Manage your application settings</p>
						</div>

						<div className="flex gap-2">
							{!editing && (
								<Button
									variant="outline"
									onClick={() => setEditing(true)}
									className="cursor-pointer hover:text-primary active:scale-95 hover:brightness-120"
								>
									Edit
								</Button>
							)}
							{editing && (
								<>
									<Button
										variant="outline"
										onClick={handleCancel}
										disabled={saving}
										className="cursor-pointer hover:text-primary"
									>
										Cancel
									</Button>
									<Button
										onClick={handleSaveAll}
										disabled={!anyDirty || saving || hasValidationErrors}
										className="cursor-pointer hover:text-primary"
									>
										{saving ? "Saving..." : "Save All"}
									</Button>
								</>
							)}
						</div>
					</div>

					{editing && hasValidationErrors && (
						<p className="mb-2 text-red-500">Fix validation errors before saving.</p>
					)}

					<div className="space-y-5">
						<ConfigSectionAuth
							value={newConfig.Auth}
							editing={editing}
							dirtyFields={dirty.Auth}
							onChange={(field, value) => updateConfigField("Auth", field, value)}
							errorsUpdate={(errs) => updateSectionErrors("Auth", errs as Record<string, string>)}
						/>

						<ConfigSectionLogging
							value={newConfig.Logging}
							editing={editing}
							dirtyFields={dirty.Logging}
							onChange={(field, value) => updateConfigField("Logging", field, value)}
							errorsUpdate={(errs) => updateSectionErrors("Logging", errs as Record<string, string>)}
						/>

						<ConfigSectionMediaServer
							value={newConfig.MediaServer}
							editing={editing}
							configAlreadyLoaded={false}
							dirtyFields={dirty.MediaServer}
							onChange={(field, value) => updateConfigField("MediaServer", field, value)}
							errorsUpdate={(errs) => updateSectionErrors("MediaServer", errs as Record<string, string>)}
						/>

						<ConfigSectionMediux
							value={newConfig.Mediux}
							editing={editing}
							configAlreadyLoaded={false}
							dirtyFields={dirty.Mediux}
							onChange={(field, value) => updateConfigField("Mediux", field, value)}
							errorsUpdate={(errs) => updateSectionErrors("Mediux", errs as Record<string, string>)}
						/>

						<ConfigSectionAutoDownload
							value={newConfig.AutoDownload}
							editing={editing}
							dirtyFields={dirty.AutoDownload}
							onChange={(f, v) => updateConfigField("AutoDownload", f, v)}
							errorsUpdate={(errs) => updateSectionErrors("AutoDownload", errs as Record<string, string>)}
						/>

						<ConfigSectionImages
							value={newConfig.Images}
							editing={editing}
							dirtyFields={dirty.Images}
							onChange={updateImagesField}
							errorsUpdate={(errs) => updateSectionErrors("Images", errs as Record<string, string>)}
						/>

						<ConfigSectionTMDB
							value={newConfig.TMDB}
							editing={editing}
							dirtyFields={dirty.TMDB}
							onChange={(f, v) => updateConfigField("TMDB", f, v)}
							errorsUpdate={(errs) => updateSectionErrors("TMDB", errs as Record<string, string>)}
						/>

						<ConfigSectionLabelsAndTags
							value={newConfig.LabelsAndTags}
							editing={editing}
							dirtyFields={
								dirty.LabelsAndTags as {
									Applications?: Array<
										Partial<
											Record<
												string,
												boolean | { Enabled?: boolean; Add?: boolean; Remove?: boolean }
											>
										>
									>;
								}
							}
							onChange={(field, val) => updateConfigField("LabelsAndTags", field, val)}
							errorsUpdate={(errs) =>
								updateSectionErrors("LabelsAndTags", errs as Record<string, string>)
							}
						/>

						<ConfigSectionNotifications
							value={newConfig.Notifications}
							editing={editing}
							dirtyFields={
								dirty.Notifications as {
									Enabled?: boolean;
									Providers?: Partial<
										Record<
											string,
											| boolean
											| {
													Enabled?: boolean;
													Webhook?: boolean;
													UserKey?: boolean;
													Token?: boolean;
											  }
										>
									>[];
								}
							}
							onChange={(field, val) => updateConfigField("Notifications", field, val)}
							errorsUpdate={(errs) =>
								updateSectionErrors("Notifications", errs as Record<string, string>)
							}
						/>
					</div>

					{editing && (
						<div className="sticky bottom-0 mt-10 z-30">
							<div className="mx-auto w-fit bg-background/90 backdrop-blur border rounded-md shadow px-4 py-3 flex items-center gap-3">
								<span className="text-sm">{anyDirty ? "Unsaved changes" : "No changes yet"}</span>
								<Button size="sm" variant="outline" onClick={handleCancel} disabled={saving}>
									Cancel
								</Button>
								<Button onClick={handleSaveAll} disabled={!anyDirty || saving || hasValidationErrors}>
									{saving ? "Saving..." : "Save All"}
								</Button>
							</div>
						</div>
					)}

					{/* User Preferences Section */}
					<div id="preferences-section" ref={preferencesRef}>
						<UserPreferencesCard />
					</div>
				</>
			)}

			{/* Debug Mode Toggle */}
			<div className="flex items-center justify-between mt-6 border-t pt-4">
				<ToggleGroup
					type="single"
					variant={debugEnabled ? "default" : "outline"}
					value={debugEnabled ? "enabled" : "disabled"}
					onValueChange={(value) => toggleDebugMode(value === "enabled")}
				>
					<ToggleGroupItem value="enabled" variant={debugEnabled ? "default" : "outline"}>
						<span className="flex items-center gap-2 cursor-pointer">
							Debug Mode:
							{debugEnabled ? (
								<span className="text-green-500">Enabled</span>
							) : (
								<span className="text-destructive">Disabled</span>
							)}
						</span>
					</ToggleGroupItem>
				</ToggleGroup>

				<Button
					variant="ghost"
					className="text-red-600 border-none shadow-none hover:bg-red-50 cursor-pointer"
					onClick={async () => {
						localStorage.clear();
						await ClearAllStores();
						toast.success("App Cache Cleared. Reloading...", { duration: 750 });
						setTimeout(() => {
							router.replace("/settings");
						}, 1000);
					}}
				>
					Clear App Cache
				</Button>
			</div>
		</div>
	);
};

export default SettingsPage;
