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
import {
	Pagination,
	PaginationContent,
	PaginationEllipsis,
	PaginationItem,
	PaginationLink,
	PaginationNext,
	PaginationPrevious,
} from "@/components/ui/pagination";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

const SavedSetsPage: React.FC = () => {
	const [savedSets, setSavedSets] = useState<DBMediaItemWithPosterSets[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(false);
	const [errorMessage, setErrorMessage] = useState<string>("");
	const isFetchingRef = useRef(false);
	const { searchQuery } = useHomeSearchStore();
	const [filterAutoDownloadOnly, setFilterAutoDownloadOnly] = useState(false);
	const [recheckStatus, setRecheckStatus] = useState<{
		[mediaID: string]: {
			status: "error" | "warning" | "success";
			messages: string[];
		};
	}>({});
	const { itemsPerPage, setItemsPerPage } = useHomeSearchStore();
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

	// Helper function to ensure status is of correct type
	type RecheckStatus = "error" | "warning" | "success";
	const validateStatus = (status: string): RecheckStatus => {
		if (
			status === "error" ||
			status === "warning" ||
			status === "success"
		) {
			return status;
		}
		return "error"; // Default fallback
	};

	const forceRecheck = async () => {
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
				id: "no-sets-to-recheck",
				duration: 2000,
			});
		}

		// Show loading toast
		toast.loading("Rechecking sets...", {
			id: "recheck-loading",
			duration: 0, // Keep it open until we manually close it
		});

		// Go through each set and recheck
		for (const set of setsToRecheck) {
			try {
				const recheckResp = await postForceRecheckDBItemForAutoDownload(
					set
				);
				if (
					recheckResp.status !== "success" &&
					recheckResp.status !== "warning"
				) {
					throw new Error(recheckResp.data || "Unknown error");
				}

				setRecheckStatus((prev) => ({
					...prev,
					[set.MediaItem.RatingKey]: {
						status: validateStatus(recheckResp.status),
						messages: Array.isArray(recheckResp.data)
							? recheckResp.data
							: typeof recheckResp.data === "string"
							? recheckResp.data
									.split(",")
									.map((msg) => msg.trim())
							: ["Recheck successful"],
					},
				}));
			} catch (error) {
				console.error("Recheck error:", error);
				setRecheckStatus((prev) => ({
					...prev,
					[set.MediaItem.RatingKey]: {
						status: "error" as const,
						messages: [
							error instanceof Error
								? error.message
								: "An unknown error occurred",
						],
					},
				}));
			}
		}

		// Close loading toast
		toast.dismiss("recheck-loading");
		toast.success("Recheck completed", {
			id: "recheck-complete",
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
					onClick={() => forceRecheck()}
				>
					Force Autodownload Recheck (
					{
						savedSets.filter(
							(set) =>
								set.PosterSets &&
								set.PosterSets.some((ps) => ps.AutoDownload)
						).length
					}
					)
					<RefreshIcon className="h-3 w-3 ml-1" />
				</Button>
			</div>

			{/* Items Per Page Selection */}
			<div className="w-full flex items-center mb-2">
				<Label
					htmlFor="items-per-page-trigger"
					className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4"
				>
					Items per page:
				</Label>
				<Select
					value={itemsPerPage.toString()}
					onValueChange={(value) => {
						const newItemsPerPage = parseInt(value);
						if (!isNaN(newItemsPerPage)) {
							setItemsPerPage(newItemsPerPage);
							setCurrentPage(1);
						}
					}}
				>
					<SelectTrigger id="items-per-page-trigger">
						<SelectValue placeholder="Select" />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="10">10</SelectItem>
						<SelectItem value="20">20</SelectItem>
						<SelectItem value="30">30</SelectItem>
						<SelectItem value="50">50</SelectItem>
						<SelectItem value="100">100</SelectItem>
						<SelectScrollUpButton />
						<SelectScrollDownButton />
					</SelectContent>
				</Select>
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
										Messages
									</th>

									<th
										className="px-3 py-2 text-left text-sm font-medium text-muted-foreground"
										onClick={() => setRecheckStatus({})}
									>
										<Button
											variant="ghost"
											size="icon"
											onClick={() => {
												setRecheckStatus({});
											}}
											aria-label="Clear Recheck Status"
											className="hover:bg-muted/50 active:bg-muted/70 disabled:opacity-50 disabled:pointer-events-none"
											disabled={
												Object.keys(recheckStatus)
													.length === 0
											}
										>
											<XCircle className="h-4 w-4 text-muted-foreground" />
										</Button>
									</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border">
								{Object.entries(recheckStatus).map(
									([ratingKey, status]) => (
										<tr
											key={ratingKey}
											className={cn(
												"hover:bg-muted/50",
												status.status === "success" &&
													"bg-green-50/50",
												status.status === "warning" &&
													"bg-yellow-50/50",
												status.status === "error" &&
													"bg-red-50/50"
											)}
										>
											<td className="px-3 py-1.5 text-sm font-medium whitespace-nowrap">
												{savedSets.find(
													(set) =>
														set.MediaItem
															.RatingKey ===
														ratingKey
												)?.MediaItem.Title ||
													"Unknown Title"}
											</td>
											<td className="px-3 py-1.5 text-sm whitespace-nowrap">
												<span
													className={cn(
														"inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
														status.status ===
															"success" &&
															"bg-green-100 text-green-700",
														status.status ===
															"warning" &&
															"bg-yellow-100 text-yellow-700",
														status.status ===
															"error" &&
															"bg-red-100 text-red-700"
													)}
												>
													{status.status
														.charAt(0)
														.toUpperCase() +
														status.status.slice(1)}
												</span>
											</td>
											<td className="px-3 py-1.5 text-sm">
												<div className="flex flex-col gap-1">
													{status.messages.map(
														(msg, index) => (
															<div
																key={index}
																className={cn(
																	"flex items-start gap-2",
																	status.status ===
																		"success" &&
																		"text-green-700",
																	status.status ===
																		"warning" &&
																		"text-yellow-700",
																	status.status ===
																		"error" &&
																		"text-red-700"
																)}
															>
																<span className="mt-1">
																	•
																</span>
																<span className="flex-1">
																	{msg}
																</span>
															</div>
														)
													)}
												</div>
											</td>
											<td className="px-3 py-1.5 text-sm whitespace-nowrap"></td>
										</tr>
									)
								)}
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
						/>
					))
				) : (
					<p className="text-muted-foreground">
						No saved sets found.
					</p>
				)}
			</div>

			{/* Pagination */}
			<div className="flex justify-center mt-8">
				<Pagination>
					<PaginationContent>
						{/* Previous Page Button */}
						{totalPages > 1 && currentPage > 1 && (
							<PaginationItem>
								<PaginationPrevious
									onClick={() => {
										const newPage = Math.max(
											currentPage - 1,
											1
										);
										setCurrentPage(newPage);
										window.scrollTo({
											top: 0,
											behavior: "smooth",
										});
									}}
								/>
							</PaginationItem>
						)}

						{/* Page Input */}
						<PaginationItem>
							<div className="flex items-center gap-2">
								<Input
									type="number"
									min={1}
									max={totalPages}
									value={currentPage}
									onChange={(e) => {
										const value = parseInt(e.target.value);
										if (
											!isNaN(value) &&
											value >= 1 &&
											value <= totalPages
										) {
											setCurrentPage(value);
											window.scrollTo({
												top: 0,
												behavior: "smooth",
											});
										}
									}}
									className="w-16 h-8 text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
									onBlur={(e) => {
										const value = parseInt(e.target.value);
										if (isNaN(value) || value < 1) {
											setCurrentPage(1);
										} else if (value > totalPages) {
											setCurrentPage(totalPages);
										}
									}}
								/>
							</div>
						</PaginationItem>

						{/* Next Page Button */}
						{totalPages > 1 && currentPage < totalPages && (
							<PaginationItem>
								<PaginationNext
									onClick={() => {
										const newPage = Math.min(
											currentPage + 1,
											totalPages
										);
										setCurrentPage(newPage);
										window.scrollTo({
											top: 0,
											behavior: "smooth",
										});
									}}
								/>
							</PaginationItem>
						)}

						{/* Ellipsis and End Page */}
						{totalPages > 3 && currentPage < totalPages - 1 && (
							<>
								<PaginationItem>
									<PaginationEllipsis />
								</PaginationItem>
								<PaginationItem>
									<PaginationLink
										onClick={() => {
											setCurrentPage(totalPages);
											window.scrollTo({
												top: 0,
												behavior: "smooth",
											});
										}}
									>
										{totalPages}
									</PaginationLink>
								</PaginationItem>
							</>
						)}
					</PaginationContent>
				</Pagination>
			</div>

			<Button
				variant="outline"
				size="sm"
				className={cn(
					"fixed z-100 right-3 bottom-10 sm:bottom-15 rounded-full shadow-lg transition-all duration-300 bg-background border-primary-dynamic text-primary-dynamic hover:bg-primary-dynamic hover:text-primary cursor-pointer"
				)}
				onClick={() => fetchSavedSets()}
				aria-label="refresh"
			>
				<RefreshIcon className="h-3 w-3 mr-1" />
				<span className="text-xs hidden sm:inline">Refresh</span>
			</Button>
		</div>
	);
};

export default SavedSetsPage;
