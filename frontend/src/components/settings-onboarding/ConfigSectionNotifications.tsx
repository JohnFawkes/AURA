"use client";

import { ValidateURL } from "@/helper/validation/validate-url";
import { sendTestNotification } from "@/services/settings-onboarding/api-notifications-test";
import { HelpCircle, Plus, Trash2 } from "lucide-react";

import React, { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import {
	AppConfigNotificationDiscord,
	AppConfigNotificationGotify,
	AppConfigNotificationProviders,
	AppConfigNotificationPushover,
	AppConfigNotifications,
} from "@/types/config/config-app";

interface ConfigSectionNotificationsProps {
	value: AppConfigNotifications;
	editing: boolean;
	dirtyFields?: {
		Enabled?: boolean;
		Providers?: Array<
			Partial<
				Record<string, boolean | { Enabled?: boolean; Webhook?: boolean; UserKey?: boolean; Token?: boolean }>
			>
		>;
	};
	onChange: <K extends keyof AppConfigNotifications>(field: K, value: AppConfigNotifications[K]) => void;
	errorsUpdate?: (errors: Record<string, string>) => void;
}

const PROVIDER_TYPES = ["Discord", "Pushover", "Gotify"] as const;

export const ConfigSectionNotifications: React.FC<ConfigSectionNotificationsProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("");

	// Local select state for adding providers
	const [newProviderType, setNewProviderType] = useState<string>("Discord");

	const providers = React.useMemo(() => (Array.isArray(value.Providers) ? value.Providers : []), [value.Providers]);

	// ----- Validation -----
	const errors = React.useMemo(() => {
		const errs: Record<string, string> = {};

		if (value.Enabled) {
			// At least one enabled provider recommended
			if (!providers.some((p) => p.Enabled)) {
				errs["Providers"] = "Enable at least one provider or disable notifications.";
			}
		}

		providers.forEach((p, idx) => {
			const prefix = `Providers[${idx}]`;
			if (value.Enabled && p.Enabled) {
				if (p.Provider === "Discord") {
					const discord = p.Discord;
					const webhook = (discord?.Webhook || "").trim();
					if (!webhook) {
						errs[`${prefix}.Discord.Webhook`] = "Webhook URL required.";
					} else if (!/^https?:\/\//i.test(webhook)) {
						errs[`${prefix}.Discord.Webhook`] = "Webhook must start with http(s)://";
					}
				} else if (p.Provider === "Pushover") {
					const push = p.Pushover;
					if (!(push?.UserKey || "").trim()) {
						errs[`${prefix}.Pushover.UserKey`] = "User key required.";
					}
					if (!(push?.Token || "").trim()) {
						errs[`${prefix}.Pushover.Token`] = "App token required.";
					}
				} else if (p.Provider === "Gotify") {
					const gotify = p.Gotify;
					const rawURL = (gotify?.URL || "").trim();
					if (!rawURL) {
						errs[`${prefix}.Gotify.URL`] = "URL required.";
					} else {
						const urlErr = ValidateURL(rawURL);
						if (urlErr) errs[`${prefix}.Gotify.URL`] = urlErr;
					}
					if (!(gotify?.Token || "").trim()) {
						errs[`${prefix}.Gotify.Token`] = "App token required.";
					}
				}
			}
		});

		return errs;
	}, [value.Enabled, providers]);

	// Emit errors
	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	// ----- Mutators -----
	const setProviders = (next: AppConfigNotificationProviders[]) => onChange("Providers", next);

	const addProvider = () => {
		if (!editing) return;
		const type = newProviderType;
		let newEntry: AppConfigNotificationProviders;
		if (type === "Discord") {
			newEntry = {
				Provider: "Discord",
				Enabled: true,
				Discord: { Enabled: true, Webhook: "" },
			};
		} else if (type === "Pushover") {
			newEntry = {
				Provider: "Pushover",
				Enabled: true,
				Pushover: { Enabled: true, UserKey: "", Token: "" },
			};
		} else {
			newEntry = {
				Provider: "Gotify",
				Enabled: true,
				Gotify: { Enabled: true, URL: "", Token: "" },
			};
		}
		setProviders([...providers, newEntry]);
	};

	const removeProvider = (idx: number) => {
		if (!editing) return;
		const next = providers.slice();
		next.splice(idx, 1);
		setProviders(next);
	};

	const updateProvider = <K extends keyof AppConfigNotificationProviders>(
		idx: number,
		field: K,
		val: AppConfigNotificationProviders[K]
	) => {
		const next = providers.slice();
		next[idx] = { ...next[idx], [field]: val };
		setProviders(next);
	};

	const updateDiscord = <K extends keyof AppConfigNotificationDiscord>(
		idx: number,
		field: K,
		val: AppConfigNotificationDiscord[K]
	) => {
		const prov = providers[idx];
		if (!prov.Discord) return;
		const next = providers.slice();
		next[idx] = {
			...prov,
			Discord: { ...prov.Discord, [field]: val },
		};
		setProviders(next);
	};

	const updatePushover = <K extends keyof AppConfigNotificationPushover>(
		idx: number,
		field: K,
		val: AppConfigNotificationPushover[K]
	) => {
		const prov = providers[idx];
		if (!prov.Pushover) return;
		const next = providers.slice();
		next[idx] = {
			...prov,
			Pushover: { ...prov.Pushover, [field]: val },
		};
		setProviders(next);
	};

	const updateGotify = <K extends keyof AppConfigNotificationGotify>(
		idx: number,
		field: K,
		val: AppConfigNotificationGotify[K]
	) => {
		const prov = providers[idx];
		if (!prov.Gotify) return;
		const next = providers.slice();
		next[idx] = {
			...prov,
			Gotify: { ...prov.Gotify, [field]: val },
		};
		setProviders(next);
	};

	return (
		<Card className="p-5 space-y-6">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Notifications</h2>
				<Button
					variant="outline"
					size="sm"
					disabled={editing}
					hidden={editing}
					onClick={() => sendTestNotification()}
				>
					Test Notifications
				</Button>
			</div>

			{/* Global Enabled */}
			<div
				className={cn(
					"flex items-center justify-between border rounded-md p-3 transition",
					"border-muted",
					dirtyFields.Enabled && "border-amber-500"
				)}
			>
				<Label>Enabled</Label>
				<div className="flex items-center gap-2">
					<Switch
						disabled={!editing}
						checked={value.Enabled}
						onCheckedChange={(v) => onChange("Enabled", v)}
					/>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-notifications-enabled"
								>
									<HelpCircle className="h-4 w-4" />
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-72 text-xs leading-snug"
							>
								<p className="mb-2">
									Turn on to send events through enabled providers (Discord, Pushover, Gotify).
								</p>
								<p className="text-[10px] text-muted-foreground">
									Each provider must also be enabled individually.
								</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
			</div>

			{/* Providers */}
			<div
				className={cn(
					"space-y-4",
					(errors["Providers"] || dirtyFields.Providers) && "rounded-md",
					errors["Providers"]
						? "border border-red-500 p-3"
						: dirtyFields.Providers && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Providers</Label>
					{editing && (
						<div className="flex items-center gap-2">
							<Select value={newProviderType} onValueChange={(v) => setNewProviderType(v)}>
								<SelectTrigger className="h-8 w-36">
									<SelectValue placeholder="Type" />
								</SelectTrigger>
								<SelectContent>
									{PROVIDER_TYPES.map((p) => (
										<SelectItem key={p} value={p}>
											{p}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
							<Button type="button" variant="outline" size="sm" onClick={addProvider}>
								<Plus className="h-4 w-4 mr-1" />
								Add
							</Button>
						</div>
					)}
				</div>

				{providers.length === 0 && <p className="text-[11px] text-muted-foreground">No providers added.</p>}

				{providers.map((p, idx) => {
					const providerDirty = dirtyFields.Providers?.[idx] as Partial<{
						Discord?: Partial<AppConfigNotificationDiscord>;
						Pushover?: Partial<AppConfigNotificationPushover>;
						Gotify?: Partial<AppConfigNotificationGotify>;
						Enabled?: boolean;
					}>;
					const providerErrorEntries = Object.entries(errors).filter(([k]) =>
						k.startsWith(`Providers[${idx}]`)
					);
					const providerErrors = providerErrorEntries.map(([, msg]) => msg);

					// Field-level helpers (key-based)
					const hasError = (suffix: string) => providerErrorEntries.some(([k]) => k.endsWith(suffix));

					return (
						<div
							key={idx}
							className={cn(
								"space-y-3 rounded-md border p-3 transition",
								providerErrors.length
									? "border-red-500"
									: providerDirty
										? "border-amber-500"
										: "border-muted"
							)}
						>
							<div className="flex items-center justify-between">
								<div className="flex items-center gap-3">
									<p className="font-medium text-sm">{p.Provider}</p>
									<Switch
										disabled={!editing}
										checked={p.Enabled}
										onCheckedChange={(v) => updateProvider(idx, "Enabled", v)}
									/>
								</div>
								{editing && (
									<Button
										variant="ghost"
										size="icon"
										onClick={() => removeProvider(idx)}
										aria-label="help-notifications-remove-provider"
									>
										<Trash2 className="h-4 w-4" />
									</Button>
								)}
							</div>

							{/* Discord Fields */}
							{p.Provider === "Discord" && p.Enabled && (
								<div
									className={cn(
										"space-y-1",
										hasError("Discord.Webhook") && "rounded-md",
										hasError("Discord.Webhook")
											? "border border-red-500 p-3"
											: providerDirty?.Discord?.Webhook && "border border-amber-500 p-3"
									)}
								>
									<div className="flex items-center justify-between">
										<Label>Webhook URL</Label>
										{editing && (
											<Popover>
												<PopoverTrigger asChild>
													<Button
														variant="outline"
														className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
														aria-label="help-notifications-discord-webhook"
													>
														<HelpCircle className="h-4 w-4" />
													</Button>
												</PopoverTrigger>
												<PopoverContent
													side="right"
													align="center"
													sideOffset={8}
													className="w-72 text-xs leading-snug"
												>
													<p className="mb-2 font-medium">Discord Webhook</p>
													<ol className="list-decimal ml-4 space-y-1 text-[11px]">
														<li>Server Settings &gt; Integrations &gt; Webhooks</li>
														<li>Create Webhook (choose channel)</li>
														<li>Copy the webhook URL and paste here</li>
													</ol>
													<p className="mt-2 text-[10px] text-muted-foreground">
														Must begin with https://discord.com/api/webhooks/...
													</p>
												</PopoverContent>
											</Popover>
										)}
									</div>
									<Input
										disabled={!editing}
										placeholder="https://discord.com/api/webhooks/..."
										value={p.Discord?.Webhook || ""}
										onChange={(e) => {
											const val = e.target.value;
											updateDiscord(idx, "Webhook", val);
										}}
										aria-invalid={providerErrors.some((e) => e.includes("Webhook"))}
									/>
									{providerErrorEntries
										.filter(([k]) => k.endsWith("Discord.Webhook"))
										.map(([, msg], i) => (
											<p key={i} className="text-xs text-red-500">
												{msg}
											</p>
										))}
								</div>
							)}

							{/* Pushover Fields */}
							{p.Provider === "Pushover" && p.Enabled && (
								<div className="grid gap-3 md:grid-cols-2">
									<div
										className={cn(
											"space-y-1",
											hasError("Pushover.UserKey") && "rounded-md",
											hasError("Pushover.UserKey")
												? "border border-red-500 p-3"
												: providerDirty?.Pushover?.UserKey && "border border-amber-500 p-3"
										)}
									>
										<div className="flex items-center justify-between">
											<Label>User Key</Label>
											{editing && (
												<Popover>
													<PopoverTrigger asChild>
														<Button
															variant="outline"
															className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
															aria-label="help-notification-pushover-user-key"
														>
															<HelpCircle className="h-4 w-4" />
														</Button>
													</PopoverTrigger>
													<PopoverContent
														side="right"
														align="center"
														sideOffset={8}
														className="w-64 text-xs leading-snug"
													>
														<p className="mb-1 font-medium">Pushover User Key</p>
														<p className="text-[11px] mb-2">
															Found on your Pushover dashboard after logging in.
														</p>
														<p className="text-[10px] text-muted-foreground">
															https://pushover.net/
														</p>
													</PopoverContent>
												</Popover>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="User key"
											value={p.Pushover?.UserKey || ""}
											onChange={(e) => updatePushover(idx, "UserKey", e.target.value)}
											aria-invalid={providerErrors.some((e) => e.includes("UserKey"))}
										/>
										{providerErrorEntries
											.filter(([k]) => k.endsWith("Pushover.UserKey"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
									<div
										className={cn(
											"space-y-1",
											hasError("Pushover.Token") && "rounded-md",
											hasError("Pushover.Token")
												? "border border-red-500 p-3"
												: providerDirty?.Pushover?.Token && "border border-amber-500 p-3"
										)}
									>
										<div className="flex items-center justify-between">
											<Label>App Token</Label>
											{editing && (
												<Popover>
													<PopoverTrigger asChild>
														<Button
															variant="outline"
															className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
															aria-label="help-notifications-pushover-app-token"
														>
															<HelpCircle className="h-4 w-4" />
														</Button>
													</PopoverTrigger>
													<PopoverContent
														side="right"
														align="center"
														sideOffset={8}
														className="w-64 text-xs leading-snug"
													>
														<p className="mb-1 font-medium">Pushover App Token</p>
														<p className="text-[11px] mb-2">
															Create or view under "Your Applications" in Pushover.
														</p>
														<p className="text-[10px] text-muted-foreground">
															Needed to send messages via the API.
														</p>
													</PopoverContent>
												</Popover>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="App token"
											value={p.Pushover?.Token || ""}
											onChange={(e) => updatePushover(idx, "Token", e.target.value)}
											aria-invalid={providerErrors.some((e) => e.includes("Token"))}
										/>
										{providerErrorEntries
											.filter(([k]) => k.endsWith("Pushover.Token"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
								</div>
							)}

							{/* Gotify Fields */}
							{p.Provider === "Gotify" && p.Enabled && (
								<div className="grid gap-3 md:grid-cols-2">
									<div
										className={cn(
											"space-y-1",
											hasError("Gotify.URL") && "rounded-md",
											hasError("Gotify.URL")
												? "border border-red-500 p-3"
												: providerDirty?.Gotify?.URL && "border border-amber-500 p-3"
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
															aria-label="help-notifications-gotify-url"
														>
															<HelpCircle className="h-4 w-4" />
														</Button>
													</PopoverTrigger>
													<PopoverContent
														side="right"
														align="center"
														sideOffset={8}
														className="w-64 text-xs leading-snug"
													>
														<p className="mb-1 font-medium">Gotify URL</p>
														<p className="text-[11px] mb-2">
															The base URL of your Gotify server. Domains may omit port.
															IPv4 addresses must include a port. Example:
															https://gotify.domain.com, http://192.168.1.10:8070 or
															http://gotify:8070
														</p>
													</PopoverContent>
												</Popover>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="URL"
											value={p.Gotify?.URL || ""}
											onChange={(e) => updateGotify(idx, "URL", e.target.value)}
											aria-invalid={providerErrors.some((e) => e.includes("URL"))}
										/>
										{providerErrorEntries
											.filter(([k]) => k.endsWith("Gotify.URL"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
									<div
										className={cn(
											"space-y-1",
											hasError("Gotify.Token") && "rounded-md",
											hasError("Gotify.Token")
												? "border border-red-500 p-3"
												: providerDirty?.Gotify?.Token && "border border-amber-500 p-3"
										)}
									>
										<div className="flex items-center justify-between">
											<Label>App Token</Label>
											{editing && (
												<Popover>
													<PopoverTrigger asChild>
														<Button
															variant="outline"
															className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
															aria-label="help-notifications-pushover-app-token"
														>
															<HelpCircle className="h-4 w-4" />
														</Button>
													</PopoverTrigger>
													<PopoverContent
														side="right"
														align="center"
														sideOffset={8}
														className="w-64 text-xs leading-snug"
													>
														<p className="mb-1 font-medium">Gotify App Token</p>
														<p className="text-[11px] mb-2">
															Create or view under "Apps" in Gotify.
														</p>
														<p className="text-[10px] text-muted-foreground">
															Needed to send messages via the API.
														</p>
													</PopoverContent>
												</Popover>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="App token"
											value={p.Gotify?.Token || ""}
											onChange={(e) => updateGotify(idx, "Token", e.target.value)}
											aria-invalid={providerErrors.some((e) => e.includes("Token"))}
										/>
										{providerErrorEntries
											.filter(([k]) => k.endsWith("Gotify.Token"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
								</div>
							)}

							{providerErrors.length > 0 && (
								<p className="text-[10px] text-red-500">Resolve provider errors above.</p>
							)}
						</div>
					);
				})}

				{errors["Providers"] && <p className="text-xs text-red-500">{errors["Providers"]}</p>}
			</div>
		</Card>
	);
};
