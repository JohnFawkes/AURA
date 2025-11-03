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
import { ConfigSectionSonarrRadarr } from "@/components/settings-onboarding/ConfigSectionSonarrRadarr";
import { UserPreferencesCard } from "@/components/settings-onboarding/UserPreferences";
import { ConfirmDestructiveDialogActionButton } from "@/components/shared/dialog-destructive-action";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

import { ClearAllStores } from "@/lib/stores/clear-all-stores";

import { APIResponse } from "@/types/api/api-response";
import {
	AppConfig,
	AppConfigNotificationDiscord,
	AppConfigNotificationGotify,
	AppConfigNotificationPushover,
	AppConfigNotificationWebhook,
} from "@/types/config/config-app";
import { defaultAppConfig } from "@/types/config/config-default-app";

type ObjectSectionKeys = {
	[K in keyof AppConfig]-?: NonNullable<AppConfig[K]> extends object ? K : never;
}[keyof AppConfig];

type SectionDirty<S extends ObjectSectionKeys = ObjectSectionKeys> = S extends "SonarrRadarr"
	? Partial<{ Type: boolean; Library: boolean; URL: boolean; APIKey: boolean }>
	: Partial<Record<keyof NonNullable<AppConfig[S]>, boolean>>;

interface ImagesDirty {
	CacheImages?: { Enabled?: boolean };
	SaveImagesLocally?: { Enabled?: boolean; Path?: string };
}

type NotificationsDirty = {
	Enabled?: boolean;
	Providers?: Array<
		Partial<
			Record<
				string,
				| boolean
				| {
						Enabled?: boolean;
						Webhook?: boolean;
						UserKey?: boolean;
						Token?: boolean;
						URL?: boolean;
						Headers?: Record<string, boolean>;
				  }
			>
		>
	>;
};

type DirtyState = {
	[K in ObjectSectionKeys]?: K extends "Images"
		? ImagesDirty
		: K extends "SonarrRadarr"
			? Array<SectionDirty<"SonarrRadarr">>
			: K extends "Notifications"
				? NotificationsDirty
				: SectionDirty<K>;
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

	const [activeTab, setActiveTab] = useState("app-settings");

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

	const fetchAndSetConfig = async (reload: boolean = false) => {
		try {
			setLoading(true);
			const response = await fetchConfig(reload);
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

	// Fetch configuration data on mount
	useEffect(() => {
		if (typeof window !== "undefined") {
			document.title = "aura | Settings";
		}
		if (isMounted.current) return;
		isMounted.current = true;

		fetchAndSetConfig();
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

	const updateConfigField = <S extends ObjectSectionKeys, F extends keyof AppConfig[S]>(
		section: S,
		field: F,
		value: AppConfig[S][F]
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
			// --- SonarrRadarr dirty tracking ---
			if (section === "SonarrRadarr" && field === "Applications") {
				const originalApps = initialConfig.SonarrRadarr.Applications ?? [];
				const newApps = (value as AppConfig["SonarrRadarr"]["Applications"]) ?? [];
				const dirtyArr = newApps.map((app, idx) => {
					const orig = originalApps[idx];
					const dirtyObj: SectionDirty<"SonarrRadarr"> = {};
					for (const key of ["Type", "Library", "URL", "APIKey"] as const) {
						if (app[key] !== orig?.[key]) dirtyObj[key] = true;
					}
					return dirtyObj;
				});
				return { ...prev, SonarrRadarr: dirtyArr };
			}

			// --- Notifications dirty tracking ---
			if (section === "Notifications" && field === "Providers") {
				const originalProviders = initialConfig.Notifications.Providers;
				const newProviders = value as AppConfig["Notifications"]["Providers"];
				const dirtyProviders = newProviders.map((prov, idx) => {
					const orig = originalProviders[idx];
					const dirtyObj: Partial<
						Record<
							string,
							{
								Enabled?: boolean;
								Webhook?: boolean;
								UserKey?: boolean;
								Token?: boolean;
								URL?: boolean;
								Headers?: Record<string, boolean>;
							}
						>
					> = {};

					for (const key of ["Discord", "Pushover", "Gotify", "Webhook"] as const) {
						if (prov[key] && orig?.[key]) {
							const fieldDirty: {
								Enabled?: boolean;
								Webhook?: boolean;
								UserKey?: boolean;
								URL?: boolean;
								Token?: boolean;
								Headers?: Record<string, boolean>;
							} = {};
							if (key === "Discord") {
								const discordKeys = ["Enabled", "Webhook"] as const;
								for (const subKey of discordKeys) {
									if (
										(prov[key] as AppConfigNotificationDiscord)[subKey] !==
										(orig[key] as AppConfigNotificationDiscord)[subKey]
									) {
										fieldDirty[subKey] = true;
									}
								}
							} else if (key === "Pushover") {
								const pushoverKeys = ["Enabled", "UserKey", "Token"] as const;
								for (const subKey of pushoverKeys) {
									if (
										(prov[key] as AppConfigNotificationPushover)[subKey] !==
										(orig[key] as AppConfigNotificationPushover)[subKey]
									) {
										fieldDirty[subKey] = true;
									}
								}
							} else if (key === "Gotify") {
								const gotifyKeys = ["Enabled", "Token", "URL"] as const;
								for (const subKey of gotifyKeys) {
									if (
										(prov[key] as AppConfigNotificationGotify)[subKey] !==
										(orig[key] as AppConfigNotificationGotify)[subKey]
									) {
										fieldDirty[subKey] = true;
									}
								}
							} else if (key === "Webhook") {
								const webhookKeys = ["Enabled", "URL"] as const;
								for (const subKey of webhookKeys) {
									if (
										(prov[key] as AppConfigNotificationWebhook)[subKey] !==
										(orig[key] as AppConfigNotificationWebhook)[subKey]
									) {
										fieldDirty[subKey] = true;
									}
								}
								// Deep compare Headers
								const newHeaders = (prov[key] as AppConfigNotificationWebhook).Headers ?? {};
								const origHeaders = (orig[key] as AppConfigNotificationWebhook).Headers ?? {};
								const headersDirty: Record<string, boolean> = {};
								const allHeaderKeys = Array.from(
									new Set([...Object.keys(newHeaders), ...Object.keys(origHeaders)])
								);
								for (const hKey of allHeaderKeys) {
									if (newHeaders[hKey] !== origHeaders[hKey]) {
										headersDirty[hKey] = true;
									}
								}
								fieldDirty.Headers = Object.keys(headersDirty).length > 0 ? headersDirty : undefined;
							}
							if (Object.keys(fieldDirty).length > 0) {
								dirtyObj[key] = fieldDirty;
							}
						}
					}
					return dirtyObj;
				});
				return {
					...prev,
					Notifications: {
						...prev.Notifications,
						Providers: dirtyProviders,
					},
				};
			}

			const originalValue = initialConfig[section]?.[field];
			const reverted: boolean = structuralEqual(originalValue, value);

			const prevSectionDirty = (prev[section] ?? {}) as SectionDirty<S>;
			let nextSectionDirty = prevSectionDirty;

			// Only run for non-SonarrRadarr sections
			if (section !== "SonarrRadarr") {
				// Explicit cast via unknown to satisfy TypeScript
				const dirtyKey = field as unknown as keyof SectionDirty<S>;
				if (reverted) {
					if (prevSectionDirty[dirtyKey]) {
						const clone = { ...prevSectionDirty };
						delete clone[dirtyKey];
						nextSectionDirty = clone;
					}
				} else if (!prevSectionDirty[dirtyKey]) {
					nextSectionDirty = { ...prevSectionDirty, [dirtyKey]: true };
				}

				const nextState: DirtyState = { ...prev };
				if (Object.keys(nextSectionDirty).length === 0) {
					delete nextState[section];
				} else {
					nextState[section] = nextSectionDirty as DirtyState[S];
				}
				return nextState;
			}

			return prev;
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
					{/* If on Dev version (show reload config button) */}
					{process.env.NEXT_PUBLIC_APP_VERSION && process.env.NEXT_PUBLIC_APP_VERSION.endsWith("dev") && (
						<div className="flex justify-center md:justify-end">
							<Button
								variant="ghost"
								onClick={() => fetchAndSetConfig(true)}
								disabled={loading}
								className="cursor-pointer bg-green-500/10 hover:text-primary active:scale-95 hover:brightness-120 mb-4"
							>
								Reload Configuration
							</Button>
						</div>
					)}
					<div className="flex items-center justify-between mb-4">
						<div>
							<h1 className="text-3xl font-bold">
								{activeTab === "user-preferences" ? "User Preferences" : "Settings"}
							</h1>
							<p className="text-gray-600 dark:text-gray-400">
								{activeTab === "user-preferences"
									? "Manage your user preferences"
									: "Manage your application settings"}
							</p>
						</div>
						{activeTab !== "user-preferences" && (
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
						)}
					</div>

					<Tabs defaultValue="app-settings" value={activeTab} onValueChange={setActiveTab} className="w-full">
						<TabsList className="rounded-md p-1 w-full flex">
							<TabsTrigger
								value="app-settings"
								className="flex-1 cursor-pointer text-primary data-[state=active]:bg-primary data-[state=active]:text-background dark:data-[state=active]:bg-primary dark:data-[state=active]:text-background hover:brightness-120 active:scale-95"
							>
								App Settings
							</TabsTrigger>
							<TabsTrigger
								value="user-preferences"
								className="flex-1 cursor-pointer text-primary data-[state=active]:bg-primary data-[state=active]:text-background dark:data-[state=active]:bg-primary dark:data-[state=active]:text-background hover:brightness-120 active:scale-95"
							>
								User Preferences
							</TabsTrigger>
						</TabsList>

						<TabsContent value="app-settings" className="mt-6 w-full">
							<div className="space-y-5 w-full">
								<ConfigSectionMediux
									value={newConfig.Mediux}
									editing={editing}
									configAlreadyLoaded={true}
									dirtyFields={dirty.Mediux}
									onChange={(field, value) => updateConfigField("Mediux", field, value)}
									errorsUpdate={(errs) =>
										updateSectionErrors("Mediux", errs as Record<string, string>)
									}
								/>
								<ConfigSectionMediaServer
									value={newConfig.MediaServer}
									editing={editing}
									configAlreadyLoaded={true}
									dirtyFields={dirty.MediaServer}
									onChange={(field, value) => updateConfigField("MediaServer", field, value)}
									errorsUpdate={(errs) =>
										updateSectionErrors("MediaServer", errs as Record<string, string>)
									}
								/>
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
									errorsUpdate={(errs) =>
										updateSectionErrors("Logging", errs as Record<string, string>)
									}
								/>
								<ConfigSectionImages
									value={newConfig.Images}
									editing={editing}
									dirtyFields={
										dirty.Images
											? {
													...dirty.Images,
													SaveImagesLocally: dirty.Images.SaveImagesLocally
														? {
																...dirty.Images.SaveImagesLocally,
																Path:
																	typeof dirty.Images.SaveImagesLocally.Path ===
																	"string"
																		? !!dirty.Images.SaveImagesLocally.Path
																		: dirty.Images.SaveImagesLocally.Path,
															}
														: undefined,
												}
											: undefined
									}
									onChange={updateImagesField}
									errorsUpdate={(errs) =>
										updateSectionErrors("Images", errs as Record<string, string>)
									}
									mediaServerType={newConfig.MediaServer.Type}
								/>
								<ConfigSectionAutoDownload
									value={newConfig.AutoDownload}
									editing={editing}
									dirtyFields={dirty.AutoDownload}
									onChange={(f, v) => updateConfigField("AutoDownload", f, v)}
									errorsUpdate={(errs) =>
										updateSectionErrors("AutoDownload", errs as Record<string, string>)
									}
								/>

								{/* <ConfigSectionTMDB
									value={newConfig.TMDB}
									editing={editing}
									dirtyFields={dirty.TMDB}
									onChange={(f, v) => updateConfigField("TMDB", f, v)}
									errorsUpdate={(errs) => updateSectionErrors("TMDB", errs as Record<string, string>)}
								/> */}

								<ConfigSectionSonarrRadarr
									value={newConfig.SonarrRadarr}
									editing={editing}
									dirtyFields={
										dirty.SonarrRadarr
											? {
													Applications: dirty.SonarrRadarr as Partial<{
														Type: boolean;
														Library: boolean;
														URL: boolean;
														APIKey: boolean;
													}>[],
												}
											: undefined
									}
									onChange={(field, val) => updateConfigField("SonarrRadarr", field, val)}
									errorsUpdate={(errs) =>
										updateSectionErrors("SonarrRadarr", errs as Record<string, string>)
									}
									configAlreadyLoaded={true}
									libraries={newConfig.MediaServer.Libraries || []}
								/>

								{(newConfig.MediaServer.Type === "Plex" ||
									(Array.isArray(newConfig.SonarrRadarr.Applications) &&
										newConfig.SonarrRadarr.Applications.length > 0)) && (
									<ConfigSectionLabelsAndTags
										value={newConfig.LabelsAndTags}
										editing={editing}
										dirtyFields={
											dirty.LabelsAndTags as {
												Applications?: Array<
													Partial<
														Record<
															string,
															| boolean
															| { Enabled?: boolean; Add?: boolean; Remove?: boolean }
														>
													>
												>;
											}
										}
										mediaServerType={newConfig.MediaServer.Type}
										srOptions={Array.from(
											new Set(
												(newConfig.SonarrRadarr.Applications ?? [])
													.map((app) => app.Type)
													.filter((type) => !!type)
											)
										)}
										onChange={(field, val) => updateConfigField("LabelsAndTags", field, val)}
										errorsUpdate={(errs) =>
											updateSectionErrors("LabelsAndTags", errs as Record<string, string>)
										}
									/>
								)}

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
															URL?: boolean;
															Headers?: Record<string, boolean>;
													  }
												>
											>[];
										}
									}
									onChange={(field, val) => updateConfigField("Notifications", field, val)}
									errorsUpdate={(errs) =>
										updateSectionErrors("Notifications", errs as Record<string, string>)
									}
									configAlreadyLoaded={true}
								/>
							</div>
						</TabsContent>

						<TabsContent value="user-preferences" className="mt-6 w-full">
							<div id="preferences-section" ref={preferencesRef} className="w-full">
								<UserPreferencesCard />
							</div>
						</TabsContent>
					</Tabs>

					{activeTab === "app-settings" && editing && hasValidationErrors && (
						<p className="mb-2 text-red-500">Fix validation errors before saving.</p>
					)}

					{activeTab === "app-settings" && editing && (
						<div className="sticky bottom-0 mt-10 z-30">
							<div
								className={`mx-auto w-fit bg-background/90 backdrop-blur border rounded-md shadow px-4 py-3 flex items-center gap-3 ${anyDirty && "border-amber-500"}`}
							>
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
				</>
			)}

			{/* Debug Mode Toggle & Cache Clear */}
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
								<span className="text-destructive hover:text-red-500">Disabled</span>
							)}
						</span>
					</ToggleGroupItem>
				</ToggleGroup>

				<ConfirmDestructiveDialogActionButton
					onConfirm={async () => {
						localStorage.clear();
						await ClearAllStores();
						toast.success("App Cache Cleared. Reloading...", { duration: 750 });
						setTimeout(() => {
							router.replace("/settings");
						}, 1000);
					}}
					title="Clear App Cache?"
					description="This will clear all local storage and IndexedDB data. Are you sure you want to continue?"
					confirmText="Yes, Clear Cache"
					cancelText="Cancel"
					variant="ghost"
					className="text-destructive border-1 shadow-none hover:text-red-500 cursor-pointer"
				>
					Clear App Cache
				</ConfirmDestructiveDialogActionButton>
			</div>
		</div>
	);
};

export default SettingsPage;
