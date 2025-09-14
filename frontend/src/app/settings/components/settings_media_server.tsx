"use client";

import { getMediaServerLibraryOptions } from "@/app/settings/services/library_options";
import { checkMediaServerNewInfoConnectionStatus } from "@/app/settings/services/media_server_check_connection";
import { Plus, RefreshCcw } from "lucide-react";

import React, { useEffect, useRef, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

import { cn } from "@/lib/utils";

import { AppConfigMediaServer, AppConfigMediaServerLibrary } from "@/types/config";

interface MediaServerSectionProps {
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

// Domain / host validation
const domainHostRegex = /^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/;
const singleLabelHostRegex = /^[a-zA-Z0-9-]+$/;

function isValidIPv4(host: string): boolean {
	if (!/^[0-9.]+$/.test(host)) return false;
	const parts = host.split(".");
	if (parts.length !== 4) return false;
	return parts.every((p) => {
		if (p.length === 0 || (p.length > 1 && p.startsWith("0"))) return false;
		const n = Number(p);
		return Number.isInteger(n) && n >= 0 && n <= 255;
	});
}

export function ValidateURL(raw: string): string | null {
	const value = raw.trim();
	if (!/^https?:\/\//i.test(value)) return "Must start with http:// or https://";
	let parsed: URL;
	try {
		parsed = new URL(value);
	} catch {
		return "Invalid URL format.";
	}
	const protocol = parsed.protocol.toLowerCase();
	if (protocol !== "http:" && protocol !== "https:") return "Only http and https are allowed.";
	const host = parsed.hostname;
	const isIPv4 = isValidIPv4(host);
	const looksNumeric = /^[0-9.]+$/.test(host);
	if (looksNumeric && !isIPv4) return "Invalid IPv4 address (must have 4 octets).";
	const validatePort = () => {
		if (!parsed.port) return "Port is required.";
		const portNum = Number(parsed.port);
		if (!(portNum > 0 && portNum <= 65535)) return "Invalid port number.";
		return "";
	};
	if (isIPv4) {
		if (!parsed.port) return "IP address requires a port number.";
		const msg = validatePort();
		if (msg) return msg;
	} else if (host.includes(".")) {
		if (!domainHostRegex.test(host)) return "Invalid domain.";
		if (parsed.port) {
			const portNum = Number(parsed.port);
			if (!(portNum > 0 && portNum <= 65535)) return "Invalid port number.";
		}
	} else {
		if (!singleLabelHostRegex.test(host)) return "Invalid host.";
		const msg = validatePort();
		if (msg) return msg;
	}
	return null;
}

export const MediaServerSection: React.FC<MediaServerSectionProps> = ({
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

	const valueRef = React.useRef(value);
	React.useEffect(() => {
		valueRef.current = value;
	}, [value]);

	const [libraryFetchLoading, setLibraryFetchLoading] = useState(false);

	const typeNormalized = value.Type.trim();
	const newLibRef = useRef<HTMLInputElement | null>(null);

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
			errs.UserID = "User ID is required for this server type.";
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

	const runRemoteValidation = React.useCallback(async () => {
		const current = valueRef.current; // latest value
		if (!current.Token.trim()) {
			setRemoteTokenError("Token is required.");
			return;
		}
		if (!current.URL.trim()) {
			setRemoteTokenError("URL is required.");
			return;
		}

		setTestingToken(true);
		const { ok, message, data } = await checkMediaServerNewInfoConnectionStatus(current);
		setTestingToken(false);

		if (ok) {
			setRemoteTokenError(null);

			// Set UserID only if server provided one and we don't already have it (or it changed)
			if (data?.UserID && data.UserID !== current.UserID) {
				onChange("UserID", data.UserID);
			}
		} else {
			setRemoteTokenError(message || "Token invalid");
		}
	}, [onChange]);

	useEffect(() => {
		if (configAlreadyLoaded) runRemoteValidation();
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
		const { ok, data } = await getMediaServerLibraryOptions(value);
		setLibraryFetchLoading(false);
		if (!ok || !Array.isArray(data)) {
			return;
		}
		replaceLibraries(data);
	};

	return (
		<Card className="p-5 space-y-1">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Media Server</h2>
				<Button
					variant="outline"
					size="sm"
					disabled={editing || testingToken}
					hidden={editing}
					onClick={() => runRemoteValidation()}
				>
					{testingToken ? "Testing..." : "Test Connection"}
				</Button>
			</div>

			{/* Type */}
			<div
				className={cn(
					"space-y-1",
					(dirtyFields.Type || errors.Type) && "rounded-md",
					errors.Type ? "border border-red-500 p-3" : dirtyFields.Type && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Type</Label>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-media-server-type"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-64 text-xs leading-snug"
							>
								<p>Select the media server platform (Plex, Emby, Jellyfin).</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
				<Select disabled={!editing} value={value.Type} onValueChange={(v) => onChange("Type", v)}>
					<SelectTrigger className="w-full" id="media-server-type-trigger">
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
			<div
				className={cn(
					"space-y-1",
					(dirtyFields.URL || errors.URL) && "rounded-md",
					errors.URL ? "border border-red-500 p-3" : dirtyFields.URL && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>URL</Label>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-media-server-url"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-72 text-xs leading-snug"
							>
								<p>
									Base server URL. Domains may omit port. IPv4 addresses must include a port. Example:
									https://plex.domain.com, http://192.168.1.10:32400 or http://plex:32400
								</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
				<Input
					disabled={!editing}
					placeholder="https://server.example.com"
					value={value.URL}
					onChange={(e) => onChange("URL", e.target.value)}
					onBlur={() => runRemoteValidation()}
				/>
				{errors.URL && <p className="text-xs text-red-500">{errors.URL}</p>}
			</div>

			{/* Token */}
			<div
				className={cn(
					"space-y-1",
					(dirtyFields.Token || errors.Token) && "rounded-md",
					errors.Token ? "border border-red-500 p-3" : dirtyFields.Token && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Token</Label>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-media-server-token"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-72 text-xs leading-snug"
							>
								<p>API token used to authenticate requests to the media server.</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
				<Input
					disabled={!editing}
					type="text"
					placeholder="API token"
					value={value.Token}
					onChange={(e) => onChange("Token", e.target.value)}
					onBlur={() => runRemoteValidation()}
				/>
				{errors.Token && <p className="text-xs text-red-500">{errors.Token}</p>}
			</div>

			{/* User ID (Emby / Jellyfin) */}
			{USER_ID_REQUIRED_TYPES.has(value.Type) && (
				<div
					className={cn(
						"space-y-1",
						(dirtyFields.UserID || errors.UserID) && "rounded-md",
						errors.UserID
							? "border border-red-500 p-3"
							: dirtyFields.UserID && "border border-amber-500 p-3"
					)}
				>
					<div className="flex items-center justify-between">
						<Label>User ID</Label>
						{editing && (
							<Popover>
								<PopoverTrigger asChild>
									<Button
										variant="outline"
										className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
										aria-label="help-media-server-user-id"
									>
										?
									</Button>
								</PopoverTrigger>
								<PopoverContent
									side="right"
									align="center"
									sideOffset={8}
									className="w-64 text-xs leading-snug"
								>
									<p>Required for Emby / Jellyfin. The internal user identifier.</p>
								</PopoverContent>
							</Popover>
						)}
					</div>
					<Input
						disabled={true}
						value={value.UserID ?? ""}
						placeholder="Emby/Jellyfin user id"
						aria-invalid={!!errors.UserID}
					/>
					{errors.UserID && <p className="text-xs text-red-500">{errors.UserID}</p>}
				</div>
			)}

			{/* Season Naming Convention (Plex) */}
			{SEASON_NAMING_CONVENTION_REQUIRED_TYPES.has(value.Type) && (
				<div
					className={cn(
						"space-y-1",
						(dirtyFields.SeasonNamingConvention || errors.SeasonNamingConvention) && "rounded-md",
						errors.SeasonNamingConvention
							? "border border-red-500 p-3"
							: dirtyFields.SeasonNamingConvention && "border border-amber-500 p-3"
					)}
				>
					<div className="flex items-center justify-between">
						<Label>Season Naming Convention</Label>
						{editing && (
							<Popover>
								<PopoverTrigger asChild>
									<Button
										variant="outline"
										className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
										aria-label="help-media-server-season-naming-convention"
									>
										?
									</Button>
								</PopoverTrigger>
								<PopoverContent
									side="right"
									align="center"
									sideOffset={8}
									className="w-72 text-xs leading-snug"
								>
									<div className="space-y-3">
										<div>
											<p className="font-medium mb-1">Season Naming Convention</p>
											<p className="text-[11px] text-muted-foreground">
												How Plex season folders / labels are formatted.
											</p>
										</div>
										<ul className="space-y-1">
											<li className="flex items-center gap-2">
												<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono text-[10px]">
													1
												</span>
												<span className="text-[11px]">Season 1</span>
											</li>
											<li className="flex items-center gap-2">
												<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono text-[10px]">
													2
												</span>
												<span className="text-[11px]">Season 01 (zero‑padded)</span>
											</li>
										</ul>
										<p className="text-[10px] text-muted-foreground">
											Used for display / folder naming logic.
										</p>
									</div>
								</PopoverContent>
							</Popover>
						)}
					</div>
					<Select
						disabled={!editing}
						value={value.SeasonNamingConvention}
						onValueChange={(v) => onChange("SeasonNamingConvention", v)}
					>
						<SelectTrigger className="w-full" id="media-server-season-naming-convention-trigger">
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
			<div
				className={cn(
					"space-y-3",
					(dirtyFields.Libraries || errors.Libraries) && "rounded-md",
					errors.Libraries
						? "border border-red-500 p-3"
						: dirtyFields.Libraries && "border border-amber-500 p-3"
				)}
			>
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
						<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-[11px] text-muted-foreground">
							No libraries added
						</span>
					)}
					{libraries.map((lib, i) => (
						<Badge
							key={i}
							className={cn(
								"cursor-pointer select-none transition duration-200 px-3 py-1 text-[11px] font-normal",
								editing
									? "bg-secondary text-secondary-foreground hover:bg-red-500 hover:text-white"
									: "bg-muted text-muted-foreground"
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
