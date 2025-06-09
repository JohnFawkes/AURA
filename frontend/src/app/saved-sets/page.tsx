"use client";
import React, {
	useCallback,
	useEffect,
	useState,
	useRef,
	useMemo,
} from "react";
import { DBMediaItemWithPosterSets } from "@/types/databaseSavedSet";
import {
	AutodownloadResult,
	fetchAllItemsFromDB,
	postForceRecheckDBItemForAutoDownload,
} from "@/services/api.db";
import Loader from "@/components/ui/loader";
import ErrorMessage from "@/components/ui/error-message";
import SavedSetsCard from "@/components/ui/saved-sets-cards";
import { Button } from "@/components/ui/button";
import { RefreshCcw as RefreshIcon, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { useHomeSearchStore } from "@/lib/homeSearchStore";
import { searchMediaItems } from "@/hooks/searchMediaItems";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { log } from "@/lib/logger";
import { SelectItemsPerPage } from "@/components/items-per-page-select";
import { CustomPagination } from "@/components/custom-pagination";
import { RefreshButton } from "@/components/shared/buttons/refresh-button";

const SavedSetsPage: React.FC = () => {
	const [savedSets, setSavedSets] = useState<DBMediaItemWithPosterSets[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(false);
	const [errorMessage, setErrorMessage] = useState<string>("");
	const isFetchingRef = useRef(false);
	const { searchQuery } = useHomeSearchStore();
	const [filterAutoDownloadOnly, setFilterAutoDownloadOnly] = useState(false);
	const [recheckStatus, setRecheckStatus] = useState<
		Record<string, AutodownloadResult>
	>({});

	const { itemsPerPage } = useHomeSearchStore();
	const [currentPage, setCurrentPage] = useState(1);

	const fetchSavedSets = useCallback(async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;
		try {
			setLoading(true);
			const resp = await fetchAllItemsFromDB();
			if (resp.status !== "success") {
				throw new Error(resp.message);
			}
			const sets = resp.data;
			if (!sets) {
				throw new Error("No sets found");
			}
			setSavedSets(sets);
		} catch (error) {
			setError(true);
			setErrorMessage(
				error instanceof Error
					? error.message
					: "An unknown error occurred"
			);
		} finally {
			setLoading(false);
			isFetchingRef.current = false;
			toast.success("Saved sets fetched successfully", {
				id: "fetch-saved-sets-success",
				duration: 2000,
			});
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
			const filteredMediaItems = searchMediaItems(
				mediaItems,
				searchQuery
			);
			const filteredKeys = new Set(
				filteredMediaItems.map((item) => item.RatingKey)
			);
			filtered = savedSets.filter((set) =>
				filteredKeys.has(set.MediaItem.RatingKey)
			);
		}

		if (filterAutoDownloadOnly) {
			filtered = filtered.filter(
				(set) =>
					set.PosterSets &&
					set.PosterSets.some((ps) => ps.AutoDownload === true)
			);
		}

		const sorted = filtered.slice().sort((a, b) => {
			const getMaxDownloadTimestamp = (
				set: DBMediaItemWithPosterSets
			) => {
				if (!set.PosterSets || set.PosterSets.length === 0) return 0;
				return set.PosterSets.reduce((max, ps) => {
					const time = new Date(ps.LastDownloaded).getTime();
					return time > max ? time : max;
				}, 0);
			};

			const aMax = getMaxDownloadTimestamp(a);
			const bMax = getMaxDownloadTimestamp(b);
			return bMax - aMax;
		});

		return sorted;
	}, [savedSets, searchQuery, filterAutoDownloadOnly]);

	const paginatedSets = useMemo(() => {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		return filteredAndSortedSavedSets.slice(startIndex, endIndex);
	}, [currentPage, itemsPerPage, filteredAndSortedSavedSets]);

	const totalPages = Math.ceil(
		filteredAndSortedSavedSets.length / itemsPerPage
	);

	if (loading) {
		return <Loader className="mt-10" message="Loading saved sets..." />;
	}

	if (error) {
		return (
			<div className="flex flex-col items-center p-6 gap-4">
				<ErrorMessage message={errorMessage} />
			</div>
		);
	}

	const handleRecheckItem = async (
		title: string,
		item: DBMediaItemWithPosterSets
	): Promise<void> => {
		try {
			const recheckResp = await postForceRecheckDBItemForAutoDownload(
				item
			);
			if (recheckResp.status !== "success") {
				throw new Error(recheckResp.message);
			}
			setRecheckStatus((prev) => ({
				...prev,
				[title]: recheckResp.data as AutodownloadResult,
			}));
		} catch (error) {
			toast.error(
				error instanceof Error
					? error.message
					: "An unknown error occurred",
				{
					id: "recheck-error",
					duration: 2000,
				}
			);
		}
	};

	const forceRecheckAll = async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;

		setRecheckStatus({}); // Reset recheck status

		// Get all saved sets that have AutoDownload enabled
		const setsToRecheck = savedSets.filter(
			(set) =>
				set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)
		);

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
			toast.loading(
				`Rechecking ${index + 1} of ${setsToRecheck.length} - ${
					set.MediaItem.Title
				}`,
				{
					id: "force-recheck",
					duration: 0, // Keep it open until we manually close it
				}
			);
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
					<Label
						htmlFor="library-filter"
						className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4"
					>
						Filters:
					</Label>
					<Badge
						key={"filter-auto-download-only"}
						className="cursor-pointer mb-2 sm:mb-0 sm:mr-4"
						variant={filterAutoDownloadOnly ? "default" : "outline"}
						onClick={() => {
							setFilterAutoDownloadOnly(!filterAutoDownloadOnly);
						}}
					>
						{filterAutoDownloadOnly
							? "AutoDownload Only"
							: "All Items"}
					</Badge>
				</div>
				<Button
					variant="secondary"
					size="sm"
					onClick={() => forceRecheckAll()}
					className="flex items-center gap-1 text-xs sm:text-sm"
				>
					<span className="hidden sm:inline">
						Force Autodownload Recheck
					</span>
					<span className="sm:hidden">Recheck</span>
					<span className="whitespace-nowrap">
						(
						{
							savedSets.filter(
								(set) =>
									set.PosterSets &&
									set.PosterSets.some((ps) => ps.AutoDownload)
							).length
						}
						)
					</span>
					<RefreshIcon className="h-3 w-3" />
				</Button>
			</div>

			{/* Items Per Page Selection */}
			<div className="w-full flex items-center mb-2">
				<SelectItemsPerPage setCurrentPage={setCurrentPage} />
			</div>

			{Object.keys(recheckStatus).length > 0 && (
				<div className="w-full mb-4">
					<h3 className="text-lg font-semibold mb-2">
						Recheck Status:
					</h3>
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
									.sort(([, a], [, b]) =>
										a.MediaItemTitle.localeCompare(
											b.MediaItemTitle
										)
									)
									.map(([title, result]) => (
										<tr key={title}>
											<td className="px-3 py-2 text-sm">
												{result.MediaItemTitle}
											</td>
											<td className="px-3 py-2">
												<Badge
													className={cn(
														"inline-flex items-center rounded-full px-2 py-0.5 text-sm font-medium",
														{
															"bg-green-100 text-green-700":
																result.OverAllResult ===
																"Success",
															"bg-yellow-100 text-yellow-700":
																result.OverAllResult ===
																"Warning",
															"bg-red-100 text-red-700":
																result.OverAllResult ===
																"Error",
															"bg-gray-100 text-gray-700":
																result.OverAllResult ===
																"Skipped",
														}
													)}
												>
													{result.OverAllResult}
												</Badge>
											</td>
											<td className="px-3 py-2">
												<div className="space-y-1">
													<p className="text-md text-muted-foreground">
														{
															result.OverAllResultMessage
														}
													</p>
													{result.Sets.map(
														(set, index) => (
															<div
																key={`${set.PosterSetID}-${index}`}
																className="text-xs text-muted-foreground pl-4"
															>
																• Set{" "}
																{
																	set.PosterSetID
																}
																: {set.Result} -{" "}
																{set.Reason}
															</div>
														)
													)}
												</div>
											</td>
											<td className="px-3 py-2">
												<RefreshIcon
													className="h-4 w-4 cursor-pointer text-primary-dynamic hover:text-primary"
													onClick={async () => {
														const item =
															savedSets.find(
																(set) =>
																	set
																		.MediaItem
																		.Title ===
																	title
															);
														if (!item) return;
														// Remove the item from the recheck status list
														setRecheckStatus(
															(prev) => {
																const newStatus =
																	{
																		...prev,
																	};
																delete newStatus[
																	title
																];
																return newStatus;
															}
														);
														await handleRecheckItem(
															title,
															item
														);
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

			<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-2">
				{paginatedSets.length > 0 ? (
					paginatedSets.map((savedSet) => (
						<SavedSetsCard
							key={savedSet.MediaItem.RatingKey}
							savedSet={savedSet}
							onUpdate={fetchSavedSets}
							handleRecheckItem={handleRecheckItem}
						/>
					))
				) : (
					<p className="text-muted-foreground">
						No saved sets found.
					</p>
				)}
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
