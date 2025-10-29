"use client";

import { AlertTriangle, Delete, Edit, Loader, MoreHorizontal, RefreshCcw, RefreshCwOff } from "lucide-react";

import { useState } from "react";

import Link from "next/link";

import { hasSelectedTypesOverlapOnAutoDownload } from "@/components/shared/saved-sets-cards";
import {
	SavedSetDeleteModal,
	SavedSetEditModal,
	SavedSetsList,
	onCloseSavedSetsEditDeleteModals,
	refreshPosterSet,
	renderTypeBadges,
	savedSetsConfirmDelete,
	savedSetsConfirmEdit,
} from "@/components/shared/saved-sets-shared";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { TableCell, TableRow } from "@/components/ui/table";
import { P } from "@/components/ui/typography";

import { useMediaStore } from "@/lib/stores/global-store-media-store";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

const SavedSetsTableRow: React.FC<{
	savedSet: DBMediaItemWithPosterSets;
	onUpdate: () => void;
	handleRecheckItem: (title: string, item: DBMediaItemWithPosterSets) => void;
}> = ({ savedSet, onUpdate, handleRecheckItem }) => {
	// Initialize edit state from the savedSet.PosterSets array.
	const [editSets, setEditSets] = useState(() =>
		savedSet.PosterSets.map((set) => ({
			id: set.PosterSetID,
			previousDateUpdated: set.PosterSet.DateUpdated,
			set: set.PosterSet || "Unknown",
			selectedTypes: set.SelectedTypes,
			autoDownload: set.AutoDownload,
			toDelete: false,
		}))
	);

	// State to track Modal visibility
	const [isEditModalOpen, setIsEditModalOpen] = useState(false);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

	// State to track any error messages during updates.
	const [updateError, setUpdateError] = useState<APIResponse<unknown> | null>(null);

	// State to prevent multiple simultaneous operations.
	const [isMounted, setIsMounted] = useState(false);

	// Access global stores
	const { setMediaItem } = useMediaStore();

	// State to track if we are currently refreshing poster sets
	const [isRefreshing, setIsRefreshing] = useState(false);

	// Check if all sets are marked for deletion
	const allToDelete = editSets.every((set) => set.toDelete);

	// Track unignore loading state
	const [unignoreLoading, setUnignoreLoading] = useState(false);
	const onlyIgnore = savedSet.PosterSets.length === 1 && savedSet.PosterSets[0].PosterSetID === "ignore";

	const latestTimestamp = Math.max(...savedSet.PosterSets.map((ps) => new Date(ps.LastDownloaded).getTime()));
	const latestDate =
		latestTimestamp > 0
			? new Date(latestTimestamp).toLocaleDateString("en-US") +
				" " +
				new Date(latestTimestamp).toLocaleTimeString("en-US", {
					hour: "numeric",
					minute: "numeric",
					hour12: true,
				})
			: "N/A";

	return (
		<>
			<TableRow key={savedSet.TMDB_ID}>
				<TableCell>
					{savedSet.MediaItem.Type === "show" ? (
						<div>
							{savedSet.PosterSets.some((set) => set.AutoDownload) ? (
								<Popover>
									<PopoverTrigger asChild>
										<RefreshCcw className="text-green-500 cursor-help" size={24} />
									</PopoverTrigger>
									<PopoverContent className="p-2 max-w-xs">
										<p className="text-sm">
											Auto Download is enabled for this item. It will be periodically checked for
											new updates based on your Auto Download settings.
										</p>
									</PopoverContent>
								</Popover>
							) : (
								<Popover>
									<PopoverTrigger asChild>
										<RefreshCwOff className="text-red-500 cursor-help" size={24} />
									</PopoverTrigger>
									<PopoverContent className="p-2 max-w-xs">
										<p className="text-sm">
											Auto Download is disabled for this item. Click the edit button to enable it
											on one or more poster sets.
										</p>
									</PopoverContent>
								</Popover>
							)}
							{hasSelectedTypesOverlapOnAutoDownload(savedSet.PosterSets) && (
								<Popover>
									<PopoverTrigger asChild>
										<AlertTriangle className="text-yellow-500 cursor-help" size={24} />
									</PopoverTrigger>
									<PopoverContent className="p-2 max-w-xs">
										<p className="text-sm">
											Some poster sets have overlapping selected types with Auto Download enabled.
											This may cause unexpected behavior.
										</p>
									</PopoverContent>
								</Popover>
							)}
						</div>
					) : (
						<></>
					)}
				</TableCell>
				<TableCell className="font-medium">
					{
						<Link
							//href={formatMediaItemUrl(savedSet.MediaItem)}
							href={"/media-item/"}
							className="text-primary hover:underline"
							onClick={() => {
								setMediaItem(savedSet.MediaItem);
							}}
						>
							{savedSet.MediaItem.Title}
						</Link>
					}
				</TableCell>
				<TableCell>{savedSet.MediaItem.Year}</TableCell>
				<TableCell>{savedSet.MediaItem.LibraryTitle}</TableCell>
				<TableCell>{latestDate}</TableCell>
				<TableCell>
					<SavedSetsList
						savedSet={savedSet}
						layout="table"
						onUpdate={onUpdate}
						unignoreLoading={unignoreLoading}
						setUnignoreLoading={setUnignoreLoading}
						setUpdateError={setUpdateError}
					/>
				</TableCell>
				<TableCell>
					{savedSet.PosterSets.some(
						(set) =>
							Array.isArray(set.SelectedTypes) && set.SelectedTypes.some((type) => type.trim() !== "")
					) ? (
						<div className="flex flex-wrap gap-2">{renderTypeBadges(savedSet)}</div>
					) : (
						<P className="text-sm text-muted-foreground">No types selected.</P>
					)}
				</TableCell>

				{isRefreshing && (
					<div className="absolute inset-0 flex items-center justify-center bg-black/50 z-10">
						<Loader className="animate-spin h-8 w-8 text-primary" />
						<span className="ml-2 text-white">Refreshing sets...</span>
					</div>
				)}

				<TableCell className="text-right">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button
								variant="ghost"
								className="cursor-pointer p-1 hover:bg-muted/50 focus:bg-muted/50"
								size="icon"
							>
								<MoreHorizontal />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								className="cursor-pointer"
								onClick={async () => {
									await refreshPosterSet({
										editSets,
										setEditSets,
										savedSet,
										setIsRefreshing,
										setUpdateError,
									});
									setIsEditModalOpen(true);
								}}
							>
								<Edit className="ml-2" />
								{isRefreshing ? "Refreshing..." : "Edit"}
							</DropdownMenuItem>

							{savedSet.PosterSets.some(
								(set) => set.AutoDownload || savedSet.MediaItem.Type === "movie"
							) && (
								<DropdownMenuItem
									className="cursor-pointer"
									onClick={() => {
										handleRecheckItem(savedSet.MediaItem.Title, savedSet);
									}}
								>
									<RefreshCcw className="ml-2" />
									{savedSet.MediaItem.Type === "movie"
										? "Check Movie for Key Changes"
										: "Force Autodownload Recheck"}
								</DropdownMenuItem>
							)}
							<DropdownMenuItem
								onClick={() => setIsDeleteModalOpen(true)}
								className="text-destructive cursor-pointer"
							>
								<Delete className="ml-2" />
								Delete
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</TableCell>
			</TableRow>

			{/* Edit Modal */}
			<SavedSetEditModal
				open={isEditModalOpen}
				onClose={() =>
					onCloseSavedSetsEditDeleteModals({
						setIsEditModalOpen,
						setIsDeleteModalOpen,
						setUpdateError,
						setIsMounted,
					})
				}
				editSets={editSets}
				setEditSets={setEditSets}
				savedSet={savedSet}
				onlyIgnore={onlyIgnore}
				allToDelete={allToDelete}
				updateError={updateError}
				confirmEdit={() =>
					savedSetsConfirmEdit({
						editSets,
						savedSet,
						onUpdate,
						isMounted,
						setIsMounted,
						setUpdateError,
						setIsEditModalOpen,
						setIsDeleteModalOpen,
						allToDelete,
					})
				}
			/>

			{/* Delete Modal */}
			<SavedSetDeleteModal
				open={isDeleteModalOpen}
				onClose={() =>
					onCloseSavedSetsEditDeleteModals({
						setIsEditModalOpen,
						setIsDeleteModalOpen,
						setUpdateError,
						setIsMounted,
					})
				}
				title={savedSet.MediaItem.Title}
				confirmDelete={() =>
					savedSetsConfirmDelete({
						savedSet,
						onUpdate,
						isMounted,
						setIsMounted,
						setUpdateError,
						setIsDeleteModalOpen,
					})
				}
			/>
		</>
	);
};

export default SavedSetsTableRow;
