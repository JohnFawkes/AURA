import { ValidateURL } from "@/helper/validation/validate-url";
import { checkSonarrRadarrNewAPIKeyStatusResult } from "@/services/settings-onboarding/api-sonarr-radarr-test-connection";
import { Plus, TestTube, Trash2 } from "lucide-react";

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { GetConnectionColor } from "@/components/settings-onboarding/ConfigSectionMediaServer";
import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { cn } from "@/lib/cn";

import {
	AppConfigMediaServerLibrary,
	AppConfigSonarrRadarrApp,
	AppConfigSonarrRadarrApps,
} from "@/types/config/config-app";

interface ConfigSectionSonarrRadarrProps {
	value: AppConfigSonarrRadarrApps;
	editing: boolean;
	dirtyFields?: {
		Applications?: Array<Partial<{ Type: boolean; Library: boolean; URL: boolean; APIKey: boolean }>>;
	};
	onChange: <K extends keyof AppConfigSonarrRadarrApps>(field: K, value: AppConfigSonarrRadarrApps[K]) => void;
	errorsUpdate?: (errors: Record<string, string>) => void;
	configAlreadyLoaded: boolean;
	libraries: AppConfigMediaServerLibrary[];
}

const SR_TYPES = ["Sonarr", "Radarr"] as const;

export type ConfigConnectionStatus = {
	status: "ok" | "error" | "unknown";
	color: "green-500" | "red-500" | "gray-500";
};

export const CONNECTION_STATUS_COLORS_BG: Record<string, string> = {
	ok: "bg-green-500",
	error: "bg-red-500",
	unknown: "bg-gray-400",
};

export const ConfigSectionSonarrRadarr: React.FC<ConfigSectionSonarrRadarrProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
	configAlreadyLoaded,
	libraries,
}) => {
	const prevErrorsRef = useRef<string>("");
	const hasRunInitialValidation = useRef(false);

	// Local select state for adding new application
	const [newAppType, setNewAppType] = useState<string>("Sonarr");

	const apps = useMemo(() => (Array.isArray(value.Applications) ? value.Applications : []), [value.Applications]);

	// State to track app connection testing
	const [appConnectionStatus, setAppConnectionStatus] = useState<Record<number, ConfigConnectionStatus>>({});
	const [remoteTokenErrors, setRemoteTokenErrors] = useState<Record<number, string | null>>({});

	// ----- Validation -----
	const errors = useMemo(() => {
		const errs: Record<string, string> = {};

		apps.forEach((app, index) => {
			if (!app.Type) {
				errs[`Applications.[${index}].Type`] = "Type is required";
			} else {
				if (!SR_TYPES.includes(app.Type as (typeof SR_TYPES)[number])) {
					errs[`Applications.[${index}].Type`] = "Invalid Type, must be Sonarr or Radarr";
				}
			}
			if (!app.Library) {
				errs[`Applications.[${index}].Library`] = "Library is required";
			}
			if (!app.URL) {
				const rawURL = (app.URL || "").trim();
				if (!rawURL) errs[`Applications.[${index}].URL`] = "URL is required";
				else {
					const urlErr = ValidateURL(rawURL);
					if (urlErr) errs[`Applications.[${index}].URL`] = urlErr;
				}
			}
			if (!app.APIKey) {
				errs[`Applications.[${index}].APIKey`] = "API Key is required";
			}
			// Remote error (overrides local message if present)
			if (remoteTokenErrors[index]) {
				errs[`Applications.[${index}].APIKey`] = remoteTokenErrors[index] || "Connection failed";
			}
		});

		return errs;
	}, [apps, remoteTokenErrors]);

	// Emit errors if changed
	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	// ----- Mutators -----
	const setApps = (next: AppConfigSonarrRadarrApp[]) => onChange("Applications", next);

	const addApp = () => {
		if (!editing) return;
		const type = newAppType as (typeof SR_TYPES)[number];
		let newEntry: AppConfigSonarrRadarrApp;
		if (type === "Sonarr") {
			newEntry = { Type: "Sonarr", Library: "", URL: "", APIKey: "" };
		} else {
			newEntry = { Type: "Radarr", Library: "", URL: "", APIKey: "" };
		}
		setApps([...apps, newEntry]);
	};

	const removeApp = (idx: number) => {
		if (!editing) return;
		const next = apps.slice();
		next.splice(idx, 1);
		setApps(next);
	};

	const updateApp = <K extends keyof AppConfigSonarrRadarrApp>(
		idx: number,
		field: K,
		value: AppConfigSonarrRadarrApp[K]
	) => {
		if (!editing) return;
		const next = apps.slice();
		next[idx] = { ...next[idx], [field]: value };
		setApps(next);
	};

	const runRemoteValidation = useCallback(
		async (idx: number, showToast = true) => {
			const app = apps[idx];
			if (!app || !app.URL || !app.APIKey) return;

			// Set to unknown while testing
			setAppConnectionStatus((s) => ({
				...s,
				[idx]: { status: "unknown", color: GetConnectionColor("unknown") },
			}));

			try {
				const start = Date.now();
				const { ok, message } = await checkSonarrRadarrNewAPIKeyStatusResult(app, showToast);
				const elapsed = Date.now() - start;
				const minDelay = 400; // milliseconds

				if (elapsed < minDelay) {
					await new Promise((resolve) => setTimeout(resolve, minDelay - elapsed));
				}

				if (ok) {
					setRemoteTokenErrors((s) => ({ ...s, [idx]: null }));
					setAppConnectionStatus((s) => ({ ...s, [idx]: { status: "ok", color: GetConnectionColor("ok") } }));
				} else {
					setRemoteTokenErrors((s) => ({ ...s, [idx]: message || "Connection failed" }));
					setAppConnectionStatus((s) => ({
						...s,
						[idx]: { status: "error", color: GetConnectionColor("error") },
					}));
				}
			} catch {
				setRemoteTokenErrors((s) => ({ ...s, [idx]: "Connection failed" }));
				setAppConnectionStatus((s) => ({
					...s,
					[idx]: { status: "error", color: GetConnectionColor("error") },
				}));
			}
		},
		[apps]
	);

	useEffect(() => {
		if (configAlreadyLoaded && !hasRunInitialValidation.current) {
			// Run remote validation for all apps that have URL and APIKey set
			apps.forEach((app, idx) => {
				if (app.URL && app.APIKey) {
					// Delay slightly to allow UI to settle
					setTimeout(() => {
						runRemoteValidation(idx, false);
					}, 200 * idx);
				}
			});
			hasRunInitialValidation.current = true;
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [configAlreadyLoaded, runRemoteValidation]);

	// ----- Render -----
	return (
		<Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold text-blue-500">Sonarr & Radarr</h2>
				<div className="flex items-center">
					{editing && libraries.length > 0 && (
						<div className="flex items-center gap-2">
							<Select value={newAppType} onValueChange={(v) => setNewAppType(v)}>
								<SelectTrigger className="h-8 w-36">
									<SelectValue placeholder="Type" />
								</SelectTrigger>
								<SelectContent>
									{SR_TYPES.map((sr) => (
										<SelectItem key={sr} value={sr}>
											{sr}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
							<Button type="button" variant="outline" size="sm" onClick={addApp}>
								<Plus className="h-4 w-4 mr-1" />
								Add
							</Button>
						</div>
					)}
				</div>
			</div>

			{/* Applications */}
			<div className={cn("space-y-3", "rounded-md")}>
				{libraries.length === 0 ? (
					<p className="text-sm text-red-500">
						No Media Server libraries found. Please configure your Media Server first.
					</p>
				) : apps.length === 0 ? (
					<p className="text-sm text-muted-foreground">No applications added yet.</p>
				) : (
					<>
						{apps.map((app, idx) => {
							const appDirty = dirtyFields.Applications?.[idx] as Partial<{
								Type: string;
								Library: string;
								URL: string;
								APIKey: string;
							}>;
							const appErrorEntries = Object.entries(errors).filter(([k]) =>
								k.startsWith(`Applications.[${idx}]`)
							);
							const appErrors = appErrorEntries.map(([, msg]) => msg);

							// Field-level helpers (key-based)
							const hasError = (suffix: string) => appErrorEntries.some(([k]) => k.endsWith(suffix));
							const statusObj = appConnectionStatus[idx] || {
								status: "unknown",
								color: "gray-500",
							};
							return (
								<div
									key={idx}
									className={cn(
										"space-y-3 rounded-md border p-3 transition",
										appErrors.length
											? "border-red-500"
											: appDirty && Object.values(appDirty).some(Boolean)
												? "border-amber-500"
												: "border-muted"
									)}
								>
									<div className="flex items-center justify-between gap-3">
										<div className="flex items-center gap-2">
											<h2 className={`text-xl font-semibold text-${statusObj.color}`}>
												{app.Type}
											</h2>
											<span
												className={`h-2 w-2 rounded-full ${CONNECTION_STATUS_COLORS_BG[statusObj.status]} animate-pulse`}
												title={`Connection status: ${statusObj.status}`}
											/>
										</div>
										<div className="flex items-center gap-2">
											<Button
												variant="outline"
												size="sm"
												disabled={!app.URL || !app.APIKey}
												hidden={!app.APIKey && !app.URL}
												onClick={() => {
													runRemoteValidation(idx);
												}}
												aria-label="test-app-connection"
											>
												<TestTube className="h-4 w-4 mr-1" />{" "}
											</Button>
											{editing && (
												<Button
													variant="ghost"
													size="icon"
													onClick={() => removeApp(idx)}
													aria-label="help-apps-remove-app"
													className="bg-red-700"
												>
													<Trash2 className="h-4 w-4" />
												</Button>
											)}
										</div>
									</div>

									{/* Library Field */}
									<div className={cn("space-y-1", "rounded-md")}>
										<div className="flex items-center justify-between">
											<Label>Library</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-apps-sr-library">
													<p className="mb-2 font-medium">Library</p>
													<p>
														The Media Server library that this {app.Type} instance will
														manage.
													</p>
												</PopoverHelp>
											)}
										</div>
										<Select
											disabled={!editing}
											value={app.Library || ""}
											onValueChange={(v) => updateApp(idx, "Library", v)}
										>
											<SelectTrigger
												className={cn(
													"h-8",
													!app.Library && "text-muted-foreground",
													!hasError("Library") &&
														appDirty?.Library &&
														"rounded-md border border-amber-500 p-3"
												)}
											>
												<SelectValue
													placeholder={
														libraries.length === 0 ? "No libraries found" : "Select library"
													}
												/>
											</SelectTrigger>
											{libraries.length > 0 && (
												<SelectContent>
													{libraries.map((lib) => (
														<SelectItem key={lib.Name} value={lib.Name}>
															{lib.Name}
														</SelectItem>
													))}
												</SelectContent>
											)}
										</Select>
										{appErrorEntries
											.filter(([k]) => k.endsWith("Library"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>

									{/* URL Field */}
									<div className={cn("space-y-1", "rounded-md")}>
										<div className="flex items-center justify-between">
											<Label>URL</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-apps-sr-url">
													<p className="font-medium mb-1">App URL</p>
													<p>Examples:</p>
													<ul className="list-disc list-inside mb-1">
														<li>https://{app.Type.toLowerCase()}.domain.com</li>
														<li>
															http://192.168.1.10:
															{app.Type === "Sonarr" ? "8989" : "7878"}
														</li>
														<li>
															http://{app.Type.toLowerCase()}:
															{app.Type === "Sonarr" ? "8989" : "7878"}
														</li>
													</ul>
													<p>Rules:</p>
													<ul className="list-disc list-inside">
														<li>Must be a valid URL</li>
														<li>Must include http:// or https://</li>
														<li>IPv4 addresses must include a port</li>
													</ul>
												</PopoverHelp>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="http://app.domain.com"
											value={app.URL || ""}
											onChange={(e) => updateApp(idx, "URL", e.target.value)}
											className={cn(appDirty?.URL && "rounded-md border border-amber-500 p-3")}
										/>
										{appErrorEntries
											.filter(([k]) => k.endsWith("URL"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>

									{/* API Key Field */}
									<div className={cn("space-y-1", "rounded-md")}>
										<div className="flex items-center justify-between">
											<Label>API Key</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-apps-sr-apikey">
													<p className="mb-2 font-medium">API Key</p>
													<p>The API Key for your {app.Type} instance.</p>
													<p>
														This can usually be found in the settings of your application.
													</p>
												</PopoverHelp>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="API Key"
											value={app.APIKey || ""}
											onChange={(e) => updateApp(idx, "APIKey", e.target.value)}
											onBlur={() => {
												runRemoteValidation(idx);
											}}
											className={cn(appDirty?.APIKey && "rounded-md border border-amber-500 p-3")}
										/>
										{appErrorEntries
											.filter(([k]) => k.endsWith("APIKey"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
								</div>
							);
						})}{" "}
					</>
				)}
			</div>
		</Card>
	);
};
