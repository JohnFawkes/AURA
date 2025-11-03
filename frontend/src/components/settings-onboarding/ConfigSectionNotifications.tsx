"use client";

import { ValidateURL } from "@/helper/validation/validate-url";
import { sendTestNotification } from "@/services/settings-onboarding/api-notifications-test";
import { Plus, TestTube, Trash2 } from "lucide-react";

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { GetConnectionColor } from "@/components/settings-onboarding/ConfigSectionMediaServer";
import {
	CONNECTION_STATUS_COLORS_BG,
	ConfigConnectionStatus,
} from "@/components/settings-onboarding/ConfigSectionSonarrRadarr";
import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import {
	AppConfigNotificationDiscord,
	AppConfigNotificationGotify,
	AppConfigNotificationProviders,
	AppConfigNotificationPushover,
	AppConfigNotificationWebhook,
	AppConfigNotifications,
} from "@/types/config/config-app";

interface ConfigSectionNotificationsProps {
	value: AppConfigNotifications;
	editing: boolean;
	dirtyFields?: {
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
	onChange: <K extends keyof AppConfigNotifications>(field: K, value: AppConfigNotifications[K]) => void;
	errorsUpdate?: (errors: Record<string, string>) => void;
	configAlreadyLoaded: boolean;
}

const PROVIDER_TYPES = ["Discord", "Pushover", "Gotify", "Webhook"] as const;

export const ConfigSectionNotifications: React.FC<ConfigSectionNotificationsProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
	configAlreadyLoaded,
}) => {
	const prevErrorsRef = useRef<string>("");
	const hasRunInitialValidation = useRef(false);

	// Local select state for adding providers
	const [newProviderType, setNewProviderType] = useState<string>("Discord");

	// State to track editing header keys
	const [editingHeaderKeys, setEditingHeaderKeys] = useState<Record<number, Record<string, string>>>({});

	const providers = useMemo(() => (Array.isArray(value.Providers) ? value.Providers : []), [value.Providers]);

	// State to track app connection testing
	const [appConnectionStatus, setAppConnectionStatus] = useState<Record<number, ConfigConnectionStatus>>({});
	const [remoteTokenErrors, setRemoteTokenErrors] = useState<Record<number, string | null>>({});

	// ----- Validation -----
	const errors = useMemo(() => {
		const errs: Record<string, string> = {};

		if (value.Enabled) {
			// At least one enabled provider recommended
			if (!providers.some((p) => p.Enabled)) {
				errs["Providers"] = "Enable at least one provider or disable notifications.";
			}
		}

		providers.forEach((p, idx) => {
			if (value.Enabled && p.Enabled) {
				if (p.Provider === "Discord") {
					const discord = p.Discord;
					const webhook = (discord?.Webhook || "").trim();
					if (!webhook) {
						errs[`Providers.[${idx}].Discord.Webhook`] = "Webhook URL required.";
					} else if (!/^https?:\/\//i.test(webhook)) {
						errs[`Providers.[${idx}].Discord.Webhook`] = "Webhook must start with http(s)://";
					}
					if (remoteTokenErrors[idx]) {
						errs[`Providers.[${idx}].Discord.Webhook`] = remoteTokenErrors[idx] || "Connection failed.";
					}
				} else if (p.Provider === "Pushover") {
					const push = p.Pushover;
					if (!(push?.UserKey || "").trim()) {
						errs[`Providers.[${idx}].Pushover.UserKey`] = "User key required.";
					}
					if (!(push?.Token || "").trim()) {
						errs[`Providers.[${idx}].Pushover.Token`] = "App token required.";
					}
					if (remoteTokenErrors[idx]) {
						errs[`Providers.[${idx}].Pushover.Token`] = remoteTokenErrors[idx] || "Connection failed.";
					}
				} else if (p.Provider === "Gotify") {
					const gotify = p.Gotify;
					const rawURL = (gotify?.URL || "").trim();
					if (!rawURL) {
						errs[`Providers.[${idx}].Gotify.URL`] = "URL required.";
					} else {
						const urlErr = ValidateURL(rawURL);
						if (urlErr) errs[`Providers.[${idx}].Gotify.URL`] = urlErr;
					}
					if (!(gotify?.Token || "").trim()) {
						errs[`Providers.[${idx}].Gotify.Token`] = "App token required.";
					}
					if (remoteTokenErrors[idx]) {
						errs[`Providers.[${idx}].Gotify.Token`] = remoteTokenErrors[idx] || "Connection failed.";
					}
				} else if (p.Provider === "Webhook") {
					const webhook = p.Webhook;
					const rawURL = (webhook?.URL || "").trim();
					if (!rawURL) {
						errs[`Providers.[${idx}].Webhook.URL`] = "URL required.";
					} else {
						const urlErr = ValidateURL(rawURL);
						if (urlErr) errs[`Providers.[${idx}].Webhook.URL`] = urlErr;
					}
				}
			}
		});

		return errs;
	}, [value.Enabled, providers, remoteTokenErrors]);

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
		} else if (type === "Gotify") {
			newEntry = {
				Provider: "Gotify",
				Enabled: true,
				Gotify: { Enabled: true, URL: "", Token: "" },
			};
		} else if (type === "Webhook") {
			newEntry = {
				Provider: "Webhook",
				Enabled: true,
				Webhook: { Enabled: true, URL: "", Headers: {} },
			};
		} else {
			return;
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

	const updateWebhook = <K extends keyof AppConfigNotificationWebhook>(
		idx: number,
		field: K,
		val: AppConfigNotificationWebhook[K]
	) => {
		const prov = providers[idx];
		if (!prov.Webhook) return;
		const next = providers.slice();
		next[idx] = {
			...prov,
			Webhook: { ...prov.Webhook, [field]: val },
		};
		setProviders(next);
	};

	const runRemoteValidation = useCallback(
		async (idx: number, showToast = true) => {
			const provider = providers[idx];
			if (!provider || !provider.Enabled) return;

			// Set to unknown while testing
			setAppConnectionStatus((s) => ({
				...s,
				[idx]: { status: "unknown", color: GetConnectionColor("unknown") },
			}));

			try {
				const start = Date.now();
				const { ok, message } = await sendTestNotification(provider, showToast);
				const elapsed = Date.now() - start;
				const minDelay = 400;

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
		[providers]
	);

	useEffect(() => {
		if (configAlreadyLoaded && !hasRunInitialValidation.current) {
			// Run remote validation for all apps that have URL and APIKey set
			providers.forEach((p, idx) => {
				if (p.Enabled) {
					if (p.Provider === "Discord" && p.Discord?.Webhook) {
						setTimeout(() => runRemoteValidation(idx, false), 200 * idx);
					} else if (p.Provider === "Pushover" && p.Pushover?.UserKey && p.Pushover?.Token) {
						setTimeout(() => runRemoteValidation(idx, false), 200 * idx);
					} else if (p.Provider === "Gotify" && p.Gotify?.URL && p.Gotify?.Token) {
						setTimeout(() => runRemoteValidation(idx, false), 200 * idx);
					} else if (p.Provider === "Webhook" && p.Webhook?.URL) {
						setTimeout(() => runRemoteValidation(idx, false), 200 * idx);
					}
				}
			});
			hasRunInitialValidation.current = true;
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [configAlreadyLoaded, runRemoteValidation]);

	return (
		<Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold text-blue-500">Notifications</h2>
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
						<PopoverHelp ariaLabel="help-notifications-enabled">
							<p className="mb-2">
								Turn on to send events through enabled providers (Discord, Pushover, Gotify, Custom
								Webhook).
							</p>
							<p className="text-muted-foreground">Each provider can also be enabled individually.</p>
						</PopoverHelp>
					)}
				</div>
			</div>

			{/* Providers */}
			<div className={cn("space-y-4")}>
				<div className="flex items-center justify-end">
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

				{providers.length === 0 && (
					<p className="text-sm text-muted-foreground">No notification providers added.</p>
				)}

				{providers.map((p, idx) => {
					const providerDirty = dirtyFields.Providers?.[idx];
					const discordDirty =
						typeof providerDirty?.Discord === "object" && providerDirty?.Discord !== null
							? providerDirty.Discord
							: {};
					const pushoverDirty =
						typeof providerDirty?.Pushover === "object" && providerDirty?.Pushover !== null
							? providerDirty.Pushover
							: {};
					const gotifyDirty =
						typeof providerDirty?.Gotify === "object" && providerDirty?.Gotify !== null
							? providerDirty.Gotify
							: {};
					const webhookDirty =
						typeof providerDirty?.Webhook === "object" && providerDirty?.Webhook !== null
							? providerDirty.Webhook
							: {};
					const providerErrorEntries = Object.entries(errors).filter(([k]) =>
						k.startsWith(`Providers.[${idx}]`)
					);
					const providerErrors = providerErrorEntries.map(([, msg]) => msg);
					const connStatus = appConnectionStatus[idx] || {
						status: "unknown",
						color: "gray-500",
					};
					return (
						<div
							key={idx}
							className={cn(
								"space-y-3 rounded-md border p-3 transition",
								providerErrors.length ? "border-red-500" : "border-muted"
							)}
						>
							<div className="flex items-center justify-between">
								<div
									className={cn(
										"flex items-center gap-3",
										providerDirty?.Enabled && "rounded-md border border-amber-500"
									)}
								>
									<h2 className={`text-xl font-semibold text-${connStatus.color}`}>{p.Provider}</h2>
									<span
										className={`h-2 w-2 rounded-full ${CONNECTION_STATUS_COLORS_BG[connStatus.status]} animate-pulse`}
										title={`Connection status: ${connStatus.status}`}
									/>

									<Switch
										disabled={!editing}
										checked={p.Enabled}
										onCheckedChange={(v) => updateProvider(idx, "Enabled", v)}
									/>
								</div>
								<div className="flex items-center gap-2">
									<Button
										variant="outline"
										size="sm"
										disabled={
											!p.Enabled ||
											(p.Provider === "Discord" && !p.Discord?.Webhook) ||
											(p.Provider === "Pushover" &&
												(!p.Pushover?.UserKey || !p.Pushover?.Token)) ||
											(p.Provider === "Gotify" && (!p.Gotify?.URL || !p.Gotify?.Token)) ||
											(p.Provider === "Webhook" && !p.Webhook?.URL)
										}
										hidden={
											((p.Provider === "Discord" && !p.Discord?.Webhook) ||
												(p.Provider === "Pushover" &&
													(!p.Pushover?.UserKey || !p.Pushover?.Token)) ||
												(p.Provider === "Gotify" && (!p.Gotify?.URL || !p.Gotify?.Token)) ||
												(p.Provider === "Webhook" && !p.Webhook?.URL)) &&
											!editing
										}
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
											onClick={() => removeProvider(idx)}
											aria-label="help-apps-remove-app"
											className="bg-red-700"
										>
											<Trash2 className="h-4 w-4" />
										</Button>
									)}
								</div>
							</div>

							{/* Discord Fields */}
							{p.Provider === "Discord" && p.Enabled && (
								<div className={cn("space-y-1")}>
									<div className="flex items-center justify-between">
										<Label>Webhook URL</Label>
										{editing && (
											<PopoverHelp ariaLabel="help-notifications-discord-webhook">
												<p className="mb-2 font-medium">Discord Webhook</p>
												<ol className="list-decimal ml-4 space-y-1 ">
													<li>Server Settings &gt; Integrations &gt; Webhooks</li>
													<li>Create Webhook (choose channel)</li>
													<li>Copy the webhook URL and paste here</li>
												</ol>
												<p className="mt-2 text-muted-foreground">
													Must begin with https://discord.com/api/webhooks/...
												</p>
											</PopoverHelp>
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
										className={cn(discordDirty?.Webhook && "border border-amber-500 p-3")}
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
									<div className={cn("space-y-1")}>
										<div className="flex items-center justify-between">
											<Label>User Key</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-notification-pushover-user-key">
													<p className="mb-1 font-medium">Pushover User Key</p>
													<p className=" mb-2">
														Found on your Pushover dashboard after logging in.
													</p>
													<p className="text-muted-foreground">https://pushover.net/</p>
												</PopoverHelp>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="User key"
											value={p.Pushover?.UserKey || ""}
											onChange={(e) => updatePushover(idx, "UserKey", e.target.value)}
											className={cn(pushoverDirty?.UserKey && "border border-amber-500 p-3")}
										/>
										{providerErrorEntries
											.filter(([k]) => k.endsWith("Pushover.UserKey"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
									<div className={cn("space-y-1")}>
										<div className="flex items-center justify-between">
											<Label>App Token</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-notifications-pushover-app-token">
													<p className="mb-1 font-medium">Pushover App Token</p>
													<p className=" mb-2">
														Create or view under "Your Applications" in Pushover.
													</p>
													<p className="text-muted-foreground">
														Needed to send messages via the API.
													</p>
												</PopoverHelp>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="App token"
											value={p.Pushover?.Token || ""}
											onChange={(e) => updatePushover(idx, "Token", e.target.value)}
											className={cn(pushoverDirty?.Token && "border border-amber-500 p-3")}
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
									<div className={cn("space-y-1")}>
										<div className="flex items-center justify-between">
											<Label>URL</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-notifications-gotify-url">
													<p className="mb-1 font-medium">Gotify URL</p>
													<p className=" mb-2">
														The base URL of your Gotify server. Domains may omit port. IPv4
														addresses must include a port
													</p>
													<p>Exmaples:</p>
													<ul className="list-disc ml-4 space-y-1 text-muted-foreground">
														<li>https://gotify.domain.com</li>
														<li>http://192.168.1.10:8070</li>
														<li>http://gotify:8070</li>
													</ul>
												</PopoverHelp>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="URL"
											value={p.Gotify?.URL || ""}
											onChange={(e) => updateGotify(idx, "URL", e.target.value)}
											className={cn(gotifyDirty?.URL && "border border-amber-500 p-3")}
										/>
										{providerErrorEntries
											.filter(([k]) => k.endsWith("Gotify.URL"))
											.map(([, msg], i) => (
												<p key={i} className="text-xs text-red-500">
													{msg}
												</p>
											))}
									</div>
									<div className={cn("space-y-1")}>
										<div className="flex items-center justify-between">
											<Label>App Token</Label>
											{editing && (
												<PopoverHelp ariaLabel="help-notifications-gotify-app-token">
													<p className="mb-1 font-medium">
														<span className="font-mono">Gotify App Token</span>
													</p>
													<p className=" mb-2">
														Generate or view your app token under{" "}
														<span className="font-semibold">Apps</span> in Gotify.
													</p>
													<p className=" text-muted-foreground">
														This token is required to send messages via the API.
													</p>
												</PopoverHelp>
											)}
										</div>
										<Input
											disabled={!editing}
											placeholder="App token"
											value={p.Gotify?.Token || ""}
											onChange={(e) => updateGotify(idx, "Token", e.target.value)}
											className={cn(gotifyDirty?.Token && "border border-amber-500 p-3")}
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

							{/* Webhook Fields */}
							{p.Provider === "Webhook" && p.Enabled && (
								<div className={cn("space-y-1")}>
									<div className="flex items-center justify-between">
										<Label>URL</Label>
										{editing && (
											<PopoverHelp ariaLabel="help-notifications-webhook-url">
												<p className="mb-2 font-medium">Custom Webhook URL</p>
												<p className="text-muted-foreground">
													The URL to send POST requests to for notifications.
												</p>
											</PopoverHelp>
										)}
									</div>
									<Input
										disabled={!editing}
										placeholder="https://example.com/webhook"
										value={p.Webhook?.URL || ""}
										onChange={(e) => {
											const val = e.target.value;
											updateWebhook(idx, "URL", val);
										}}
										className={cn(webhookDirty?.URL && "border border-amber-500 p-3")}
									/>
									{providerErrorEntries
										.filter(([k]) => k.endsWith("Webhook.URL"))
										.map(([, msg], i) => (
											<p key={i} className="text-xs text-red-500">
												{msg}
											</p>
										))}

									{/* Headers Input */}
									<div className="flex items-center justify-between mt-2">
										<Label>Custom Headers</Label>
										{editing && (
											<PopoverHelp ariaLabel="help-notifications-webhook-headers">
												<p className="mb-2 font-medium">Custom Headers</p>
												<p className="text-muted-foreground">
													Add any custom headers to include in the webhook POST request. Enter
													as key/value pairs.
												</p>
											</PopoverHelp>
										)}
									</div>
									<div className="space-y-2">
										{Object.entries(p.Webhook?.Headers || {}).map(([key, value], i) => (
											<div key={key + i} className="flex gap-2 items-center justify-between">
												<Input
													disabled={!editing}
													placeholder="Header Name"
													value={
														editingHeaderKeys[idx]?.[key] !== undefined
															? editingHeaderKeys[idx][key]
															: key
													}
													onChange={(e) => {
														const val = e.target.value;
														setEditingHeaderKeys((prev) => ({
															...prev,
															[idx]: {
																...(prev[idx] || {}),
																[key]: val,
															},
														}));
													}}
													onBlur={(e) => {
														const rawKey = e.target.value.trim();
														const newKey = rawKey.replace(/\s+/g, "_");
														if (newKey && newKey !== key) {
															const headers = { ...(p.Webhook?.Headers || {}) };
															headers[newKey] = headers[key];
															delete headers[key];
															updateWebhook(idx, "Headers", headers);
															// Clean up local state for this key
															setEditingHeaderKeys((prev) => {
																const next = { ...(prev[idx] || {}) };
																delete next[key];
																return { ...prev, [idx]: next };
															});
														} else {
															// Clean up local state if unchanged
															setEditingHeaderKeys((prev) => {
																const next = { ...(prev[idx] || {}) };
																delete next[key];
																return { ...prev, [idx]: next };
															});
														}
													}}
													className={cn(
														"w-1/2",
														webhookDirty?.Headers?.[key] && "border border-amber-500 p-3"
													)}
												/>
												<Input
													disabled={!editing}
													placeholder="Header Value"
													value={value}
													onChange={(e) => {
														const headers = { ...(p.Webhook?.Headers || {}) };
														headers[key] = e.target.value;
														updateWebhook(idx, "Headers", headers);
													}}
													className={cn(
														"w-1/2",
														webhookDirty?.Headers?.[key] && "border border-amber-500 p-3"
													)}
												/>
												{editing && (
													<Button
														variant="ghost"
														size="icon"
														onClick={() => {
															const headers = { ...(p.Webhook?.Headers || {}) };
															delete headers[key];
															updateWebhook(idx, "Headers", headers);
														}}
														aria-label="Remove header"
														className="bg-red-700"
													>
														<Trash2 className="h-4 w-4" />
													</Button>
												)}
											</div>
										))}
										{editing && (
											<Button
												type="button"
												variant="outline"
												size="sm"
												onClick={() => {
													const headers = { ...(p.Webhook?.Headers || {}) };
													let i = 1;
													let newKey = "Header";
													while (headers[newKey + i]) i++;
													headers[newKey + i] = "";
													updateWebhook(idx, "Headers", headers);
												}}
											>
												<Plus className="h-4 w-4 mr-1" />
												Add Header
											</Button>
										)}
									</div>
								</div>
							)}
						</div>
					);
				})}

				{errors["Providers"] && <p className="text-xs text-red-500">{errors["Providers"]}</p>}
			</div>
		</Card>
	);
};
