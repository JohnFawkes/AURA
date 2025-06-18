"use client";

import { AutodownloadResult, fetchAllItemsFromDB, postForceRecheckDBItemForAutoDownload } from "@/services/api.db";
import { ReturnErrorMessage } from "@/services/api.shared";
import {
	ArrowDownAZ,
	ArrowDownZA,
	ClockArrowDown,
	ClockArrowUp,
	RefreshCcw as RefreshIcon,
	XCircle,
} from "lucide-react";
import { toast } from "sonner";

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import { SelectItemsPerPage } from "@/components/shared/items-per-page-select";
import Loader from "@/components/shared/loader";
import { RefreshButton } from "@/components/shared/refresh-button";
import SavedSetsCard from "@/components/shared/saved-sets-cards";
import { SortControl } from "@/components/shared/sort-control";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

import { usePageStore, useSavedSetsPageStore, useSearchQueryStore } from "@/lib/homeSearchStore";
import { log } from "@/lib/logger";
import { cn } from "@/lib/utils";

import { searchMediaItems } from "@/hooks/searchMediaItems";

import { APIResponse } from "@/types/apiResponse";
import { DBMediaItemWithPosterSets } from "@/types/databaseSavedSet";

const SavedSetsPage: React.FC = () => {
	const [savedSets, setSavedSets] = useState<DBMediaItemWithPosterSets[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const isFetchingRef = useRef(false);
	const { searchQuery } = useSearchQueryStore();
	const [filterAutoDownloadOnly, setFilterAutoDownloadOnly] = useState(false);
	const [recheckStatus, setRecheckStatus] = useState<Record<string, AutodownloadResult>>({});

	const { itemsPerPage } = usePageStore();
	const [currentPage, setCurrentPage] = useState(1);

	// State to track the selected sorting option
	const { sortOption, setSortOption } = useSavedSetsPageStore();
	const { sortOrder, setSortOrder } = useSavedSetsPageStore();

	// Set sortOption to "date" if its not title or date
	if (sortOption !== "title" && sortOption !== "date") {
		setSortOption("date");
	}

	const fetchSavedSets = useCallback(async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;
		try {
			setLoading(true);
			const response = await fetchAllItemsFromDB();

			if (response.status === "error") {
				setError(response);
				setSavedSets([]);
				return;
			}

			if (!response.data) {
				setError(ReturnErrorMessage<unknown>(new Error("No sets found")));
				setSavedSets([]);
				return;
			}

			setSavedSets(response.data);
			setError(null);

			toast.success("Saved sets fetched successfully");
		} catch (error) {
			setError(ReturnErrorMessage<unknown>(error));
			setSavedSets([]);
		} finally {
			setLoading(false);
			isFetchingRef.current = false;
		}
	}, []);

	useEffect(() => {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = "Aura | Saved Sets";
		}
		fetchSavedSets();
	}, [fetchSavedSets]);

	// This useMemo will first filter the savedSets using your search logic,
	// then sort the resulting array from newest to oldest using the LastDownloaded values.
	const filteredAndSortedSavedSets = useMemo(() => {
		let filtered = savedSets;

		if (searchQuery.trim() !== "") {
			const mediaItems = savedSets.map((set) => set.MediaItem);
			const filteredMediaItems = searchMediaItems(mediaItems, searchQuery);
			const filteredKeys = new Set(filteredMediaItems.map((item) => item.RatingKey));
			filtered = savedSets.filter((set) => filteredKeys.has(set.MediaItem.RatingKey));
		}

		if (filterAutoDownloadOnly) {
			filtered = filtered.filter(
				(set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload === true)
			);
		}

		// Sort the filtered sets based on the selected sort option and order
		const sorted = filtered.slice();
		if (sortOption === "title") {
			if (sortOrder === "asc") {
				sorted.sort((a, b) => a.MediaItem.Title.localeCompare(b.MediaItem.Title));
			} else {
				sorted.sort((a, b) => b.MediaItem.Title.localeCompare(a.MediaItem.Title));
			}
		} else if (sortOption === "date") {
			const getMaxDownloadTimestamp = (set: DBMediaItemWithPosterSets) => {
				if (!set.PosterSets || set.PosterSets.length === 0) return 0;
				return set.PosterSets.reduce((max, ps) => {
					const time = new Date(ps.LastDownloaded).getTime();
					return time > max ? time : max;
				}, 0);
			};

			sorted.sort((a, b) => {
				const aMax = getMaxDownloadTimestamp(a);
				const bMax = getMaxDownloadTimestamp(b);
				return sortOrder === "asc" ? aMax - bMax : bMax - aMax;
			});
		}

		return sorted;
	}, [savedSets, searchQuery, filterAutoDownloadOnly, sortOption, sortOrder]);

	const paginatedSets = useMemo(() => {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		return filteredAndSortedSavedSets.slice(startIndex, endIndex);
	}, [currentPage, itemsPerPage, filteredAndSortedSavedSets]);

	const totalPages = Math.ceil(filteredAndSortedSavedSets.length / itemsPerPage);

	if (loading) {
		return <Loader className="mt-10" message="Loading saved sets..." />;
	}

	if (error) {
		return (
			<div className="flex flex-col items-center p-6 gap-4">
				<ErrorMessage error={error} />
			</div>
		);
	}

	const handleRecheckItem = async (title: string, item: DBMediaItemWithPosterSets): Promise<void> => {
		try {
			const response = await postForceRecheckDBItemForAutoDownload(item);

			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to recheck item");
				return;
			}

			setRecheckStatus((prev) => ({
				...prev,
				[title]: response.data as AutodownloadResult,
			}));
		} catch (error) {
			const errorResponse = ReturnErrorMessage<unknown>(error);
			toast.error(errorResponse.error?.Message || "An unexpected error occurred");
		}
	};

	const forceRecheckAll = async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;

		setRecheckStatus({}); // Reset recheck status

		// Get all saved sets that have AutoDownload enabled
		const setsToRecheck = savedSets.filter((set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload));

		log("Sets to recheck:", setsToRecheck);

		if (setsToRecheck.length === 0) {
			toast.warning("No sets with AutoDownload enabled found", {
				id: "force-recheck",
				duration: 2000,
			});
		}

		// Show loading toast
		toast.loading(`Rechecking ${setsToRecheck.length} sets...`, {
			id: "force-recheck",
			duration: 0, // Keep it open until we manually close it
		});

		for (const [index, set] of setsToRecheck.entries()) {
			toast.loading(`Rechecking ${index + 1} of ${setsToRecheck.length} - ${set.MediaItem.Title}`, {
				id: "force-recheck",
				duration: 0, // Keep it open until we manually close it
			});
			await handleRecheckItem(set.MediaItem.Title, set);
		}

		// Close loading toast
		toast.success("Recheck completed", {
			id: "force-recheck",
			duration: 2000,
		});

		isFetchingRef.current = false;
	};

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			<div className="w-full flex items-center justify-between mb-4">
				<div className="flex items-center gap-2">
					<Label htmlFor="library-filter" className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4">
						Filters:
					</Label>
					<Badge
						key={"filter-auto-download-only"}
						className="cursor-pointer text-sm mb-2 sm:mb-0 sm:mr-4"
						variant={filterAutoDownloadOnly ? "default" : "outline"}
						onClick={() => {
							setFilterAutoDownloadOnly(!filterAutoDownloadOnly);
						}}
					>
						{filterAutoDownloadOnly ? "AutoDownload Only" : "All Items"}
					</Badge>
				</div>
				{
					// Only show the force recheck button if there are sets with AutoDownload enabled
					savedSets.some((set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)) && (
						<Button
							variant="secondary"
							size="sm"
							onClick={() => forceRecheckAll()}
							className="flex items-center gap-1 text-xs sm:text-sm"
						>
							<span className="hidden sm:inline">Force Autodownload Recheck</span>
							<span className="sm:hidden">Recheck</span>
							<span className="whitespace-nowrap">
								(
								{
									savedSets.filter(
										(set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)
									).length
								}
								)
							</span>
							<RefreshIcon className="h-3 w-3" />
						</Button>
					)
				}
			</div>

			{/* Sorting controls */}
			<div className="w-full flex items-center mb-2">
				<SortControl
					options={[
						{
							value: "date",
							label: "Date Updated",
							ascIcon: <ClockArrowUp />,
							descIcon: <ClockArrowDown />,
						},
						{ value: "title", label: "Title", ascIcon: <ArrowDownAZ />, descIcon: <ArrowDownZA /> },
					]}
					sortOption={sortOption}
					sortOrder={sortOrder}
					setSortOption={(value) => {
						setSortOption(value as "title" | "date" | "");
						if (value === "title") setSortOrder("asc");
						else if (value === "date") setSortOrder("desc");
					}}
					setSortOrder={setSortOrder}
				/>
			</div>
			{/* Items Per Page Selection */}
			{
				// Only show the items per page selection if there are more than 10 sets
				filteredAndSortedSavedSets.length > 10 && (
					<div className="w-full flex items-center mb-2">
						<SelectItemsPerPage setCurrentPage={setCurrentPage} />
					</div>
				)
			}

			{Object.keys(recheckStatus).length > 0 && (
				<div className="w-full mb-4">
					<h3 className="text-lg font-semibold mb-2">Recheck Status:</h3>
					<div className="rounded-md border">
						<table className="w-full divide-y divide-border">
							<thead className="bg-muted/50">
								<tr>
									<th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground w-[200px]">
										Title
									</th>
									<th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground w-[80px]">
										Status
									</th>
									<th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground">
										Details
									</th>
									<th className="px-3 py-2 text-right">
										<Button
											variant="ghost"
											size="icon"
											onClick={() => setRecheckStatus({})}
											className="h-8 w-8"
										>
											<XCircle className="h-4 w-4" />
										</Button>
									</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border">
								{Object.entries(recheckStatus)
									// Sort entries by MediaItemTitle
									.sort(([, a], [, b]) => a.MediaItemTitle.localeCompare(b.MediaItemTitle))
									.map(([title, result]) => (
										<tr key={title}>
											<td className="px-3 py-2 text-sm">{result.MediaItemTitle}</td>
											<td className="px-3 py-2">
												<Badge
													className={cn(
														"inline-flex items-center rounded-full px-2 py-0.5 text-sm font-medium",
														{
															"bg-green-100 text-green-700":
																result.OverAllResult === "Success",
															"bg-yellow-100 text-yellow-700":
																result.OverAllResult === "Warning",
															"bg-red-100 text-red-700": result.OverAllResult === "Error",
															"bg-gray-100 text-gray-700":
																result.OverAllResult === "Skipped",
														}
													)}
												>
													{result.OverAllResult}
												</Badge>
											</td>
											<td className="px-3 py-2">
												<div className="space-y-1">
													<p className="text-md text-muted-foreground">
														{result.OverAllResultMessage}
													</p>
													{result.Sets.map((set, index) => (
														<div
															key={`${set.PosterSetID}-${index}`}
															className="text-xs text-muted-foreground pl-4"
														>
															• Set {set.PosterSetID}: {set.Result} - {set.Reason}
														</div>
													))}
												</div>
											</td>
											<td className="px-3 py-2">
												<RefreshIcon
													className="h-4 w-4 cursor-pointer text-primary-dynamic hover:text-primary"
													onClick={async () => {
														const item = savedSets.find(
															(set) => set.MediaItem.Title === title
														);
														if (!item) return;
														// Remove the item from the recheck status list
														setRecheckStatus((prev) => {
															const newStatus = {
																...prev,
															};
															delete newStatus[title];
															return newStatus;
														});
														await handleRecheckItem(title, item);
													}}
												/>
											</td>
										</tr>
									))}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{/* If there are no saved sets, show a message */}
			{filteredAndSortedSavedSets.length === 0 && !loading && !error && !Object.keys(recheckStatus).length && (
				<div className="w-full">
					<ErrorMessage error={ReturnErrorMessage<string>("No saved sets found")} />
				</div>
			)}

			<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-2">
				{paginatedSets.length > 0 &&
					paginatedSets.map((savedSet) => (
						<SavedSetsCard
							key={savedSet.MediaItem.RatingKey}
							savedSet={savedSet}
							onUpdate={fetchSavedSets}
							handleRecheckItem={handleRecheckItem}
						/>
					))}
			</div>

			{/* Pagination */}
			<CustomPagination
				currentPage={currentPage}
				totalPages={totalPages}
				setCurrentPage={setCurrentPage}
				scrollToTop={true}
			/>

			{/* Refresh Button */}
			<RefreshButton onClick={fetchSavedSets} />
		</div>
	);
};

export default SavedSetsPage;
