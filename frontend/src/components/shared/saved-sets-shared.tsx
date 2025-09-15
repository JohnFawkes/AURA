import { deleteMediaItemFromDB, patchSavedItemInDB } from "@/services/api.db";
import { fetchMediaServerItemContent } from "@/services/api.mediaserver";
import { fetchSetByID } from "@/services/api.mediux";
import { User } from "lucide-react";

import React from "react";

import Link from "next/link";

import DownloadModal from "@/components/shared/download-modal";
import { ErrorMessage } from "@/components/shared/error-message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";
import { DBMediaItemWithPosterSets } from "@/types/databaseSavedSet";
import { PosterSet } from "@/types/posterSets";

export const refreshPosterSet = async ({
	editSets,
	setEditSets,
	savedSet,
	setIsRefreshing,
	setUpdateError,
}: {
	editSets: EditSet[];
	setEditSets: React.Dispatch<React.SetStateAction<EditSet[]>>;
	savedSet: DBMediaItemWithPosterSets;
	setIsRefreshing: (v: boolean) => void;
	setUpdateError: (v: APIResponse<unknown> | null) => void;
}) => {
	try {
		setIsRefreshing(true);
		// Track if any requests failed
		let hasError = false;
		let errorResponse: APIResponse<unknown> | null = null;

		await Promise.all(
			editSets.map(async (set) => {
				if (set.id === "ignore") {
					return;
				}

				// Update the media item in the backend store by calling fetchMediaServerItemContent
				const resp = await fetchMediaServerItemContent(
					savedSet.MediaItem.RatingKey,
					savedSet.MediaItem.LibraryTitle
				);

				if (!resp || resp.status === "error") {
					log("Error fetching media item content:", resp.error?.Message || "Unknown error");
					// Store the first error we encounter
					if (!hasError) {
						hasError = true;
						errorResponse = resp;
					}
					return;
				}

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
									previousDateUpdated: item.set.DateUpdated,
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

export const renderTypeBadges = (savedSet: DBMediaItemWithPosterSets) => {
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

export const handleStopIgnoring = async (
	savedSet: DBMediaItemWithPosterSets,
	onUpdate: () => void,
	unignoreLoading: boolean,
	setUnignoreLoading: (v: boolean) => void,
	setUpdateError: (v: APIResponse<unknown> | null) => void
) => {
	if (unignoreLoading) return;
	setUnignoreLoading(true);
	const resp = await deleteMediaItemFromDB(savedSet);
	if (!resp || resp.status === "error") {
		log("Error removing ignore placeholder:", resp?.error?.Message || "Unknown error");
		setUpdateError(resp);
		setUnignoreLoading(false);
		return;
	}
	setUpdateError(null);
	onUpdate();
	setUnignoreLoading(false);
};

export const onCloseSavedSetsEditDeleteModals = ({
	setIsEditModalOpen,
	setIsDeleteModalOpen,
	setUpdateError,
	setIsMounted,
}: {
	setIsEditModalOpen: (v: boolean) => void;
	setIsDeleteModalOpen: (v: boolean) => void;
	setUpdateError: (v: APIResponse<unknown> | null) => void;
	setIsMounted: (v: boolean) => void;
}) => {
	setIsEditModalOpen(false);
	setIsDeleteModalOpen(false);
	setUpdateError(null);
	setIsMounted(false);
};

export const savedSetsConfirmEdit = async ({
	editSets,
	savedSet,
	onUpdate,
	isMounted,
	setIsMounted,
	setUpdateError,
	setIsEditModalOpen,
	setIsDeleteModalOpen,
	allToDelete,
}: {
	editSets: EditSet[];
	savedSet: DBMediaItemWithPosterSets;
	onUpdate: () => void;
	isMounted: boolean;
	setIsMounted: (v: boolean) => void;
	setUpdateError: (v: APIResponse<unknown> | null) => void;
	setIsEditModalOpen: (v: boolean) => void;
	setIsDeleteModalOpen: (v: boolean) => void;
	allToDelete: boolean;
}) => {
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

export const savedSetsConfirmDelete = async ({
	savedSet,
	onUpdate,
	isMounted,
	setIsMounted,
	setUpdateError,
	setIsDeleteModalOpen,
}: {
	savedSet: DBMediaItemWithPosterSets;
	onUpdate: () => void;
	isMounted: boolean;
	setIsMounted: (v: boolean) => void;
	setUpdateError: (v: APIResponse<unknown> | null) => void;
	setIsDeleteModalOpen: (v: boolean) => void;
}) => {
	if (isMounted) return;
	setIsMounted(true);
	const resp = await deleteMediaItemFromDB(savedSet);
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

export interface EditSet {
	id: string;
	set: PosterSet;
	previousDateUpdated: string;
	selectedTypes: string[];
	autoDownload: boolean;
	toDelete: boolean;
}

export interface SavedSetEditModalProps {
	open: boolean;
	onClose: () => void;
	editSets: EditSet[];
	setEditSets: React.Dispatch<React.SetStateAction<EditSet[]>>;
	savedSet: DBMediaItemWithPosterSets;
	onlyIgnore: boolean;
	allToDelete: boolean;
	updateError: APIResponse<unknown> | null;
	confirmEdit: () => void;
}

export const SavedSetEditModal: React.FC<SavedSetEditModalProps> = ({
	open,
	onClose,
	editSets,
	setEditSets,
	savedSet,
	onlyIgnore,
	allToDelete,
	updateError,
	confirmEdit,
}) => {
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
			const hasSeason0 = editSet.set.SeasonPosters.some((season) => season.Season && season.Season.Number === 0);
			if (hasSeason0) {
				availableTypes.push("specialSeasonPoster");
			}
			// Check to see if any of the Season Posters are not Season 0
			const hasNonSeason0 = editSet.set.SeasonPosters.some(
				(season) => season.Season && season.Season.Number !== 0
			);
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

	return (
		<Dialog open={open} onOpenChange={onClose}>
			<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[500px] ">
				<DialogHeader>
					<DialogTitle>{onlyIgnore ? "Ignored Item" : "Edit Saved Set"}</DialogTitle>
					<DialogDescription>
						{onlyIgnore
							? "This item is currently ignored. You can stop ignoring it to allow poster sets to be added again."
							: "Edit each set individually. Toggle type badges to update selected types. Use the delete option to mark a set for deletion."}
					</DialogDescription>
				</DialogHeader>
				{onlyIgnore ? (
					<div className="space-y-4">
						<div className="border rounded-md p-4 bg-muted/40">
							<p className="text-sm text-muted-foreground">
								No poster sets are associated with this media item because it has been marked as
								ignored. Stopping ignore will remove the placeholder and allow you to add sets in the
								future.
							</p>
						</div>
					</div>
				) : (
					<div className="space-y-4">
						{editSets.map((editSet, index) => (
							<div key={editSet.id} className="border p-2 rounded-md">
								<div className="flex items-center justify-between">
									<span className="font-semibold">
										<Link
											href={`https://mediux.io/${editSet.set.Type}-set/${editSet.id}`}
											target="_blank"
											rel="noopener noreferrer"
											className="hover:underline ml-1"
										>
											{editSet.set.Title}
										</Link>
									</span>
									<Button
										variant={editSet.toDelete ? "destructive" : "outline"}
										size="sm"
										className="cursor-pointer"
										onClick={() => {
											setEditSets((prev) =>
												prev.map((item, i) =>
													i === index
														? {
																...item,
																toDelete: !item.toDelete,
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
								{editSet.set.User?.Name && (
									<DialogDescription className="text-md text-muted-foreground mb-1 flex items-center">
										<User className="text-sm text-muted-foreground" />
										<Link href={`/user/${editSet.set.User.Name}`} className="hover:underline">
											{editSet.set.User.Name}
										</Link>
									</DialogDescription>
								)}
								<DialogDescription className="ml-1">
									Set ID:{" "}
									<Link
										href={`https://mediux.io/${editSet.set.Type}-set/${editSet.id}`}
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
								<div className="flex items-center justify-between">
									<div>
										{editSet.previousDateUpdated &&
											editSet.set.DateUpdated &&
											editSet.previousDateUpdated !== editSet.set.DateUpdated && (
												<div className="text-green-600 text-xs mt-1">Set has updates</div>
											)}
									</div>
									<div className="flex items-center">
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
							</div>
						))}
					</div>
				)}
				{updateError && <ErrorMessage error={updateError} />}
				<DialogFooter>
					<Button variant="outline" className="cursor-pointer" onClick={onClose}>
						Cancel
					</Button>
					<Button
						hidden={onlyIgnore}
						className="cursor-pointer"
						variant={allToDelete ? "destructive" : "default"}
						onClick={confirmEdit}
					>
						{allToDelete ? (editSets.length === 1 ? "Delete Set" : "Delete All") : "Save"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
};

export interface SavedSetDeleteModalProps {
	open: boolean;
	onClose: () => void;
	title: string;
	confirmDelete: () => void;
}

export const SavedSetDeleteModal: React.FC<SavedSetDeleteModalProps> = ({ open, onClose, title, confirmDelete }) => (
	<Dialog open={open} onOpenChange={onClose}>
		<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[500px] ">
			<DialogHeader>
				<DialogTitle>Confirm Delete</DialogTitle>
				<DialogDescription>
					Are you sure you want to delete all sets for "{title}"? This action cannot be undone.
				</DialogDescription>
			</DialogHeader>
			<DialogFooter>
				<Button className="cursor-pointer" variant="outline" onClick={onClose}>
					Cancel
				</Button>
				<Button variant="destructive" className="cursor-pointer" onClick={confirmDelete}>
					Delete
				</Button>
			</DialogFooter>
		</DialogContent>
	</Dialog>
);

export interface SavedSetsListProps {
	savedSet: DBMediaItemWithPosterSets;
	layout: "table" | "card";
	onUpdate: () => void;
	unignoreLoading: boolean;
	setUnignoreLoading: (v: boolean) => void;
	setUpdateError: (v: APIResponse<unknown> | null) => void;
	onSelectSet: (ps: DBMediaItemWithPosterSets["PosterSets"][number]) => void;
}

export const SavedSetsList: React.FC<SavedSetsListProps> = ({
	savedSet,
	layout,
	onUpdate,
	unignoreLoading,
	setUnignoreLoading,
	setUpdateError,
	onSelectSet,
}) => {
	const onlyIgnore = savedSet.PosterSets.length === 1 && savedSet.PosterSets[0].PosterSetID === "ignore";

	if (onlyIgnore) {
		return (
			<div className="w-full flex flex-col gap-2">
				<span className="text-sm text-muted-foreground">
					This item is currently <span className="font-medium">ignored</span>.
				</span>
				<div>
					<Button
						size="sm"
						variant="outline"
						className="cursor-pointer"
						disabled={unignoreLoading}
						onClick={() =>
							handleStopIgnoring(savedSet, onUpdate, unignoreLoading, setUnignoreLoading, setUpdateError)
						}
					>
						{unignoreLoading ? "Updating..." : "Stop Ignoring"}
					</Button>
				</div>
			</div>
		);
	}

	const sets = savedSet.PosterSets.filter((s) => s.PosterSetID !== "ignore");
	if (sets.length === 0) return null;

	const heading = sets.length > 1 ? "Sets:" : "Set:";

	if (layout === "card") {
		return (
			<div className="w-full">
				<span className="text-sm text-muted-foreground mb-1 block">{heading}</span>
				<ul className="flex flex-col gap-1">
					{sets.map((ps) => (
						<li
							key={ps.PosterSetID}
							className="flex flex-row items-center justify-between rounded-sm px-2 py-1 hover:bg-muted/50"
						>
							<Link
								href={`/sets/${ps.PosterSetID}`}
								className="text-primary hover:underline text-sm shrink-0"
								onClick={() => onSelectSet(ps)}
							>
								{ps.PosterSetID}
							</Link>
							<Link
								href={`/user/${ps.PosterSet.User.Name}`}
								className="text-primary hover:underline text-xs text-right truncate ml-4"
							>
								{ps.PosterSet.User.Name || ""}
							</Link>
						</li>
					))}
				</ul>
			</div>
		);
	}

	// layout === "table"
	return (
		<div className="w-full">
			<span className="text-sm text-muted-foreground mb-1 block">{heading}</span>
			<ul className="flex flex-col gap-0">
				{sets.map((ps) => (
					<li
						key={ps.PosterSetID}
						className="flex flex-row gap-4 items-center rounded-sm px-2 py-1 hover:bg-muted/50"
					>
						<Link
							href={`/sets/${ps.PosterSetID}`}
							className="text-primary hover:underline text-sm"
							onClick={() => onSelectSet(ps)}
						>
							{ps.PosterSetID}
						</Link>
						<span className="text-muted-foreground text-xs">—</span>
						<Link href={`/user/${ps.PosterSet.User.Name}`} className="text-primary hover:underline text-xs">
							{ps.PosterSet.User.Name || ""}
						</Link>
					</li>
				))}
			</ul>
		</div>
	);
};
