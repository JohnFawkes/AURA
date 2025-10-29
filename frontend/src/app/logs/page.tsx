"use client";

import { fetchLogContents, postClearOldLogs } from "@/services/settings-onboarding/api-logs-actions";
import { EllipsisIcon, Filter, SaveIcon } from "lucide-react";
import { toast } from "sonner";

import { useEffect, useState } from "react";

import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { LogsFilter } from "@/components/shared/logs-filter";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
	Drawer,
	DrawerContent,
	DrawerDescription,
	DrawerHeader,
	DrawerTitle,
	DrawerTrigger,
} from "@/components/ui/drawer";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Small } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";

import { APIResponse, LogAction, LogData } from "@/types/api/api-response";

function parseLogs(rawLogs: string): LogData[] {
	try {
		return rawLogs
			.split("\n")
			.map((line) => {
				try {
					const obj = JSON.parse(line);
					// Optionally validate obj here
					return obj as LogData;
				} catch {
					return undefined;
				}
			})
			.filter((log): log is LogData => !!log);
	} catch {
		return [];
	}
}

const RouteToFunctionMap: { label: string; value: string; section: string }[] = [
	{ label: "User Login", value: "/api/login", section: "AUTH" },

	// Config Routes
	{ label: "Get Config", value: "/api/config", section: "CONFIG" },
	{ label: "Get Config Status", value: "/api/config/status", section: "CONFIG" },
	{ label: "Reload Config", value: "/api/config/reload", section: "CONFIG" },
	{ label: "Update Config", value: "/api/config/update", section: "CONFIG" },
	{ label: "Validate Mediux Token", value: "/api/config/validate/mediux", section: "CONFIG" },
	{ label: "Validate Media Server Info", value: "/api/config/validate/mediaserver", section: "CONFIG" },
	{ label: "Validate Sonarr Connection", value: "/api/config/validate/sonarr", section: "CONFIG" },
	{ label: "Validate Radarr Connection", value: "/api/config/validate/radarr", section: "CONFIG" },
	{ label: "Test Notifications", value: "/api/config/validate/notification", section: "CONFIG" },

	// Logging Routes
	{ label: "Get Logs", value: "/api/log", section: "LOGS" },
	{ label: "Clear Logs", value: "/api/log/clear", section: "LOGS" },

	// Temp Images Routes
	{ label: "Clear Temp Images", value: "/api/temp-images/clear", section: "TEMP IMAGES" },

	// Media Server Routes
	{ label: "Get Media Server Status", value: "/api/mediaserver/status", section: "MEDIA" },
	{ label: "Get Media Server Type", value: "/api/mediaserver/type", section: "MEDIA" },
	{ label: "Get Media Server Library Options", value: "/api/mediaserver/library-options", section: "MEDIA" },
	{ label: "Get Media Server Library Sections", value: "/api/mediaserver/sections", section: "MEDIA" },
	{ label: "Get Media Server Sections & Items", value: "/api/mediaserver/sections/items", section: "MEDIA" },
	{ label: "Get Item Content", value: "/api/mediaserver/item", section: "MEDIA" },
	{ label: "Download and Update", value: "/api/mediaserver/download", section: "MEDIA" },
	{ label: "Add Item to Download Queue", value: "/api/mediaserver/add-to-queue", section: "MEDIA" },

	// MediUX Routes
	{ label: "Get All Sets", value: "/api/mediux/sets", section: "MEDIUX" },
	{ label: "Get Sets From User", value: "/api/mediux/sets-by-user", section: "MEDIUX" },
	{ label: "Get Set by ID", value: "/api/mediux/set-by-id", section: "MEDIUX" },
	{ label: "Get Images From Set", value: "/api/mediux/image", section: "MEDIUX" },
	{ label: "Get User Following/Hiding Sets", value: "/api/mediux/user-follow-hiding", section: "MEDIUX" },

	// Database Routes
	{ label: "Get All Items", value: "/api/db/get-all", section: "DATABASE" },
	{ label: "Delete Item", value: "/api/db/delete", section: "DATABASE" },
	{ label: "Update Item", value: "/api/db/update", section: "DATABASE" },
	{ label: "Add Item", value: "/api/db/add", section: "DATABASE" },
	{ label: "Force Recheck on Item", value: "/api/db/force-recheck", section: "DATABASE" },
];

function getFunctionFromRoute(route: string | undefined): string {
	if (!route) return "Background Task";
	const mapping = RouteToFunctionMap.find((map) => map.value === route);
	return mapping ? mapping.label : route;
}

export default function LogsPage() {
	const [logEntries, setLogEntries] = useState<LogData[]>([]);
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<APIResponse<string> | null>(null);
	const [isMounted, setIsMounted] = useState(false);

	// Is Wide Screen:
	const [isWideScreen, setIsWideScreen] = useState(false);

	const [numberOfActiveFilters, setNumberOfActiveFilters] = useState(0);
	const [levelsFilter, setLevelsFilter] = useState<string[]>([]);
	const [statusFilter, setStatusFilter] = useState<string[]>([]);
	const [actionsOptions, setActionsOptions] = useState<{ label: string; value: string; section: string }[]>([]);
	const [actionsFilter, setActionsFilter] = useState<string[]>([]);

	useEffect(() => {
		document.title = "aura | Logs";
	}, []);

	useEffect(() => {
		if (isMounted) return;
		setIsMounted(true);

		const fetchLogs = async () => {
			try {
				setLoading(true);
				const response = await fetchLogContents();
				if (response.status === "error") {
					setError(response);
					setLogEntries([]);
					return;
				}

				const logsString = response.data || "";
				if (logsString.trim() === "") {
					setLogEntries([]);
					setError(null);
					return;
				}

				const parsedLogs = parseLogs(logsString).sort((a, b) => {
					const aTime = new Date(a.timestamp || a.time).getTime();
					const bTime = new Date(b.timestamp || b.time).getTime();
					return bTime - aTime; // newest first
				});
				setLogEntries(parsedLogs);

				// Extract unique action names for the actions filter
				const actionNamesSet = new Set<string>();
				parsedLogs.forEach((log) => {
					if (!log.route && (!Array.isArray(log.actions) || log.actions.length === 0)) {
						return;
					}
					if (log.route?.path) {
						const label = getFunctionFromRoute(log.route.path);
						actionNamesSet.add(label);
					} else {
						actionNamesSet.add(log.message || "Background Task");
					}
				});
				const actionOptions = Array.from(actionNamesSet)
					.sort()
					.map((name) => ({
						label: name,
						value: name,
						section:
							name === "Background Task"
								? "BACKGROUND"
								: RouteToFunctionMap.find((map) => map.label === name)?.section || "AURA BACKGROUND",
					}));
				setActionsOptions(actionOptions);

				setError(null);
			} catch (error) {
				setError({
					status: "error",
					error: {
						message: "Failed to fetch logs",
						function: "fetchLogs",
						line_number: 0,
						help: "Check server connectivity and permissions.",
						detail: error instanceof Error ? { message: error.message } : {},
					},
				});
				setLogEntries([]);
			} finally {
				setLoading(false);
			}
		};

		fetchLogs();
	}, [isMounted]);

	const clearLogsFromToday = async () => {
		try {
			const response = await postClearOldLogs(true);
			if (response.status === "error") {
				toast.error(response.error?.message || "Failed to clear old logs");
				return;
			}
			toast.success(response.data || "Logs from today cleared successfully");
			setIsMounted(false); // Reset the component to refetch logs
		} catch {
			toast.error("An unexpected error occurred");
		}
	};

	useEffect(() => {
		let count = 0;
		if (levelsFilter.length > 0) count++;
		if (statusFilter.length > 0) count++;
		if (actionsFilter.length > 0) count++;
		log("INFO", "Logs Page", "Filters", "Number of active filters: " + count, {
			levelsFilter,
			statusFilter,
			actionsFilter,
		});
		setNumberOfActiveFilters(count);
	}, [levelsFilter, statusFilter, actionsFilter]);

	// Change isWideScreen on window resize
	useEffect(() => {
		const handleResize = () => {
			setIsWideScreen(window.innerWidth >= 1300);
		};
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, []);

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			<Card className="w-full flex-1">
				<CardHeader>
					<div className="flex flex-col md:flex-row md:items-center md:justify-between w-full space-y-2 md:space-y-0">
						<h2 className="text-lg font-medium text-center md:text-left">Application Logs</h2>
						<div className="flex flex-row justify-center md:justify-end space-x-2">
							{isWideScreen ? (
								<Popover>
									<PopoverTrigger asChild>
										<div>
											<Button
												variant="outline"
												className={cn(numberOfActiveFilters > 0 && "ring-2 ring-primary")}
											>
												Filters {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
												<Filter className="h-5 w-5" />
											</Button>
										</div>
									</PopoverTrigger>
									<PopoverContent
										side="right"
										align="start"
										className="w-[450px] p-2 bg-background border border-primary"
									>
										<LogsFilter
											levelsFilter={levelsFilter}
											setLevelsFilter={setLevelsFilter}
											statusFilter={statusFilter}
											setStatusFilter={setStatusFilter}
											actionsOptions={actionsOptions}
											actionsFilter={actionsFilter}
											setActionsFilter={setActionsFilter}
										/>
									</PopoverContent>
								</Popover>
							) : (
								<Drawer direction="left">
									<DrawerTrigger asChild>
										<Button
											variant="outline"
											className={cn(
												numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1"
											)}
										>
											Filters {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
											<Filter className="h-5 w-5" />
										</Button>
									</DrawerTrigger>
									<DrawerContent>
										<DrawerHeader className="my-0">
											<DrawerTitle className="mb-0">Filters</DrawerTitle>
											<DrawerDescription className="mb-0">
												Use the options below to filter your saved sets.
											</DrawerDescription>
										</DrawerHeader>
										<Separator className="my-1 w-full" />
										<LogsFilter
											levelsFilter={levelsFilter}
											setLevelsFilter={setLevelsFilter}
											statusFilter={statusFilter}
											setStatusFilter={setStatusFilter}
											actionsOptions={actionsOptions}
											actionsFilter={actionsFilter}
											setActionsFilter={setActionsFilter}
										/>
									</DrawerContent>
								</Drawer>
							)}
							<Button variant="destructive" onClick={clearLogsFromToday}>
								Clear Today's Logs
							</Button>
						</div>
					</div>
				</CardHeader>

				<CardContent>
					{loading ? (
						<Loader />
					) : error ? (
						<ErrorMessage error={error} />
					) : (
						<LogList
							logEntries={logEntries}
							levelsFilter={levelsFilter}
							statusFilter={statusFilter}
							actionsFilter={actionsFilter}
						/>
					)}
				</CardContent>
			</Card>
		</div>
	);
}

function formatBytes(bytes: number): string {
	if (bytes === 0) return "0 Bytes";
	const k = 1024;
	const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

function formatElapsedMicroseconds(us: number): string {
	if (us < 1000) return `${us} µs`;
	if (us < 1_000_000) return `${(us / 1000).toFixed(2)} ms`;
	return `${(us / 1_000_000).toFixed(2)} s`;
}

function actionOrSubActionHasStatus(actions: LogAction[] | undefined, statusFilter: string[]): boolean {
	if (!actions || statusFilter.length === 0) return true;
	return actions.some(
		(action) =>
			statusFilter.includes(action.status) ||
			(action.sub_actions && actionOrSubActionHasStatus(action.sub_actions, statusFilter))
	);
}

function actionOrSubActionHasLevel(actions: LogAction[] | undefined, levelsFilter: string[]): boolean {
	if (!actions || levelsFilter.length === 0) return true;
	return actions.some(
		(action) =>
			(action.level && levelsFilter.includes(action.level)) ||
			(action.sub_actions && actionOrSubActionHasLevel(action.sub_actions, levelsFilter))
	);
}

function LogList({
	logEntries,
	levelsFilter,
	statusFilter,
	actionsFilter,
}: {
	logEntries: LogData[];
	levelsFilter: string[];
	statusFilter: string[];
	actionsFilter: string[];
}) {
	const filteredLogEntries = logEntries.filter((log: LogData) => {
		// Ignore logs without route or actions
		if (!log.route && (!Array.isArray(log.actions) || log.actions.length === 0)) return false;

		// Apply levels filter (recursive)
		if (levelsFilter.length > 0 && !actionOrSubActionHasLevel(log.actions, levelsFilter)) return false;

		// Apply status filter (recursive)
		if (statusFilter.length > 0 && !actionOrSubActionHasStatus(log.actions, statusFilter)) return false;

		// Apply actions filter
		if (actionsFilter.length > 0) {
			const actionName = log.route ? getFunctionFromRoute(log.route.path) : log.message || "Background Task";
			if (!actionsFilter.includes(actionName)) return false;
		}
		return true;
	});

	if (filteredLogEntries.length === 0) {
		return <div className="text-muted-foreground">No structured logs available.</div>;
	}

	return (
		<div className="space-y-1">
			{filteredLogEntries.map((log, idx) => {
				const baseLabel = getFunctionFromRoute(log.route?.path);
				let mainLabel = baseLabel;
				if (mainLabel === "Background Task" && log.actions && log.actions.length > 0) {
					mainLabel = log.message || log.actions[0].name || "Background Task";
				}

				return (
					<Card
						key={idx}
						className={cn(
							"w-full",
							"border-2 border-transparent transition-colors",
							log.level === "error" ? "border-red-500" : ""
						)}
					>
						<CardHeader className={cn("gap-0 mb-0")}>
							<div className="flex flex-col">
								<div className="flex flex-row items-center justify-between mb-1 w-full">
									{mainLabel && <Small className="font-medium text-lg flex-1">{mainLabel}</Small>}
									<>
										{/* Drop Down Menu Options To Export Log Line */}
										<DropdownMenu>
											<DropdownMenuTrigger
												asChild
												className={cn(
													"cursor-pointer ml-2",
													"hover:brightness-120 active:scale-95 transition-all"
												)}
											>
												<EllipsisIcon className="h-5 w-5 text-muted-foreground" />
											</DropdownMenuTrigger>
											<DropdownMenuContent className="w-56 md:w-64" side="bottom" align="end">
												<DropdownMenuItem
													className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
													onClick={() => {
														const logDataStr = JSON.stringify(log, null, 2);
														const blob = new Blob([logDataStr], {
															type: "application/json",
														});
														const url = URL.createObjectURL(blob);
														const link = document.createElement("a");
														link.href = url;
														link.download = `log-${log.message}-${log.status}.json`;
														document.body.appendChild(link);
														link.click();
														document.body.removeChild(link);
														URL.revokeObjectURL(url);
													}}
												>
													<SaveIcon className="w-6 h-6 mr-2" />
													Export {log.status.toUpperCase()} Log
												</DropdownMenuItem>
											</DropdownMenuContent>
										</DropdownMenu>
									</>
								</div>
								{log.elapsed_us !== undefined && (
									<Small className="text-sm text-muted-foreground whitespace-nowrap">
										Elapsed Time: {formatElapsedMicroseconds(log.elapsed_us)}
									</Small>
								)}
								{log.route && (
									<Accordion type="single" collapsible>
										<AccordionItem value="route-info">
											<AccordionTrigger
												className={cn(
													"cursor-pointer",
													"hover:underline-none focus:underline-none underline-none hover:no-underline focus:no-underline"
												)}
											>
												<span className="text-sm font-semibold text-muted-foreground">
													Route Info
												</span>
											</AccordionTrigger>
											<AccordionContent>
												<div className="flex flex-col">
													{log.route.method && (
														<Small className="text-sm text-muted-foreground">
															Method: {log.route.method}
														</Small>
													)}
													{log.route.path && (
														<Small className="text-sm text-muted-foreground">
															Path: {log.route.path}
														</Small>
													)}
													{log.route.ip && (
														<Small className="text-sm text-muted-foreground">
															IP: {log.route.ip}
														</Small>
													)}
													{log.route.response_bytes !== undefined && (
														<Small className="text-sm text-muted-foreground">
															Response Size: {formatBytes(log.route.response_bytes)}
														</Small>
													)}
													{log.route.params && Object.keys(log.route.params).length > 0 && (
														<Small className="text-sm text-muted-foreground">
															Params:{" "}
															{Object.entries(log.route.params).map(([key, value]) => (
																<span key={key} className="mr-2">
																	<strong>{key}:</strong> {value}
																</span>
															))}
														</Small>
													)}
												</div>
											</AccordionContent>
										</AccordionItem>
									</Accordion>
								)}
							</div>
						</CardHeader>
						<CardContent className="mt-0">
							{Array.isArray(log.actions) && log.actions.length > 0 && (
								<Accordion type="multiple">
									{log.actions.map((action: LogAction, i: number) => [
										<LogActionAccordion
											key={action.name + action.timestamp + i}
											action={action}
											isSubAction={false}
										/>,
										<Separator
											key={"sep-" + action.name + action.timestamp + i}
											className="my-2"
										/>,
									])}
								</Accordion>
							)}
						</CardContent>
					</Card>
				);
			})}
		</div>
	);
}

function formatFullTimestamp(dateString: string): string {
	const d = new Date(dateString);
	const pad = (n: number, z = 2) => String(n).padStart(z, "0");
	const hours = d.getHours();
	const am_pm = hours >= 12 ? "PM" : "AM";
	const hour12 = hours % 12 === 0 ? 12 : hours % 12;
	return `${pad(d.getMonth() + 1)}/${pad(d.getDate())}/${d.getFullYear()} ${pad(hour12)}:${pad(d.getMinutes())}:${pad(d.getSeconds())} ${am_pm}`;
}

function LogActionAccordion({ action, isSubAction }: { action: LogAction; isSubAction?: boolean }) {
	return (
		<AccordionItem className={cn(isSubAction ? "ml-4" : "")} value={action.name + (action.timestamp || "")}>
			<AccordionTrigger
				hideChevron={action.sub_actions === undefined || action.sub_actions.length === 0}
				className={cn(
					{ "cursor-pointer": action.sub_actions && action.sub_actions.length > 0 },
					"hover:underline-none focus:underline-none underline-none hover:no-underline focus:no-underline"
				)}
			>
				<div className="flex flex-col w-full">
					<div className="font-medium">{action.name}</div>
					<div className="flex flex-col w-full">
						<div className="text-xs text-muted-foreground">
							{action.level && (
								<span
									className={
										action.level === "error"
											? "text-red-500"
											: action.level === "warn"
												? "text-yellow-600"
												: action.level === "info"
													? "text-blue-600"
													: action.level === "debug"
														? "text-green-600"
														: "text-purple-600"
									}
								>
									{action.level?.toUpperCase() || ""}
								</span>
							)}
							{/* Only show • if both level and status exist */}
							{action.level && action.status && " • "}
							{action.status && (
								<span
									className={
										action.status === "error"
											? "text-red-500"
											: action.status === "warn"
												? "text-yellow-600"
												: "text-green-600"
									}
								>
									{action.status.toUpperCase() || ""}
								</span>
							)}
						</div>
						<div className="text-xs text-muted-foreground">
							{action.timestamp && <>{formatFullTimestamp(action.timestamp)}</>}
							{/* Only show • if both timestamp and elapsed_us exist */}
							{action.timestamp && action.elapsed_us !== undefined && " • "}
							{action.elapsed_us !== undefined && <>{formatElapsedMicroseconds(action.elapsed_us)}</>}
						</div>
						{action.warnings && Object.keys(action.warnings).length > 0 && (
							<div className="text-xs text-yellow-600 mt-1">
								Warnings:
								<pre className="rounded p-2 mt-1 text-xs overflow-x-auto max-h-40 whitespace-pre-wrap break-words">
									{JSON.stringify(action.warnings, null, 2)}
								</pre>
							</div>
						)}
						{action.result && Object.keys(action.result).length > 0 && (
							<div className="text-xs mt-1">
								Results:{" "}
								<pre className="rounded p-2 mt-1 text-xs overflow-x-auto max-h-40 whitespace-pre-wrap break-words">
									{JSON.stringify(action.result, null, 2)}
								</pre>
							</div>
						)}
						{action.error && <div className="text-xs text-red-500 mt-1">Error: {action.error.message}</div>}
					</div>
				</div>
			</AccordionTrigger>
			{action.sub_actions && action.sub_actions.length > 0 && (
				<AccordionContent>
					{/* Recursively render sub-actions if present */}
					<Accordion type="multiple" className="ml-2 border-l-2 border-muted">
						{action.sub_actions.map((sub: LogAction, idx: number) => (
							<LogActionAccordion key={idx} action={sub} isSubAction={true} />
						))}
					</Accordion>
				</AccordionContent>
			)}
		</AccordionItem>
	);
}
