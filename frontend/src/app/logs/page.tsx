"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import {
	FetchLogContentsResponse,
	fetchLogContents,
	postClearOldLogs,
} from "@/services/settings-onboarding/api-logs-actions";
import { EllipsisIcon, SaveIcon, XCircle } from "lucide-react";
import { toast } from "sonner";

import { useEffect, useState } from "react";

import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import { FilterLogs } from "@/components/shared/filter-logs";
import Loader from "@/components/shared/loader";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";
import { Small } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { useLogsPageStore } from "@/lib/stores/page-store-logs";

import { APIResponse, LogAction, LogData } from "@/types/api/api-response";

export default function LogsPage() {
	// States - Loading & Error
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<APIResponse<FetchLogContentsResponse> | null>(null);

	// States - Responses and Data
	const [logEntries, setLogEntries] = useState<LogData[]>([]);
	const [totalLogEntries, setTotalLogEntries] = useState<number>(0);
	const [possibleActionsPaths, setPossibleActionsPaths] = useState<
		Record<string, { label: string; section: string }>
	>({});

	// States - Filters & Pagination
	const {
		levelsFilter,
		setLevelsFilter,
		statusFilter,
		setStatusFilter,
		actionsFilter,
		setActionsFilter,
		currentPage,
		setCurrentPage,
		itemsPerPage,
		setItemsPerPage,
	} = useLogsPageStore();

	useEffect(() => {
		document.title = "aura | Logs";
	}, []);

	useEffect(() => {
		const fetchLogs = async () => {
			try {
				setLoading(true);
				const response = await fetchLogContents(
					levelsFilter,
					statusFilter,
					actionsFilter,
					itemsPerPage,
					currentPage
				);
				if (response.status === "error") {
					setError(response);
					setLogEntries([]);
					setPossibleActionsPaths({});
					setTotalLogEntries(0);
					return;
				}
				const logEntries = response.data?.log_entries || [];

				setLogEntries(logEntries);
				setPossibleActionsPaths(response.data?.possible_actions_paths || {});
				setTotalLogEntries(response.data?.total_log_entries || 0);
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
				setPossibleActionsPaths({});
				setTotalLogEntries(0);
			} finally {
				setLoading(false);
			}
		};

		fetchLogs();
	}, [actionsFilter, currentPage, itemsPerPage, levelsFilter, statusFilter]);

	const clearLogsFromToday = async () => {
		try {
			const response = await postClearOldLogs(true);
			if (response.status === "error") {
				toast.error(response.error?.message || "Failed to clear old logs");
				return;
			}
			toast.success(response.data || "Logs from today cleared successfully");
			setLogEntries([]);
			setTotalLogEntries(0);
			setCurrentPage(1);
		} catch {
			toast.error("An unexpected error occurred");
		}
	};

	// Calculate total pages
	const totalPages = Math.ceil(totalLogEntries / itemsPerPage);

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			<Card className="w-full flex-1">
				<CardHeader>
					<div className="flex flex-col md:flex-row md:items-center md:justify-between w-full space-y-2 md:space-y-0">
						<h2 className="text-lg font-medium text-center md:text-left">Application Logs</h2>
						<div className="flex flex-row justify-center md:justify-end space-x-2">
							<FilterLogs
								levelsFilter={levelsFilter}
								setLevelsFilter={setLevelsFilter}
								statusFilter={statusFilter}
								setStatusFilter={setStatusFilter}
								actionsOptions={possibleActionsPaths}
								actionsFilter={actionsFilter}
								setActionsFilter={setActionsFilter}
								setCurrentPage={setCurrentPage}
								itemsPerPage={itemsPerPage}
								setItemsPerPage={setItemsPerPage}
							/>

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
					) : (!logEntries || logEntries.length === 0) && !error && !loading ? (
						<div className="w-full">
							<ErrorMessage
								error={ReturnErrorMessage<string>(
									[
										`No Log Entries found`,
										levelsFilter.length > 0
											? `with level${levelsFilter.length > 1 ? "s" : ""} ${levelsFilter
													.map((lvl) => `"${lvl}"`)
													.join(", ")}`
											: null,
										statusFilter.length > 0
											? `with status${statusFilter.length > 1 ? "es" : ""} ${statusFilter
													.map((st) => `"${st}"`)
													.join(", ")}`
											: null,
										actionsFilter.length > 0
											? `with action${actionsFilter.length > 1 ? "s" : ""} ${actionsFilter
													.map((act) => {
														if (
															act.startsWith("/api") &&
															possibleActionsPaths[act]?.label
														) {
															return `"${possibleActionsPaths[act].label}"`;
														}
														return `"${act}"`;
													})
													.join(", ")}`
											: null,
									]
										.filter(Boolean)
										.join("\n")
								)}
							/>
							<div className="text-center text-muted-foreground mt-4">
								<Button
									variant="outline"
									size="sm"
									onClick={() => {
										setLevelsFilter([]);
										setStatusFilter([]);
										setActionsFilter([]);
										setCurrentPage(1);
									}}
									className="text-sm"
								>
									<XCircle className="inline mr-1" />
									Clear All Filters
								</Button>
							</div>
						</div>
					) : (
						<>
							<LogList logEntries={logEntries} possibleActionsPaths={possibleActionsPaths} />

							{/* Pagination */}
							{itemsPerPage && (
								<CustomPagination
									currentPage={currentPage}
									totalPages={totalPages}
									setCurrentPage={setCurrentPage}
									scrollToTop={true}
									filterItemsLength={totalLogEntries}
									itemsPerPage={itemsPerPage}
								/>
							)}
						</>
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

function LogList({
	logEntries,
	possibleActionsPaths,
}: {
	logEntries: LogData[];
	possibleActionsPaths: Record<string, { label: string; section: string }>;
}) {
	return (
		<div className="space-y-1">
			{logEntries.map((log, idx) => {
				let mainLabel = "";
				if (log.route?.path) {
					mainLabel = possibleActionsPaths[log.route.path].label || log.route.path;
				} else if (log.actions && log.actions.length > 0) {
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
														const downloadName = mainLabel
															? mainLabel.replace(/\s+/g, "_").toLowerCase()
															: "";
														link.download = `log-${downloadName}-${log.status}.json`;
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
