"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import ErrorMessage from "@/components/ui/error-message";
import { Input } from "@/components/ui/input";
import Loader from "@/components/ui/loader";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import { fetchConfig } from "@/services/api.settings";
import { AppConfig } from "@/types/config";
import { useRouter } from "next/navigation";
import React, { useEffect, useState } from "react";

const SettingsPage: React.FC = () => {
	const router = useRouter();
	const [config, setConfig] = useState<AppConfig | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	// Fetch configuration data
	useEffect(() => {
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
				console.log("Fetched config:", appConfig);
				console.log("Loading state:", loading);
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
	}, [loading]);

	const handleViewLogs = () => {
		router.push("/logs");
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

	// Help tooltips for each field
	const helpTexts = {
		MediaServerType:
			"The type of media server (e.g., Plex, Emby, Jellyfin).",
		MediaServerURL: "The base URL of the media server.",
		MediaServerToken:
			"The authentication token for accessing the media server.",
		MediaServerLibraries:
			"The list of libraries managed by the media server.",
		TMDBApiKey:
			"The API key for accessing TMDB services. This is not used in this version.",
		MediuxToken: "The authentication token for accessing Mediux services.",
		CacheImages: "Whether to cache images locally.",
		SaveImageNextToContent:
			"Whether to save images next to the associated content.",
		AutoDownloadEnabled: "Whether auto-download is enabled.",
		AutoDownloadCron: "The cron expression for scheduling auto-downloads.",
		LoggingLevel: "The logging level (e.g., DEBUG, INFO, WARN, ERROR).",
		LoggingFile: "The file path where logs are stored.",
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
					{/* Media Server Information */}
					<Card className="mb-2">
						<CardHeader>
							<h2 className="text-xl font-semibold">
								Media Server Information
							</h2>
						</CardHeader>
						<CardContent>
							<div className="space-y-2">
								{/* Server Type */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Server Type
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={
												config?.MediaServer.Type || ""
											}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.MediaServerType}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Server URL */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Server URL
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={
												config?.MediaServer.URL || ""
											}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.MediaServerURL}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Authentication Token */}
								<div className="space-y-1">
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Authentication Token
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={config?.MediaServer.Token}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.MediaServerToken}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>
							</div>
							<div className="space-y-2 mt-4">
								{/* Libraries */}
								<h3 className="text-lg font-medium text-gray-700 dark:text-gray-300">
									Libraries
								</h3>
								<div className="flex flex-wrap gap-2">
									{config?.MediaServer.Libraries.map(
										(library, index) => (
											<Badge
												key={index}
												className="text-sm"
											>
												{library.Name}
											</Badge>
										)
									)}
								</div>
							</div>
						</CardContent>
					</Card>

					{/* Other API Information */}
					<Card className="mb-2">
						<CardHeader>
							<h2 className="text-xl font-semibold">
								Other API Information
							</h2>
						</CardHeader>
						<CardContent>
							<div className="space-y-2">
								{/* TMDB API Key */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										TMDB API Key
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={config?.TMDB.ApiKey || ""}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.TMDBApiKey}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Mediux Token */}
								<div className="space-y-1">
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Mediux Token
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={config?.Mediux.Token || ""}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.MediuxToken}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>
							</div>
						</CardContent>
					</Card>

					{/* Other Settings Info */}
					<Card className="mb-2">
						<CardHeader>
							<h2 className="text-xl font-semibold">
								Other Settings
							</h2>
						</CardHeader>
						<CardContent>
							<div className="space-y-2">
								{/* Cache Images */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Cache Images
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={
												config?.CacheImages
													? "Enabled"
													: "Disabled"
											}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.CacheImages}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Save Images Next to Content */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Save Images Next to Content
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={
												config?.SaveImageNextToContent
													? "Enabled"
													: "Disabled"
											}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{
													helpTexts.SaveImageNextToContent
												}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Auto Download */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Auto Download
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={
												config?.AutoDownload?.Enabled
													? "Enabled"
													: "Disabled"
											}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.AutoDownloadEnabled}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Auto Download Cron */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Auto Download Cron
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={
												config?.AutoDownload?.Cron || ""
											}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.AutoDownloadCron +
													" " +
													parseCronToHumanReadable(
														config?.AutoDownload
															?.Cron || ""
													)}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>
							</div>
						</CardContent>
					</Card>

					{/* Logging Info */}
					<Card className="mb-2">
						<CardHeader>
							<h2 className="text-xl font-semibold">Logging</h2>
						</CardHeader>
						<CardContent>
							<div className="space-y-2">
								{/* Logging Level */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Logging Level
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={config?.Logging?.Level || ""}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.LoggingLevel}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* Log File Path */}
								<div>
									<label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
										Log File Path
									</label>
									<div className="flex items-center gap-2">
										<Input
											value={config?.Logging?.File || ""}
											disabled
											className="w-full"
										/>
										<Tooltip>
											<TooltipTrigger>
												<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
													?
												</span>
											</TooltipTrigger>
											<TooltipContent>
												{helpTexts.LoggingFile}
											</TooltipContent>
										</Tooltip>
									</div>
								</div>

								{/* View Logs Button */}
								<div className="mt-2">
									<Button
										className="w-full"
										onClick={handleViewLogs}
									>
										View Logs
									</Button>
								</div>
							</div>
						</CardContent>
					</Card>
				</>
			)}
		</div>
	);
};

export default SettingsPage;
