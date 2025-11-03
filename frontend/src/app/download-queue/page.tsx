"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchDownloadQueueEntries } from "@/services/download-queue/fetch-queue-entries";

import React, { useCallback, useEffect, useRef, useState } from "react";

import DownloadQueueEntry from "@/components/shared/download-queue-entry";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { RefreshButton } from "@/components/shared/refresh-button";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { H3 } from "@/components/ui/typography";

import { cn } from "@/lib/cn";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

const DownloadQueuePage: React.FC = () => {
	// Refs - Fetching
	const isFetchingRef = useRef(false);

	// States - Queue Entries
	const [inProgressEntries, setInProgressEntries] = useState<DBMediaItemWithPosterSets[]>([]);
	const [errorEntries, setErrorEntries] = useState<DBMediaItemWithPosterSets[]>([]);
	const [warningEntries, setWarningEntries] = useState<DBMediaItemWithPosterSets[]>([]);

	// States - Loading & Error
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

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
	}, [fetchQueueEntries]);

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
