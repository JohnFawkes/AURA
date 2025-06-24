"use client";

import {
	fetchConfig,
	fetchMediaServerConnectionStatus,
	postClearOldLogs,
	postClearTempImagesFolder,
	postSendTestNotification,
} from "@/services/api.settings";
import { ReturnErrorMessage } from "@/services/api.shared";
import { CircleX, HeartPulseIcon, Logs } from "lucide-react";
import { toast } from "sonner";

import React, { useEffect, useRef, useState } from "react";
import { FaDiscord } from "react-icons/fa";

import { useRouter } from "next/navigation";

import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

import { APIResponse } from "@/types/apiResponse";
import { AppConfig } from "@/types/config";

type LOGGING_VALUES = "DEBUG" | "INFO" | "WARN" | "ERROR";
const LOGGING_OPTIONS: LOGGING_VALUES[] = ["DEBUG", "INFO", "WARN", "ERROR"];

const SettingsPage: React.FC = () => {
	const router = useRouter();
	const isMounted = useRef(false);
	const [config, setConfig] = useState<AppConfig | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	// State - Debug Mode
	const [debugEnabled, setDebugEnabled] = useState(false);

	useEffect(() => {
		const savedMode = localStorage.getItem("debugMode") === "true";
		setDebugEnabled(savedMode);
	}, []);

	const toggleDebugMode = (checked: boolean) => {
		setDebugEnabled(checked);
		localStorage.setItem("debugMode", checked.toString());
	};

	// Fetch configuration data
	useEffect(() => {
		if (typeof window !== "undefined") {
			document.title = "Aura | Settings";
		}
		if (isMounted.current) return;
		isMounted.current = true;

		const fetchConfigFromAPI = async () => {
			try {
				setLoading(true);
				const response = await fetchConfig();

				if (response.status === "error") {
					setError(response);
					setConfig(null);
					return;
				}
				setConfig(response.data ?? null);
				setError(null);
			} catch (error) {
				setError(ReturnErrorMessage<AppConfig>(error));
				setConfig(null);
			} finally {
				setLoading(false);
			}
		};

		fetchConfigFromAPI();
	}, []);

	const handleViewLogs = () => {
		router.push("/logs");
	};

	const handleClearOldLogs = async () => {
		try {
			const response = await postClearOldLogs();

			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to clear old logs");
				return;
			}

			toast.success(response.data || "Successfully cleared old logs");
		} catch (error) {
			const errorResponse = ReturnErrorMessage<void>(error);
			toast.error(errorResponse.error?.Message || "An unexpected error occurred");
		}
	};

	const clearTempImagesFolder = async () => {
		try {
			const response = await postClearTempImagesFolder();

			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to clear temp images folder");
				return;
			}

			toast.success(response.data || "Temp images folder cleared successfully");
		} catch (error) {
			const errorResponse = ReturnErrorMessage<void>(error);
			toast.error(errorResponse.error?.Message || "An unexpected error occurred");
		}
	};

	const checkMediaServerConnectionStatus = async () => {
		try {
			const response = await fetchMediaServerConnectionStatus();

			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to check media server status");
				return;
			}

			toast.success(`Running with version: ${response.data}`);
		} catch (error) {
			const errorResponse = ReturnErrorMessage<string>(error);
			toast.error(errorResponse.error?.Message || "Failed to check media server status");
		}
	};

	const sendTestNotification = async () => {
		try {
			const response = await postSendTestNotification();

			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to send test notification");
				return;
			}

			if (!response.data) {
				toast.error("No response from notification service");
				return;
			}

			toast.success("Test notification sent successfully. Check your Discord channel.");
		} catch (error) {
			const errorResponse = ReturnErrorMessage<string>(error);
			toast.error(errorResponse.error?.Message || "Failed to send test notification");
		}
	};

	const parseCronToHumanReadable = (cronExpression: string): string => {
		const parts = cronExpression.split(" ");
		if (parts.length !== 5) return "Current cron expression is invalid";

		const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

		const minuteText = minute === "0" ? "at the start of" : `at minute ${minute}`;
		const hourText = hour === "*" ? "every hour" : `hour ${hour}`;
		const dayOfMonthText = dayOfMonth === "*" ? "every day" : `on day ${dayOfMonth}`;
		const monthText = month === "*" ? "every month" : `in month ${month}`;
		const dayOfWeekText = dayOfWeek === "*" ? "every day of the week" : `on day ${dayOfWeek} of the week`;

		return `Currently runs ${minuteText} ${hourText}, ${dayOfMonthText}, ${monthText}, and ${dayOfWeekText}.`;
	};

	const AppConfig = {
		"Media Server": {
			Title: "Media Server Information",
			Fields: [
				{
					Label: "Server Type",
					Value: config?.MediaServer.Type,
					Tooltip: "The type of media server (e.g., Plex, Emby, Jellyfin).",
					Editable: true,
					EditType: "select",
					EditOptions: ["Plex", "Emby", "Jellyfin"],
				},
				{
					Label: "Server URL",
					Value: config?.MediaServer.URL,
					Tooltip: "The base URL of the media server.",
					Editable: true,
					EditType: "text",
				},
				{
					Label: "Authentication Token",
					Value: config?.MediaServer.Token,
					Tooltip: "The authentication token for accessing the media server.",
					Editable: true,
					EditType: "text",
				},
				{
					Label: "Libraries",
					Value: config?.MediaServer.Libraries.map((library) => library.Name).join(", "),
					Tooltip: "",
					Editable: true,
					EditType: "text",
				},
			],
			Buttons: [
				{
					Label: "Check Connection Status",
					Icon: <HeartPulseIcon />,
					onClick: checkMediaServerConnectionStatus,
				},
			],
		},
		"MediUX Settings": {
			Title: "MediUX Settings",
			Fields: [
				{
					Label: "Mediux Token",
					Value: config?.Mediux.Token,
					Tooltip: "The authentication token for accessing Mediux services.",
					Editable: true,
					EditType: "text",
				},
				{
					Label: "Download Quality",
					Value: config?.Mediux.DownloadQuality,
					Tooltip: "The quality of media to download (e.g., original, optimized).",
					Editable: true,
					EditType: "select",
					EditOptions: ["original", "optimized"],
				},
			],
		},
		"Other API": {
			Title: "Other API Information",
			Fields: [
				{
					Label: "TMDB API Key",
					Value: config?.TMDB.ApiKey,
					Tooltip: "The API key for accessing TMDB services. This is not used in this version.",
					Editable: true,
					EditType: "text",
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
					Editable: true,
					EditType: "select",
					EditOptions: ["Enabled", "Disabled"],
				},
				{
					Label: "Save Images Next to Content",
					Value: config?.SaveImageNextToContent ? "Enabled" : "Disabled",
					Tooltip: "Whether to save images next to the associated content.",
					Editable: true,
					EditType: "select",
					EditOptions: ["Enabled", "Disabled"],
				},
				{
					Label: "Auto Download",
					Value: config?.AutoDownload?.Enabled ? "Enabled" : "Disabled",
					Tooltip: "Whether auto-download is enabled.",
					Editable: true,
					EditType: "select",
					EditOptions: ["Enabled", "Disabled"],
				},
				{
					Label: "Auto Download Cron",
					Value: config?.AutoDownload?.Cron,
					Tooltip: parseCronToHumanReadable(config?.AutoDownload?.Cron || ""),
					Editable: true,
					EditType: "text",
				},
			],
		},
		Kometa: {
			Title: "Kometa",
			Fields: [
				{
					Label: "Remove Labels",
					Value: config?.Kometa?.RemoveLabels ? "Enabled" : "Disabled",
					Tooltip: "Whether to remove labels from media items.",
					Editable: true,
					EditType: "select",
					EditOptions: ["Enabled", "Disabled"],
				},
				{
					Label: "Labels",
					Value:
						Array.isArray(config?.Kometa?.Labels) && config?.Kometa?.Labels.length > 1
							? config?.Kometa?.Labels.join(", ")
							: config?.Kometa?.Labels?.[0] || "",
					Tooltip: "The list of labels to apply to media items.",
					Editable: true,
					EditType: "text",
				},
			],
		},
		Notifications: {
			Title: "Notifications",
			Fields: [
				{
					Label: "Provider",
					Value: config?.Notification?.Provider,
					Tooltip: "The provider for notifications. Currently, the only provider is Discord.",
					Editable: true,
					EditType: "select",
					EditOptions: ["Discord"],
				},
				{
					Label: "Webhook",
					Value: config?.Notification?.Webhook,
					Tooltip: "The webhook URL for the provider. This can be obtained by creating a webhook in Discord.",
					Editable: true,
					EditType: "text",
				},
			],
			Buttons: [
				{
					Label: "Send Test Notification",
					Icon: config?.Notification?.Provider === "Discord" ? <FaDiscord /> : null,
					onClick: sendTestNotification,
				},
			],
		},
		Logging: {
			Title: "Logging",
			Fields: [
				{
					Label: "Logging Level",
					Value: config?.Logging?.Level,
					Tooltip: "The logging level (e.g., DEBUG, INFO, WARN, ERROR).",
					Editable: true,
					EditType: "select",
					EditOptions: LOGGING_OPTIONS,
				},
				{
					Label: "Log File Path",
					Value: config?.Logging?.File,
					Tooltip: "The file path where logs are stored.",
					Editable: false,
				},
			],
			Buttons: [
				{
					Label: "View Logs",
					Icon: <Logs />,
					onClick: handleViewLogs,
				},
				{
					Label: "Clear Old Logs",
					Variant: "destructive",
					Icon: <CircleX />,
					onClick: handleClearOldLogs,
				},
			],
		},
		"Admin Tools": {
			Title: "Admin Tools",
			Buttons: [
				{
					Label: "Clear Temp Images Folder",
					Variant: "destructive",
					Icon: <CircleX />,
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
				<ErrorMessage error={error} />
			) : (
				<>
					<div className="flex items-center justify-between mb-6">
						<h1
							className="text-3xl font-bold mb-4"
							onDoubleClick={() => {
								if (debugEnabled) {
									toggleDebugMode(false);
									toast.info("Debug mode disabled");
								} else {
									toggleDebugMode(true);
									toast.info("Debug mode enabled");
								}
								localStorage.setItem("debugMode", (!debugEnabled).toString());
							}}
						>
							Settings
						</h1>
					</div>
					<p className="mb-6 text-gray-600 dark:text-gray-400">
						Manage your application settings and configurations here
					</p>
					{AppConfig &&
						Object.entries(AppConfig).map(([key, value]) => {
							// Filter out fields with empty values (after trimming)
							const fieldsToShow =
								"Fields" in value
									? value.Fields.filter(
											(field) => field.Value && field.Value.toString().trim() !== ""
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
							if (!fieldsToShow.length && !hasButtons) return null;

							return (
								<Card className="mb-4" key={key}>
									<CardHeader>
										<h2 className="text-xl font-semibold">{value.Title}</h2>
									</CardHeader>
									<CardContent>
										<div className="space-y-2">
											{"Fields" in value &&
												fieldsToShow.map((field, index) => (
													<div key={index}>
														<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
															{field.Label}
														</label>
														{field.Label === "Libraries" || field.Label === "Labels" ? (
															<>
																{(field.Value ?? "")
																	.split(", ")
																	.filter((item: string) => item.trim() !== "")
																	.map((item: string, idx: number) => (
																		<Badge key={idx} className="mr-2 mt-1 text-sm">
																			{item}
																		</Badge>
																	))}
															</>
														) : (
															<div className="flex items-center gap-2 mt-1">
																<Input
																	value={field.Value}
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
																		{field.Tooltip}
																	</PopoverContent>
																</Popover>
															</div>
														)}
													</div>
												))}
											{"Buttons" in value &&
												(
													value.Buttons as {
														onClick: () => void;
														Variant?: "default" | "destructive";
														Label: string;
														Icon?: React.ReactNode;
													}[]
												).map((button, index) => (
													<Button
														key={index}
														variant={button.Variant || "default"}
														onClick={button.onClick}
														className="w-full hover:text-white cursor-pointer"
													>
														{button.Label} {button.Icon && button.Icon}
													</Button>
												))}
										</div>
									</CardContent>
								</Card>
							);
						})}
					{/* Debug Mode Toggle */}
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
				</>
			)}
		</div>
	);
};

export default SettingsPage;
