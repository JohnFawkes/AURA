"use client";

import { formatExactDateTime } from "@/helper/format-date-last-updates";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchDownloadQueueEntries } from "@/services/download-queue/fetch-queue-entries";
import { DownloadQueueStatus, fetchDownloadQueueStatus } from "@/services/download-queue/get-status";
import { Globe, Wifi } from "lucide-react";

import React, { useCallback, useEffect, useRef, useState } from "react";

import DownloadQueueEntry from "@/components/shared/download-queue-entry";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { RefreshButton } from "@/components/shared/refresh-button";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { H3 } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

const DownloadQueuePage: React.FC = () => {
	// Refs - Fetching
	const isFetchingRef = useRef(false);

	// States - Queue Entries
	const [inProgressEntries, setInProgressEntries] = useState<DBMediaItemWithPosterSets[]>([]);
	const [errorEntries, setErrorEntries] = useState<DBMediaItemWithPosterSets[]>([]);
	const [warningEntries, setWarningEntries] = useState<DBMediaItemWithPosterSets[]>([]);

	// States - Queue Status
	const [queueStatus, setQueueStatus] = useState<DownloadQueueStatus>({
		time: "",
		status: "",
		message: "",
		warnings: [],
		errors: [],
	});
	const [secondsToNextRun, setSecondsToNextRun] = useState<number>(0);

	// States - Loading & Error
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const [webSocketConnected, setWebSocketConnected] = useState(false);

	// Set the Document Title
	useEffect(() => {
		document.title = `aura | Download Queue`;
	}, []);

	// Fetch Queue Entries
	const fetchQueueEntries = useCallback(async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;

		try {
			setLoading(true);

			const response = await fetchDownloadQueueEntries();

			if (response.status === "error") {
				setError(response);
				return;
			}

			setInProgressEntries(response.data?.in_progress_entries || []);
			setErrorEntries(response.data?.error_entries || []);
			setWarningEntries(response.data?.warning_entries || []);
			setError(null);
		} catch (error) {
			setError(ReturnErrorMessage<unknown>(error));
		} finally {
			isFetchingRef.current = false;
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchQueueEntries();
	}, [fetchQueueEntries, queueStatus.status]);

	useEffect(() => {
		const fetchStatus = async () => {
			if (webSocketConnected) return;
			try {
				const statusResponse = await fetchDownloadQueueStatus();
				if (statusResponse.status === "error") {
					throw new Error("Error fetching status");
				}
				const status = statusResponse.data || {
					time: new Date().toISOString(),
					status: "Error",
					message: "Unable to get status from server",
					warnings: [],
					errors: ["No status data"],
				};
				setQueueStatus(status);
			} catch {
				const errorResponse: DownloadQueueStatus = {
					time: new Date().toISOString(),
					status: "Error",
					message: "Failed to fetch status",
					warnings: [],
					errors: [],
				};
				setQueueStatus(errorResponse);
			}
		};

		fetchStatus();
		const interval = setInterval(fetchStatus, 1000 * 2); // Refresh every 2 seconds

		return () => clearInterval(interval);
	}, [webSocketConnected]);

	useEffect(() => {
		let ws: WebSocket | null = null;
		let isMounted = true;

		const connectWebSocket = () => {
			ws = new WebSocket("ws://localhost:8888/api/download-queue/status-ws");

			ws.onopen = () => {
				log("INFO", "Download Queue WS", "Connection", "WebSocket connection established");
				setWebSocketConnected(true);
			};

			ws.onmessage = (event) => {
				try {
					const status = JSON.parse(event.data);
					if (isMounted) setQueueStatus(status);
				} finally {
					// No-op
				}
			};

			ws.onerror = () => {
				if (isMounted) setWebSocketConnected(false);
			};

			ws.onclose = () => {
				if (isMounted) setWebSocketConnected(false);
				if (isMounted) setTimeout(connectWebSocket, 1000 * 10);
			};
		};

		connectWebSocket();

		return () => {
			isMounted = false;
			if (ws) ws.close();
		};
	}, []);

	useEffect(() => {
		const updateNextRunTime = () => {
			const now = new Date();
			const next = new Date(now);
			next.setSeconds(0, 0);
			if (now.getSeconds() !== 0 || now.getMilliseconds() !== 0) {
				next.setMinutes(now.getMinutes() + 1);
			}

			const diff = Math.max(0, Math.floor((next.getTime() - now.getTime()) / 1000));
			setSecondsToNextRun(diff);
		};

		updateNextRunTime();
		const interval = setInterval(updateNextRunTime, 1000);

		return () => clearInterval(interval);
	}, []);

	if (loading) {
		return <Loader className="mt-10" message="Loading download queue entries..." />;
	}

	if (error) {
		return (
			<div className="flex flex-col items-center p-6 gap-4">
				<ErrorMessage error={error} />
			</div>
		);
	}

	const defaultAccordionValues = [
		inProgressEntries.length > 0 ? "in_progress" : null,
		errorEntries.length > 0 ? "error_entries" : null,
		warningEntries.length > 0 ? "warning_entries" : null,
	].filter(Boolean) as string[];

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			<h1 className="text-3xl font-bold mb-4">Download Queue</h1>

			{typeof secondsToNextRun === "number" && (
				<div className="w-full max-w-4xl mb-2 text-xs text-muted-foreground text-right flex items-center justify-end gap-2">
					<span className="font-mono">Next Run: {secondsToNextRun}s</span>
					{webSocketConnected ? (
						<span title="WebSocket Connected">
							<Wifi className="inline-block h-4 w-4 text-green-500" />
						</span>
					) : (
						<span title="HTTP Polling">
							<Globe className="inline-block h-4 w-4 text-blue-500" />
						</span>
					)}
				</div>
			)}
			<pre
				className={cn(
					"w-full max-w-4xl mb-4 p-3 rounded text-xs whitespace-pre-wrap border",
					queueStatus.status === "Error"
						? "border-red-400 text-red-500"
						: queueStatus.status === "Warning"
							? "border-yellow-400 text-yellow-500"
							: queueStatus.status === "Success"
								? "border-green-400 text-green-500"
								: queueStatus.status === "Idle - Queue Empty"
									? "border-gray-400 text-gray-500"
									: "border-primary text-primary"
				)}
			>
				{queueStatus.time && (
					<div>
						<b>Last Run:</b> {formatExactDateTime(queueStatus.time)}
					</div>
				)}
				{queueStatus.status && (
					<div>
						<b>Status:</b> {queueStatus.status}
					</div>
				)}
				{queueStatus.message && <div className="mt-2 mb-2">{queueStatus.message}</div>}
				{queueStatus.warnings && queueStatus.warnings.length > 0 && (
					<div className="mt-1">
						<b className="text-yellow-500">Warnings:</b>
						<ul className="list-disc ml-5 text-yellow-500">
							{queueStatus.warnings.map((w, i) => (
								<li key={i}>{w}</li>
							))}
						</ul>
					</div>
				)}
				{queueStatus.errors && queueStatus.errors.length > 0 && (
					<div className="mt-1">
						<b className="text-red-500">Errors:</b>
						<ul className="list-disc ml-5 text-red-500">
							{queueStatus.errors.map((e, i) => (
								<li key={i}>{e}</li>
							))}
						</ul>
					</div>
				)}
			</pre>

			{inProgressEntries.length === 0 && errorEntries.length === 0 && warningEntries.length === 0 && (
				<p className="text-gray-500">No download queue entries found</p>
			)}

			<div className="w-full max-w-4xl">
				<Accordion type="multiple" className="mb-4" defaultValue={defaultAccordionValues}>
					{inProgressEntries.length > 0 && (
						<AccordionItem value="in_progress">
							<AccordionTrigger
								className={cn(
									"cursor-pointer",
									"hover:underline-none focus:underline-none underline-none hover:no-underline focus:no-underline"
								)}
							>
								<H3>In Progress Entries</H3>
							</AccordionTrigger>
							<AccordionContent>
								{inProgressEntries.length === 0 ? (
									<p className="text-gray-500">No entries in progress.</p>
								) : (
									<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
										{inProgressEntries.map((entry) => (
											<DownloadQueueEntry
												key={entry.TMDB_ID}
												entry={entry}
												fetchQueueEntries={fetchQueueEntries}
											/>
										))}
									</div>
								)}
							</AccordionContent>
						</AccordionItem>
					)}

					{errorEntries.length > 0 && (
						<AccordionItem value="error_entries">
							<AccordionTrigger
								className={cn(
									"cursor-pointer",
									"hover:underline-none focus:underline-none underline-none hover:no-underline focus:no-underline"
								)}
							>
								<H3>Error Entries</H3>
							</AccordionTrigger>
							<AccordionContent>
								{errorEntries.length === 0 ? (
									<p className="text-gray-500">No error entries.</p>
								) : (
									<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
										{errorEntries.map((entry) => (
											<DownloadQueueEntry
												key={entry.TMDB_ID}
												entry={entry}
												fetchQueueEntries={fetchQueueEntries}
											/>
										))}
									</div>
								)}
							</AccordionContent>
						</AccordionItem>
					)}

					{warningEntries.length > 0 && (
						<AccordionItem value="warning_entries">
							<AccordionTrigger
								className={cn(
									"cursor-pointer",
									"hover:underline-none focus:underline-none underline-none hover:no-underline focus:no-underline"
								)}
							>
								<H3>Warning Entries</H3>
							</AccordionTrigger>
							<AccordionContent>
								{warningEntries.length === 0 ? (
									<p className="text-gray-500">No warning entries.</p>
								) : (
									<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4  gap-2">
										{warningEntries.map((entry) => (
											<DownloadQueueEntry
												key={entry.TMDB_ID}
												entry={entry}
												fetchQueueEntries={fetchQueueEntries}
											/>
										))}
									</div>
								)}
							</AccordionContent>
						</AccordionItem>
					)}
				</Accordion>
			</div>

			{/* Refresh Button */}
			<RefreshButton onClick={() => fetchQueueEntries()} />
		</div>
	);
};

export default DownloadQueuePage;
