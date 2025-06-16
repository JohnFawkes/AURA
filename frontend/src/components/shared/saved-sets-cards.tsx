"use client";

import { formatMediaItemUrl } from "@/helper/formatMediaItemURL";
import { deleteMediaItemFromDB, patchSavedItemInDB } from "@/services/api.db";
import { fetchSetByID } from "@/services/api.mediux";
import { Delete, Edit, MoreHorizontal, RefreshCcw, RefreshCwOff } from "lucide-react";

import React, { useState } from "react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import DownloadModal from "@/components/shared/download-modal";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { H4, P } from "@/components/ui/typography";

import { log } from "@/lib/logger";
import { useMediaStore } from "@/lib/mediaStore";
import { usePosterSetsStore } from "@/lib/posterSetStore";

import { APIResponse } from "@/types/apiResponse";
import { DBMediaItemWithPosterSets } from "@/types/databaseSavedSet";
import { PosterSet } from "@/types/posterSets";

import { Badge } from "../ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "../ui/dropdown-menu";
import { Separator } from "../ui/separator";

const SavedSetsCard: React.FC<{
	savedSet: DBMediaItemWithPosterSets;
	onUpdate: () => void;
	handleRecheckItem: (title: string, item: DBMediaItemWithPosterSets) => void;
}> = ({ savedSet, onUpdate, handleRecheckItem }) => {
	// Initialize edit state from the savedSet.PosterSets array.
	const [editSets, setEditSets] = useState(() =>
		savedSet.PosterSets.map((set) => ({
			id: set.PosterSetID,
			set: set.PosterSet || "Unknown",
			selectedTypes: set.SelectedTypes,
			autoDownload: set.AutoDownload,
			toDelete: false,
		}))
	);

	const [isEditModalOpen, setIsEditModalOpen] = useState(false);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	// State to track any error messages during updates.
	const [updateError, setUpdateError] = useState<APIResponse<unknown> | null>(null);
	const [isMounted, setIsMounted] = useState(false);
	const { setPosterSets, setSetAuthor, setSetID, setSetTitle, setSetType } = usePosterSetsStore();
	const { setMediaItem } = useMediaStore();
	const router = useRouter();
	const [isRefreshing, setIsRefreshing] = useState(false);

	const allToDelete = editSets.every((set) => set.toDelete);

	const onClose = () => {
		setIsEditModalOpen(false);
		setIsDeleteModalOpen(false);
		setUpdateError(null);
		setIsMounted(false);
	};

	const confirmEdit = async () => {
		if (isMounted) return;
		setIsMounted(true);

		if (allToDelete) {
			setIsEditModalOpen(false);
			setIsDeleteModalOpen(true);
			setUpdateError(null);
			setIsMounted(false);
			return;
		}

		// Create a new DBSavedItem object with updated values
		const updatedSavedSet: DBMediaItemWithPosterSets = {
			...savedSet,
			PosterSets: editSets.map((editSet, idx) => ({
				PosterSetID: editSet.id,
				PosterSet: editSet.set,
				PosterSetJSON:
					typeof editSet.set === "object"
						? JSON.stringify(editSet.set)
						: savedSet.PosterSets[idx]?.PosterSetJSON || "",
				LastDownloaded: new Date().toISOString(),
				SelectedTypes: editSet.selectedTypes,
				AutoDownload: editSet.autoDownload,
				ToDelete: editSet.toDelete,
			})),
		};

		const response = await patchSavedItemInDB(updatedSavedSet);
		if (!response || response.status === "error") {
			log("Error updating saved set:", response?.error?.Message || "Unknown error");
			setUpdateError(response);
			setIsMounted(false);
			return;
		}

		setUpdateError(null);
		setIsEditModalOpen(false);
		onUpdate();

		setIsMounted(false);
	};

	const confirmDelete = async () => {
		if (isMounted) return;
		setIsMounted(true);
		const resp = await deleteMediaItemFromDB(savedSet.MediaItemID);
		if (!resp || resp.status === "error") {
			log("Error deleting saved set:", resp?.error?.Message || "Unknown error");
			setUpdateError(resp);
			setIsMounted(false);
			return;
		}
		setUpdateError(null);
		setIsDeleteModalOpen(false);
		onUpdate();
		setIsMounted(false);
	};

	const renderSetList = () => {
		return (
			<div className="w-full">
				<span
					className="text-sm text-muted-foreground mb-1 block"
					onClick={() => {
						log("SET: ", savedSet);
					}}
				>
					{savedSet.PosterSets.length > 1 ? "Sets:" : "Set:"}
				</span>
				<table className="w-full text-sm">
					<tbody>
						{savedSet.PosterSets.map((set) => (
							<tr key={set.PosterSetID} className={`hover:bg-muted/50 rounded-sm`}>
								<td className="py-1.5" style={{ width: "80%" }}>
									<Link
										href={`/sets/${set.PosterSetID}`}
										className="text-primary hover:underline"
										onClick={() => {
											setPosterSets([set.PosterSet]);
											setSetType(set.PosterSet.Type);
											setSetTitle(set.PosterSet.Title);
											setSetAuthor(set.PosterSet.User.Name);
											setSetID(set.PosterSetID);
										}}
									>
										{set.PosterSetID}
									</Link>
								</td>

								<td className="py-1.5" style={{ width: "50%" }}>
									<Link
										href={`/user/${set.PosterSet.User.Name}`}
										className="text-primary hover:underline"
									>
										{set.PosterSet.User.Name || ""}
									</Link>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		);
	};

	// Replace the hard-coded array with dynamically generated list.
	const renderEditTypeBadges = (editSet: (typeof editSets)[number], index: number) => {
		const availableTypes: string[] = [];
		if (editSet.set && editSet.set.Poster) {
			availableTypes.push("poster");
		}
		if (editSet.set && editSet.set.Backdrop) {
			availableTypes.push("backdrop");
		}
		if (editSet.set && editSet.set.SeasonPosters && editSet.set.SeasonPosters.length > 0) {
			// Check to see if any of the Season Posters are Season 0
			const hasSeason0 = editSet.set.SeasonPosters.some((season) => season.Season?.Number === 0);
			if (hasSeason0) {
				availableTypes.push("specialSeasonPoster");
			}
			// Check to see if any of the Season Posters are not Season 0
			const hasNonSeason0 = editSet.set.SeasonPosters.some((season) => season.Season?.Number !== 0);
			if (hasNonSeason0) {
				availableTypes.push("seasonPoster");
			}
		}
		if (editSet.set && editSet.set.TitleCards && editSet.set.TitleCards.length > 0) {
			availableTypes.push("titlecard");
		}

		return availableTypes.map((type) => {
			const isSelected = editSet.selectedTypes.includes(type);
			// Disable the type if this set is marked for deletion
			// or if it's not selected here, but found in any other set.
			const isTypeDisabled =
				editSet.toDelete ||
				(!isSelected && editSets.some((item, j) => j !== index && item.selectedTypes.includes(type)));
			return (
				<Badge
					key={type}
					className={`flex items-center gap-2 transition duration-200 ${
						isTypeDisabled
							? "bg-secondary opacity-50 cursor-not-allowed"
							: isSelected
								? "cursor-pointer bg-primary text-primary-foreground hover:bg-red-500"
								: "cursor-pointer bg-secondary text-secondary-foreground"
					}`}
					onClick={() => {
						if (isTypeDisabled) return;
						setEditSets((prev) =>
							prev.map((item, i) => {
								if (i !== index) return item;
								const newSelectedTypes = item.selectedTypes.includes(type)
									? item.selectedTypes.filter((t) => t !== type)
									: [...item.selectedTypes, type];
								return {
									...item,
									selectedTypes: newSelectedTypes,
								};
							})
						);
					}}
				>
					{type === "poster"
						? "Poster"
						: type === "backdrop"
							? "Backdrop"
							: type === "seasonPoster"
								? "Season Posters"
								: type === "specialSeasonPoster"
									? "Special Poster"
									: type === "titlecard"
										? "Title Card"
										: type}
				</Badge>
			);
		});
	};

	const renderTypeBadges = () => {
		// Flatten all SelectedTypes arrays from every poster set.
		const allTypes = savedSet.PosterSets.flatMap((set) => set.SelectedTypes);
		const uniqueTypes = Array.from(new Set(allTypes));
		return uniqueTypes.map((type) =>
			// Check if the type is empty or not
			type.trim() === "" ? null : (
				// Render the badge only if the type is not empty
				<Badge key={type}>
					{type === "poster"
						? "Poster"
						: type === "backdrop"
							? "Backdrop"
							: type === "seasonPoster"
								? "Season Posters"
								: type === "specialSeasonPoster"
									? "Special Poster"
									: type === "titlecard"
										? "Title Card"
										: type}
				</Badge>
			)
		);
	};

	const refreshPosterSet = async () => {
		try {
			setIsRefreshing(true);
			// Track if any requests failed
			let hasError = false;
			let errorResponse: APIResponse<unknown> | null = null;

			await Promise.all(
				editSets.map(async (set) => {
					const response = await fetchSetByID(
						savedSet.MediaItem.LibraryTitle,
						savedSet.MediaItem.RatingKey,
						set.id,
						set.set.Type
					);
					if (!response || response.status === "error") {
						log("Error fetching set by ID:", response?.error?.Message || "Unknown error");
						// Store the first error we encounter
						if (!hasError) {
							hasError = true;
							errorResponse = response;
						}
						return;
					}
					if (!response.data) {
						log("No PosterSet found in response:", response);
						return;
					}
					const posterSet: PosterSet = response.data;
					setEditSets((prev) =>
						prev.map((item) =>
							item.id === set.id
								? {
										...item,
										set: posterSet,
									}
								: item
						)
					);
				})
			);

			// Update error state based on API calls results
			if (hasError) {
				setUpdateError(errorResponse);
			} else {
				setUpdateError(null);
			}
		} catch (error) {
			// Handle unexpected errors
			setUpdateError({
				status: "error",
				elapsed: "0",
				error: {
					Message: "Unexpected error refreshing poster sets",
					HelpText: "Please try again later",
					Function: "refreshPosterSet",
					LineNumber: 0,
					Details: error instanceof Error ? error.message : String(error),
				},
			});
		} finally {
			setIsRefreshing(false);
		}
	};

	return (
		<Card className="relative w-full max-w-md mx-auto mb-4">
			<CardHeader>
				{isRefreshing && (
					<div className="absolute inset-0 flex items-center justify-center bg-black/50 z-10">
						<Loader className="animate-spin h-8 w-8 text-primary" />
						<span className="ml-2 text-white">Refreshing sets...</span>
					</div>
				)}

				{/* Top Left: Auto Download Icon */}
				{savedSet.MediaItem.Type === "show" && (
					<div className="absolute top-2 left-2">
						{savedSet.PosterSets.some((set) => set.AutoDownload) ? (
							<RefreshCcw className="text-green-500" size={24} />
						) : (
							<RefreshCwOff className="text-red-500" size={24} />
						)}
					</div>
				)}
				{/* Top Right: Dropdown Menu */}
				<div className="absolute top-2 right-2">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="ghost" size="icon">
								<MoreHorizontal />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								onClick={async () => {
									await refreshPosterSet();
									setIsEditModalOpen(true);
								}}
							>
								<Edit className="ml-2" />
								{isRefreshing ? "Refreshing..." : "Edit"}
							</DropdownMenuItem>

							<DropdownMenuItem
								onClick={() => {
									handleRecheckItem(savedSet.MediaItem.Title, savedSet);
								}}
							>
								<RefreshCcw className="ml-2" />
								Force Autodownload Recheck
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => setIsDeleteModalOpen(true)} className="text-destructive">
								<Delete className="ml-2" />
								Delete
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
				{/* Middle: Image */}
				<div className="flex justify-center mt-6">
					<AssetImage
						image={savedSet.MediaItem}
						className="w-[170px] h-auto transition-transform hover:scale-105"
					/>
				</div>
			</CardHeader>

			{/* Content */}
			<CardContent>
				{/* Title */}
				<H4
					className="text-lg font-semibold cursor-pointer hover:underline"
					onClick={() => {
						setMediaItem(savedSet.MediaItem);
						router.push(formatMediaItemUrl(savedSet.MediaItem));
					}}
				>
					{savedSet.MediaItem.Title}
				</H4>

				{/* Year */}
				<P className="text-sm text-muted-foreground">Year: {savedSet.MediaItem.Year}</P>

				{/* Library Title */}
				<P className="text-sm text-muted-foreground">Library: {savedSet.MediaItem.LibraryTitle}</P>

				{/* Last Updated */}
				<P className="text-sm text-muted-foreground">
					Last Updated:{" "}
					{(() => {
						const latestTimestamp = Math.max(
							...savedSet.PosterSets.map((ps) => new Date(ps.LastDownloaded).getTime())
						);
						const latestDate = new Date(latestTimestamp);
						return `${latestDate.toLocaleDateString("en-US")} at ${latestDate.toLocaleTimeString("en-US", {
							hour: "numeric",
							minute: "numeric",
							second: "numeric",
							hour12: true,
						})}`;
					})()}
				</P>

				<div className="flex flex-wrap gap-2">{renderSetList()}</div>

				<Separator className="my-4" />

				{savedSet.PosterSets.some(
					(set) => Array.isArray(set.SelectedTypes) && set.SelectedTypes.some((type) => type.trim() !== "")
				) ? (
					<div className="flex flex-wrap gap-2">{renderTypeBadges()}</div>
				) : (
					<P className="text-sm text-muted-foreground">No types selected.</P>
				)}
			</CardContent>

			{/* Edit Modal */}
			<Dialog open={isEditModalOpen} onOpenChange={onClose}>
				<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[500px] ">
					<DialogHeader>
						<DialogTitle>Edit Saved Set</DialogTitle>
						<DialogDescription>
							Edit each set individually. Toggle type badges to update selected types. Use the delete
							option to mark a set for deletion.
						</DialogDescription>
					</DialogHeader>

					<div className="space-y-4">
						{editSets.map((editSet, index) => (
							<div key={editSet.id} className="border p-2 rounded-md">
								<div className="flex items-center justify-between">
									<span className="font-semibold">
										<Link
											href={`https://mediux.pro/sets/${editSet.id}`}
											target="_blank"
											rel="noopener noreferrer"
											className="hover:underline"
										>
											{editSet.set.Title}
										</Link>
									</span>

									<Button
										variant={editSet.toDelete ? "destructive" : "outline"}
										size="sm"
										onClick={() => {
											setEditSets((prev) =>
												prev.map((item, i) =>
													i === index
														? {
																...item,
																toDelete: !item.toDelete,
																// Clear the selected types when marking for deletion.
																selectedTypes: !item.toDelete ? [] : item.selectedTypes,
															}
														: item
												)
											);
										}}
									>
										{editSet.toDelete ? "Undo Delete" : "Delete Set"}
									</Button>
								</div>

								{editSet.set.User.Name && (
									<DialogDescription>
										<span className="text-sm text-muted-foreground">
											By:{" "}
											<Link href={`/user/${editSet.set.User.Name}`} className="hover:underline">
												{editSet.set.User.Name}
											</Link>
										</span>
									</DialogDescription>
								)}

								<DialogDescription>
									Set ID:{" "}
									<Link
										href={`https://mediux.pro/sets/${editSet.id}`}
										target="_blank"
										rel="noopener noreferrer"
										className="hover:underline"
									>
										{editSet.id}
									</Link>
								</DialogDescription>

								<div className="flex flex-wrap gap-2 mt-2">{renderEditTypeBadges(editSet, index)}</div>
								{savedSet.MediaItem.Type === "show" && (
									<div className="flex flex-wrap gap-2 mt-2">
										<Badge
											className={`cursor-pointer transition duration-200 ${
												editSet.autoDownload
													? "bg-primary text-primary-foreground hover:bg-red-500"
													: "bg-secondary text-secondary-foreground"
											}`}
											onClick={() => {
												setEditSets((prev) =>
													prev.map((item, i) =>
														i === index
															? {
																	...item,
																	autoDownload: !item.autoDownload,
																}
															: item
													)
												);
											}}
										>
											{editSet.autoDownload ? "Autodownload" : "No Autodownload"}
										</Badge>
									</div>
								)}
								<div className="flex items-center justify-end">
									<span className="text-md text-muted-foreground mt-2 mr-2">Redownload</span>
									<DownloadModal
										setID={editSet.id}
										setTitle={editSet.set.Title}
										setType={editSet.set.Type}
										setAuthor={editSet.set.User.Name}
										posterSets={[editSet.set]}
										autoDownloadDefault={editSet.autoDownload}
									/>
								</div>
							</div>
						))}
					</div>

					{updateError && <ErrorMessage error={updateError} />}
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => {
								setEditSets(
									savedSet.PosterSets.map((set) => ({
										id: set.PosterSetID,
										set: set.PosterSet,
										selectedTypes: set.SelectedTypes,
										autoDownload: set.AutoDownload,
										toDelete: false,
									}))
								);
								onClose();
							}}
						>
							Cancel
						</Button>
						<Button variant={allToDelete ? "destructive" : "default"} onClick={confirmEdit}>
							{allToDelete ? (editSets.length === 1 ? "Delete Set" : "Delete All") : "Save"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Delete Confirmation Modal */}
			<Dialog open={isDeleteModalOpen} onOpenChange={onClose}>
				<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[500px] ">
					<DialogHeader>
						<DialogTitle>Confirm Delete</DialogTitle>
						<DialogDescription>
							Are you sure you want to delete all sets for "{savedSet.MediaItem.Title}
							"? This action cannot be undone.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button variant="destructive" onClick={confirmDelete}>
							Delete
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</Card>
	);
};

export default SavedSetsCard;
