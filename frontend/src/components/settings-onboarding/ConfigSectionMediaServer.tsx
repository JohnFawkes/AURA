"use client";

import { ValidateURL } from "@/helper/validation/validate-url";
import { checkMediaServerNewInfoConnectionStatus } from "@/services/settings-onboarding/api-mediaserver-connection";
import { fetchMediaServerLibraryOptions } from "@/services/settings-onboarding/api-mediaserver-library-options";
import { Plus, RefreshCcw } from "lucide-react";

import React, { useCallback, useEffect, useRef, useState } from "react";

import {
	CONNECTION_STATUS_COLORS_BG,
	ConfigConnectionStatus,
} from "@/components/settings-onboarding/ConfigSectionSonarrRadarr";
import { PopoverHelp } from "@/components/shared/popover-help";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

import { cn } from "@/lib/cn";

import { AppConfigMediaServer, AppConfigMediaServerLibrary } from "@/types/config/config-app";

interface ConfigSectionMediaServerProps {
	value: AppConfigMediaServer;
	editing: boolean;
	configAlreadyLoaded: boolean;
	dirtyFields?: Partial<Record<keyof AppConfigMediaServer, boolean>>;
	onChange: <K extends keyof AppConfigMediaServer>(field: K, value: AppConfigMediaServer[K]) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigMediaServer, string>>) => void;
}

const SERVER_TYPES = ["Plex", "Emby", "Jellyfin"];
const USER_ID_REQUIRED_TYPES = new Set<string>(["Emby", "Jellyfin"]);
const SEASON_NAMING_CONVENTION_OPTIONS = ["1", "2"];
const SEASON_NAMING_CONVENTION_REQUIRED_TYPES = new Set<string>(["Plex"]);

export function GetConnectionColor(status: "unknown" | "ok" | "error"): "green-500" | "red-500" | "gray-500" {
	switch (status) {
		case "ok":
			return "green-500";
		case "error":
			return "red-500";
		default:
			return "gray-500";
	}
}

export const ConfigSectionMediaServer: React.FC<ConfigSectionMediaServerProps> = ({
	value,
	editing,
	configAlreadyLoaded,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("");

	// Normalize libraries to array (avoid null errors)
	const libraries: AppConfigMediaServerLibrary[] = React.useMemo(
		() => (Array.isArray(value.Libraries) ? value.Libraries : []),
		[value.Libraries]
	);

	useEffect(() => {
		if (!Array.isArray(value.Libraries)) {
			onChange("Libraries", [] as AppConfigMediaServerLibrary[]);
		}
	}, [onChange, value.Libraries]);

	const [remoteTokenError, setRemoteTokenError] = useState<string | null>(null);
	const [testingToken, setTestingToken] = useState(false);
	const [connectionStatus, setConnectionStatus] = useState<ConfigConnectionStatus>({
		status: "unknown",
		color: GetConnectionColor("unknown"),
	});

	const valueRef = React.useRef(value);
	React.useEffect(() => {
		valueRef.current = value;
	}, [value]);

	const [libraryFetchLoading, setLibraryFetchLoading] = useState(false);

	const typeNormalized = value.Type.trim();
	const newLibRef = useRef<HTMLInputElement | null>(null);
	const hasRunInitialValidation = useRef(false);

	const errors = React.useMemo<Partial<Record<keyof AppConfigMediaServer, string>>>(() => {
		const errs: Partial<Record<keyof AppConfigMediaServer, string>> = {};

		// Type
		if (!typeNormalized) errs.Type = "Type is required.";
		else if (!SERVER_TYPES.includes(typeNormalized)) errs.Type = "Unsupported type.";

		// URL
		if (!value.URL.trim()) errs.URL = "URL is required.";
		else {
			const urlErr = ValidateURL(value.URL.trim());
			if (urlErr) errs.URL = urlErr;
		}

		// Token
		if (!value.Token.trim()) errs.Token = "Token is required.";
		if (remoteTokenError) errs.Token = remoteTokenError;

		// User ID requirement
		if (USER_ID_REQUIRED_TYPES.has(typeNormalized) && !value.UserID?.trim()) {
			errs.UserID = "User ID should be set automatically after URL & Token are valid.";
		}

		// Season naming
		if (
			SEASON_NAMING_CONVENTION_REQUIRED_TYPES.has(typeNormalized) &&
			!SEASON_NAMING_CONVENTION_OPTIONS.includes((value.SeasonNamingConvention ?? "").trim())
		) {
			errs.SeasonNamingConvention = "Season naming convention is required for this server type.";
		}

		// Libraries
		if (libraries.length === 0) {
			errs.Libraries = "Add at least one library.";
		} else {
			if (libraries.some((l) => !l.Name?.trim())) errs.Libraries = "Library names cannot be empty.";
			if (!errs.Libraries) {
				const seen = new Set<string>();
				for (const l of libraries) {
					const n = (l.Name || "").trim().toLowerCase();
					if (!n) continue;
					if (seen.has(n)) {
						errs.Libraries = "Duplicate library names are not allowed.";
						break;
					}
					seen.add(n);
				}
			}
		}

		return errs;
	}, [
		typeNormalized,
		value.URL,
		value.Token,
		value.UserID,
		value.SeasonNamingConvention,
		libraries,
		remoteTokenError,
	]);

	// Emit errors upward
	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	// Reset remote token error when URL or Token changes
	useEffect(() => {
		setRemoteTokenError(null);
	}, [value.Token, value.URL]);

	const runRemoteValidation = useCallback(
		async (showToast = true) => {
			setConnectionStatus({ status: "unknown", color: GetConnectionColor("unknown") });
			const current = valueRef.current;
			if (!current.Token.trim()) {
				setRemoteTokenError("Token is required.");
				setConnectionStatus({ status: "error", color: GetConnectionColor("error") });
				return;
			}
			if (!current.URL.trim()) {
				setRemoteTokenError("URL is required.");
				setConnectionStatus({ status: "error", color: GetConnectionColor("error") });
				return;
			}

			setTestingToken(true);
			const start = Date.now();
			const { ok, message, data } = await checkMediaServerNewInfoConnectionStatus(current, showToast);
			const elapsed = Date.now() - start;
			const minDelay = 400; // milliseconds

			if (elapsed < minDelay) {
				await new Promise((resolve) => setTimeout(resolve, minDelay - elapsed));
			}
			setTestingToken(false);

			if (ok) {
				setRemoteTokenError(null);
				setConnectionStatus({ status: "ok", color: GetConnectionColor("ok") });
				if (data?.UserID && data.UserID !== current.UserID) {
					onChange("UserID", data.UserID);
				}
			} else {
				setRemoteTokenError(message || "Token invalid");
				setConnectionStatus({ status: "error", color: GetConnectionColor("error") });
			}
		},
		[onChange]
	);

	useEffect(() => {
		if (configAlreadyLoaded && !hasRunInitialValidation.current) {
			runRemoteValidation(false); // No toast on initial load
			hasRunInitialValidation.current = true;
		}
	}, [configAlreadyLoaded, runRemoteValidation]);

	// Helpers
	const addLibraryByName = (name: string) => {
		const trimmed = name.trim();
		if (!trimmed) return;
		if (libraries.some((l) => l.Name.trim().toLowerCase() === trimmed.toLowerCase())) return;
		onChange("Libraries", [...libraries, { Name: trimmed, SectionID: "", Type: "" }]);
	};

	const removeLibraryByIndex = (index: number) => {
		onChange(
			"Libraries",
			libraries.filter((_, i) => i !== index)
		);
	};

	const replaceLibraries = (names: string[]) => {
		onChange(
			"Libraries",
			names.map((n) => ({ Name: n, SectionID: "", Type: "" }))
		);
	};

	const fetchServerLibraries = async () => {
		if (!editing || libraryFetchLoading) return;
		setLibraryFetchLoading(true);
		const { ok, data } = await fetchMediaServerLibraryOptions(value);
		setLibraryFetchLoading(false);
		if (!ok || !Array.isArray(data)) {
			return;
		}
		replaceLibraries(data);
	};

	return (
		<Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-2">
					<h2 className={`text-xl font-semibold text-${connectionStatus.color}`}>Media Server</h2>
					<span
						className={`h-2 w-2 rounded-full ${CONNECTION_STATUS_COLORS_BG[connectionStatus.status]} animate-pulse`}
						title={`Connection status: ${connectionStatus.status}`}
					/>
				</div>
				<Button
					variant="outline"
					size="sm"
					hidden={editing}
					disabled={editing || testingToken}
					onClick={() => runRemoteValidation()}
					className="cursor-pointer hover:text-primary"
				>
					{testingToken ? "Testing..." : "Test Connection"}
				</Button>
			</div>

			{/* Type */}
			<div className={cn("space-y-1")}>
				<div className="flex items-center justify-between">
					<Label>Type</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-media-server-type">
							<p>The type of Media Server </p>
							<ul className="list-disc list-inside mt-1">
								<li>Plex</li>
								<li>Emby</li>
								<li>Jellyfin</li>
							</ul>
						</PopoverHelp>
					)}
				</div>
				<Select
					disabled={!editing}
					value={value.Type}
					onValueChange={(v) => {
						value.URL = "";
						value.Token = "";
						value.UserID = "";
						value.Libraries = [];
						onChange("Type", v);
					}}
				>
					<SelectTrigger
						id="media-server-type-trigger"
						className={cn("w-full", dirtyFields.Type && "border-amber-500")}
					>
						<SelectValue placeholder="Select type..." />
					</SelectTrigger>
					<SelectContent>
						{SERVER_TYPES.map((t) => (
							<SelectItem className="cursor-pointer" key={t} value={t}>
								{t}
							</SelectItem>
						))}
						<SelectScrollUpButton />
						<SelectScrollDownButton />
					</SelectContent>
				</Select>
				{errors.Type && <p className="text-xs text-red-500">{errors.Type}</p>}
			</div>

			{/* URL */}
			<div className={cn("space-y-1")}>
				<div className="flex items-center justify-between">
					<Label>URL</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-media-server-url">
							<p className="font-medium mb-1">Base server URL</p>
							<p>Examples:</p>
							<ul className="list-disc list-inside mb-1">
								<li>https://{value.Type.toLowerCase()}.domain.com</li>
								<li>http://192.168.1.10:{value.Type === "Plex" ? "32400" : "8096"}</li>
								<li>
									http://{value.Type.toLowerCase()}:{value.Type === "Plex" ? "32400" : "8096"}
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
					placeholder="https://server.example.com"
					value={value.URL}
					onChange={(e) => onChange("URL", e.target.value)}
					onBlur={() => runRemoteValidation()}
					className={cn("w-full", dirtyFields.URL && "border-amber-500")}
				/>
				{errors.URL && <p className="text-xs text-red-500">{errors.URL}</p>}
			</div>

			{/* Token */}
			<div className={cn("space-y-1")}>
				<div className="flex items-center justify-between">
					<Label>Token</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-media-server-token">
							<p>API token used to authenticate requests to the media server.</p>
						</PopoverHelp>
					)}
				</div>
				<Input
					disabled={!editing}
					type="text"
					placeholder="API token"
					value={value.Token}
					onChange={(e) => onChange("Token", e.target.value)}
					onBlur={() => runRemoteValidation()}
					className={cn("w-full", dirtyFields.Token && "border-amber-500")}
				/>
				{errors.Token && <p className="text-xs text-red-500">{errors.Token}</p>}
			</div>

			{/* User ID (Emby / Jellyfin) */}
			{USER_ID_REQUIRED_TYPES.has(value.Type) && (
				<div className={cn("space-y-1")}>
					<div className="flex items-center justify-between">
						<Label>User ID</Label>
						{editing && (
							<PopoverHelp ariaLabel="help-media-server-user-id">
								<p>
									User ID is required for Emby/Jellyfin. It should be set automatically after URL &
									Token are valid.
								</p>
							</PopoverHelp>
						)}
					</div>
					<Input
						disabled={true}
						value={value.UserID ?? ""}
						placeholder="Emby/Jellyfin user id"
						className={cn("w-full", dirtyFields.UserID && "border-amber-500")}
					/>
					{errors.UserID && <p className="text-xs text-red-500">{errors.UserID}</p>}
				</div>
			)}

			{/* Season Naming Convention (Plex) */}
			{SEASON_NAMING_CONVENTION_REQUIRED_TYPES.has(value.Type) && (
				<div className={cn("space-y-1")}>
					<div className="flex items-center justify-between">
						<Label>Season Naming Convention</Label>
						{editing && (
							<PopoverHelp ariaLabel="help-media-server-season-naming-convention">
								<div className="space-y-3">
									<div>
										<p className="font-medium mb-1">Season Naming Convention</p>
										<p className="text-muted-foreground">
											How Plex season folders / labels are formatted.
										</p>
									</div>
									<ul className="space-y-1">
										<li className="flex items-center gap-2">
											<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono ">
												1
											</span>
											<span>Season 1</span>
										</li>
										<li className="flex items-center gap-2">
											<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono">
												2
											</span>
											<span>Season 01 (zero‑padded)</span>
										</li>
									</ul>
									<p className="text-muted-foreground">Used for display / folder naming logic.</p>
								</div>
							</PopoverHelp>
						)}
					</div>
					<Select
						disabled={!editing}
						value={value.SeasonNamingConvention}
						onValueChange={(v) => onChange("SeasonNamingConvention", v)}
					>
						<SelectTrigger
							id="media-server-season-naming-convention-trigger"
							className={cn("w-full", dirtyFields.SeasonNamingConvention && "border-amber-500")}
						>
							<SelectValue placeholder="Select convention..." />
						</SelectTrigger>
						<SelectContent>
							{SEASON_NAMING_CONVENTION_OPTIONS.map((o) => (
								<SelectItem key={o} value={o}>
									{o}
								</SelectItem>
							))}
							<SelectScrollUpButton />
							<SelectScrollDownButton />
						</SelectContent>
					</Select>
					{errors.SeasonNamingConvention && (
						<p className="text-xs text-red-500">{errors.SeasonNamingConvention}</p>
					)}
				</div>
			)}

			{/* Libraries */}
			<div className={cn("space-y-3")}>
				<div className="flex items-center">
					<Label>Libraries</Label>
					{editing && (
						<span className="flex items-center gap-2 ml-3">
							<Button
								onClick={fetchServerLibraries}
								variant="outline"
								size="icon"
								className="h-7 w-7"
								title="Refresh libraries from server"
								disabled={libraryFetchLoading}
							>
								{libraryFetchLoading ? (
									<RefreshCcw className="h-4 w-4 animate-spin" />
								) : (
									<RefreshCcw className="h-4 w-4" />
								)}
							</Button>
						</span>
					)}
				</div>

				<div className="flex flex-wrap gap-2">
					{libraries.length === 0 && (
						<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-xs text-muted-foreground">
							No libraries added
						</span>
					)}
					{libraries.map((lib, i) => (
						<Badge
							key={i}
							className={cn(
								"cursor-pointer select-none transition duration-200 px-3 py-1 text-xs font-normal",
								editing
									? "bg-secondary text-secondary-foreground hover:bg-red-500 hover:text-white"
									: "bg-muted text-muted-foreground",
								dirtyFields.Libraries && "border-amber-500"
							)}
							onClick={() => {
								if (!editing) return;
								removeLibraryByIndex(i);
							}}
							title={editing ? "Remove library" : lib.Name}
						>
							{lib.Name}
						</Badge>
					))}

					{editing && (
						<form
							onSubmit={(e) => {
								e.preventDefault();
								if (!newLibRef.current) return;
								addLibraryByName(newLibRef.current.value);
								newLibRef.current.value = "";
							}}
							className="flex items-center gap-1"
						>
							<Input
								ref={newLibRef}
								placeholder="Add library..."
								className="h-7 w-40 text-xs"
								onKeyDown={(e) => {
									if (e.key === "Enter") {
										e.preventDefault();
										const target = e.currentTarget;
										addLibraryByName(target.value);
										target.value = "";
									}
								}}
							/>
							<Button type="submit" variant="outline" size="icon" className="h-7 w-7" title="Add library">
								<Plus className="h-4 w-4" />
							</Button>
						</form>
					)}
				</div>
				{errors.Libraries && <p className="text-xs text-red-500">{errors.Libraries}</p>}
			</div>
		</Card>
	);
};
