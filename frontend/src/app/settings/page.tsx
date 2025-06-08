"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import ErrorMessage from "@/components/ui/error-message";
import { Input } from "@/components/ui/input";
import Loader from "@/components/ui/loader";
import {
	Popover,
	PopoverTrigger,
	PopoverContent,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import {
	fetchConfig,
	postClearTempImagesFolder,
} from "@/services/api.settings";
import { AppConfig } from "@/types/config";
import { useRouter } from "next/navigation";
import React, { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

const SettingsPage: React.FC = () => {
	const router = useRouter();
	const isMounted = useRef(false);
	const [config, setConfig] = useState<AppConfig | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	// Fetch configuration data
	useEffect(() => {
		if (isMounted.current) return;
		isMounted.current = true;
		if (typeof window === "undefined") {
			document.title = "Aura | Settings";
		}

		const fetchConfigFromAPI = async () => {
			try {
				const resp = await fetchConfig();
				if (!resp) {
					throw new Error("No response from API");
				}
				if (resp.status !== "success") {
					throw new Error(resp.message);
				}
				const appConfig = resp.data;
				if (!appConfig) {
					throw new Error("No config found in response");
				}
				setConfig(appConfig);
				setError(null);
				setLoading(false);
			} catch (error) {
				setConfig(null);
				setError(
					error instanceof Error ? error.message : String(error)
				);
			} finally {
				setLoading(false);
			}
		};
		fetchConfigFromAPI();
	}, []);

	const handleViewLogs = () => {
		router.push("/logs");
	};

	const clearTempImagesFolder = async () => {
		try {
			const clearTempResp = await postClearTempImagesFolder();
			if (!clearTempResp) {
				throw new Error("No response from API");
			}
			toast.success("Temp images folder cleared successfully");
		} catch (error) {
			toast.error(error instanceof Error ? error.message : String(error));
		}
	};

	const parseCronToHumanReadable = (cronExpression: string): string => {
		const parts = cronExpression.split(" ");
		if (parts.length !== 5) return "Current cron expression is invalid";

		const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

		const minuteText =
			minute === "0" ? "at the start of" : `at minute ${minute}`;
		const hourText = hour === "*" ? "every hour" : `hour ${hour}`;
		const dayOfMonthText =
			dayOfMonth === "*" ? "every day" : `on day ${dayOfMonth}`;
		const monthText = month === "*" ? "every month" : `in month ${month}`;
		const dayOfWeekText =
			dayOfWeek === "*"
				? "every day of the week"
				: `on day ${dayOfWeek} of the week`;

		return `Currently runs ${minuteText} ${hourText}, ${dayOfMonthText}, ${monthText}, and ${dayOfWeekText}.`;
	};

	const AppConfig = {
		"Media Server": {
			Title: "Media Server Information",
			Fields: [
				{
					Label: "Server Type",
					Value: config?.MediaServer.Type,
					Tooltip:
						"The type of media server (e.g., Plex, Emby, Jellyfin).",
				},
				{
					Label: "Server URL",
					Value: config?.MediaServer.URL,
					Tooltip: "The base URL of the media server.",
				},
				{
					Label: "Authentication Token",
					Value: config?.MediaServer.Token,
					Tooltip:
						"The authentication token for accessing the media server.",
				},
				{
					Label: "Libraries",
					Value: config?.MediaServer.Libraries.map(
						(library) => library.Name
					).join(", "),
					Tooltip: "",
				},
			],
		},
		"Other API": {
			Title: "Other API Information",
			Fields: [
				{
					Label: "TMDB API Key",
					Value: config?.TMDB.ApiKey,
					Tooltip:
						"The API key for accessing TMDB services. This is not used in this version.",
				},
				{
					Label: "Mediux Token",
					Value: config?.Mediux.Token,
					Tooltip:
						"The authentication token for accessing Mediux services.",
				},
			],
		},
		"Other Settings": {
			Title: "Other Settings",
			Fields: [
				{
					Label: "Cache Images",
					Value: config?.CacheImages ? "Enabled" : "Disabled",
					Tooltip: "Whether to cache images locally.",
				},
				{
					Label: "Save Images Next to Content",
					Value: config?.SaveImageNextToContent
						? "Enabled"
						: "Disabled",
					Tooltip:
						"Whether to save images next to the associated content.",
				},
				{
					Label: "Auto Download",
					Value: config?.AutoDownload?.Enabled
						? "Enabled"
						: "Disabled",
					Tooltip: "Whether auto-download is enabled.",
				},
				{
					Label: "Auto Download Cron",
					Value: config?.AutoDownload?.Cron,
					Tooltip: parseCronToHumanReadable(
						config?.AutoDownload?.Cron || ""
					),
				},
			],
		},
		Kometa: {
			Title: "Kometa",
			Fields: [
				{
					Label: "Remove Labels",
					Value: config?.Kometa?.RemoveLabels
						? "Enabled"
						: "Disabled",
					Tooltip: "Whether to remove labels from media items.",
				},
				{
					Label: "Labels",
					Value:
						Array.isArray(config?.Kometa?.Labels) &&
						config?.Kometa?.Labels.length > 1
							? config?.Kometa?.Labels.join(", ")
							: config?.Kometa?.Labels?.[0] || "",
					Tooltip: "The list of labels to apply to media items.",
				},
			],
		},
		Notifications: {
			Title: "Notifications",
			Fields: [
				{
					Label: "Provider",
					Value: config?.Notification?.Provider,
					Tooltip:
						"The provider for notifications. Currently, the only provider is Discord.",
				},
				{
					Label: "Webhook",
					Value: config?.Notification?.Webhook,
					Tooltip:
						"he webhook URL for the provider. This can be obtained by creating a webhook in Discord.",
				},
			],
		},
		Logging: {
			Title: "Logging",
			Fields: [
				{
					Label: "Logging Level",
					Value: config?.Logging?.Level,
					Tooltip:
						"The logging level (e.g., DEBUG, INFO, WARN, ERROR).",
				},
				{
					Label: "Log File Path",
					Value: config?.Logging?.File,
					Tooltip: "The file path where logs are stored.",
				},
			],
			Buttons: [{ Label: "View Logs", onClick: handleViewLogs }],
		},
		"Admin Tools": {
			Title: "Admin Tools",
			Buttons: [
				{
					Label: "Clear Temp Images Folder",
					onClick: clearTempImagesFolder,
				},
			],
		},
	};

	return (
		<div className="container mx-auto p-6">
			{/* Conditional Rendering */}
			{loading ? (
				<Loader message="Loading configuration..." />
			) : error ? (
				<ErrorMessage message={error} />
			) : (
				<>
					<h1 className="text-3xl font-bold mb-4">Settings</h1>
					{AppConfig &&
						Object.entries(AppConfig).map(([key, value]) => {
							// Filter out fields with empty values (after trimming)
							const fieldsToShow =
								"Fields" in value
									? value.Fields.filter(
											(field) =>
												field.Value &&
												field.Value.toString().trim() !==
													""
									  )
									: [];
							// Check if there are any buttons
							const hasButtons =
								"Buttons" in value &&
								(
									value.Buttons as {
										onClick: () => void;
										Label: string;
									}[]
								).length > 0;

							// If there are no non-empty fields and no buttons, skip the card
							if (!fieldsToShow.length && !hasButtons)
								return null;

							return (
								<Card className="mb-4" key={key}>
									<CardHeader>
										<h2 className="text-xl font-semibold">
											{value.Title}
										</h2>
									</CardHeader>
									<CardContent>
										<div className="space-y-2">
											{"Fields" in value &&
												fieldsToShow.map(
													(field, index) => (
														<div key={index}>
															<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
																{field.Label}
															</label>
															{field.Label ===
																"Libraries" ||
															field.Label ===
																"Labels" ? (
																<>
																	{(
																		field.Value ??
																		""
																	)
																		.split(
																			", "
																		)
																		.filter(
																			(
																				item: string
																			) =>
																				item.trim() !==
																				""
																		)
																		.map(
																			(
																				item: string,
																				idx: number
																			) => (
																				<Badge
																					key={
																						idx
																					}
																					className="mr-2 mt-1 text-sm"
																				>
																					{
																						item
																					}
																				</Badge>
																			)
																		)}
																</>
															) : (
																<div className="flex items-center gap-2 mt-1">
																	<Input
																		value={
																			field.Value
																		}
																		disabled
																		className="w-full"
																	/>
																	<Popover>
																		<PopoverTrigger className="cursor-pointer">
																			<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
																				?
																			</span>
																		</PopoverTrigger>
																		<PopoverContent className="w-60">
																			{
																				field.Tooltip
																			}
																		</PopoverContent>
																	</Popover>
																</div>
															)}
														</div>
													)
												)}
											{"Buttons" in value &&
												(
													value.Buttons as {
														onClick: () => void;
														Label: string;
													}[]
												).map((button, index) => (
													<Button
														key={index}
														onClick={button.onClick}
														className="w-full"
													>
														{button.Label}
													</Button>
												))}
										</div>
									</CardContent>
								</Card>
							);
						})}
				</>
			)}
		</div>
	);
};

export default SettingsPage;
