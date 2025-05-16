"use client";
import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardContent } from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogFooter,
	DialogTitle,
} from "@/components/ui/dialog";
import { H4, P, Small } from "@/components/ui/typography";
import { deleteItemFromDB, patchSavedSetInDB } from "@/services/api.db";
import { SavedSet } from "@/types/databaseSavedSet";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "./dropdown-menu";
import { Badge } from "./badge";
import { Separator } from "./separator";
import { CheckCircle2 as Checkmark } from "lucide-react";
import { X } from "lucide-react";
import Image from "next/image";
import { MoreHorizontal } from "lucide-react";
import { DialogDescription } from "@radix-ui/react-dialog";
import Link from "next/link";
import { cn } from "@/lib/utils";

const formatDate = (dateString: string) => {
	try {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat("en-US", {
			year: "numeric",
			month: "long",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		}).format(date);
	} catch {
		return "Invalid Date";
	}
};

// Helper to get the latest date from all sets.
const getLatestUpdate = (sets: SavedSet["Sets"]) => {
	// Map each set's LastUpdate into a Date instance and filter out invalid dates.
	const dates = sets
		.map((set) => new Date(set.LastUpdate ?? ""))
		.filter((date) => !isNaN(date.getTime()));
	if (dates.length === 0) return "";
	const latest = new Date(Math.max(...dates.map((d) => d.getTime())));
	return latest.toISOString();
};

const SavedSetsCard: React.FC<{
	savedSet: SavedSet;
	onUpdate: () => void;
}> = ({ savedSet, onUpdate }) => {
	// Initialize edit state from the savedSet.Sets array.
	const [editSets, setEditSets] = useState(() =>
		savedSet.Sets.map((set) => ({
			id: set.ID,
			set: set.Set,
			selectedTypes: set.SelectedTypes,
			autoDownload: set.AutoDownload,
			toDelete: false,
		}))
	);

	const [isEditModalOpen, setIsEditModalOpen] = useState(false);

	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

	const [updateError, setUpdateError] = useState("");
	const [isMounted, setIsMounted] = useState(false);

	const allToDelete = editSets.every((set) => set.toDelete);

	const confirmEdit = async () => {
		if (isMounted) return;
		setIsMounted(true);

		if (allToDelete) {
			setIsEditModalOpen(false);
			setIsDeleteModalOpen(true);
			setUpdateError("");
			setIsMounted(false);
			return;
		}

		const updatedSavedSet = {
			...savedSet,
			Sets: editSets.map((editSet) => ({
				ID: editSet.id,
				Set: editSet.set,
				SelectedTypes: editSet.selectedTypes,
				AutoDownload: editSet.autoDownload,
				ToDelete: editSet.toDelete,
			})),
		};

		const response = await patchSavedSetInDB(updatedSavedSet);
		if (response.status !== "success") {
			setUpdateError(response.message);
		} else {
			setUpdateError("");
			setIsEditModalOpen(false);
			onUpdate();
		}

		setIsMounted(false);
	};

	const confirmDelete = async () => {
		if (isMounted) return;
		setIsMounted(true);
		const resp = await deleteItemFromDB(savedSet.ID);
		if (resp.status !== "success") {
			setUpdateError(resp.message);
		} else {
			setIsDeleteModalOpen(false);
			setUpdateError("");
			onUpdate();
		}
		setIsMounted(false);
	};

	const renderSetBadges = () => {
		return savedSet.Sets.map((set) => (
			<Link
				key={set.ID}
				href={`https://mediux.pro/sets/${set.ID}`}
				target="_blank"
				rel="noopener noreferrer"
				className="transition transform hover:scale-105 hover:underline"
			>
				<Badge className="cursor-pointer text-sm">{set.ID}</Badge>
			</Link>
		));
	};

	const renderTypeBadges = () => {
		// Flatten all SelectedTypes arrays from every set
		const allTypes = savedSet.Sets.flatMap((set) => set.SelectedTypes);
		// Create a deduplicated array of unique types
		const uniqueTypes = Array.from(new Set(allTypes));
		return uniqueTypes.map((type) => (
			<Badge key={type}>
				{type === "poster"
					? "Poster"
					: type === "backdrop"
					? "Backdrop"
					: type === "seasonPoster"
					? "Season Posters"
					: type === "titlecard"
					? "Title Card"
					: type}
			</Badge>
		));
	};

	return (
		<Card className="relative w-full max-w-md mx-auto mb-4">
			<CardHeader>
				{/* Top Left: Auto Download Icon */}
				<div className="absolute top-2 left-2">
					{savedSet.Sets[0].AutoDownload ? (
						<Checkmark className="text-green-500" size={24} />
					) : (
						<X className="text-red-500" size={24} />
					)}
				</div>

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
								onClick={() => setIsEditModalOpen(true)}
							>
								Edit
							</DropdownMenuItem>
							<DropdownMenuItem
								onClick={() => setIsDeleteModalOpen(true)}
								className="text-destructive"
							>
								Delete
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>

				{/* Middle: Image */}
				<div className="flex justify-center mt-6">
					<Image
						src={`/api/mediaserver/image/${savedSet.MediaItem.RatingKey}/poster`}
						alt={savedSet.MediaItem.Title}
						width={150}
						height={225}
						className="rounded-md"
						unoptimized
						loading="lazy"
					/>
				</div>
			</CardHeader>

			{/* Content */}
			<CardContent>
				{/* Title */}
				<H4 className="text-lg font-semibold">
					{savedSet.MediaItem.Title}
				</H4>

				{/* Year */}
				<P
					className={cn(
						"text-sm text-muted-foreground",
						savedSet.MediaItem.Year
					)}
				>
					Year: {savedSet.MediaItem.Year}
				</P>

				{/* Library Title */}
				<P
					className={cn(
						"text-sm text-muted-foreground",
						savedSet.MediaItem.Year
					)}
				>
					Library: {savedSet.MediaItem.LibraryTitle}
				</P>

				{/* Last Updated */}
				<P
					className={cn(
						"text-sm text-muted-foreground",
						savedSet.MediaItem.Year
					)}
				>
					Last Updated:{" "}
					{formatDate(getLatestUpdate(savedSet.Sets) || "")}
				</P>

				<div className="flex flex-wrap gap-2">
					Sets: {renderSetBadges()}
				</div>

				{/* Separator */}
				<Separator className="my-4" />

				{/* Badges */}
				<div className="flex flex-wrap gap-2">{renderTypeBadges()}</div>
			</CardContent>

			{/* Edit Modal */}
			<Dialog open={isEditModalOpen} onOpenChange={setIsEditModalOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Edit Saved Set</DialogTitle>
						<DialogDescription>
							Edit each set individually. Toggle type badges to
							update selected types. Use the delete option to mark
							a set for deletion.
						</DialogDescription>
					</DialogHeader>
					<div className="space-y-4">
						{/* Iterate over each set */}
						{editSets.map((editSet, index) => (
							<div
								key={editSet.id}
								className="border p-2 rounded-md"
							>
								<div className="flex items-center justify-between">
									<span className="font-semibold">
										Set ID: {editSet.id}
									</span>
									<Button
										variant={
											editSet.toDelete
												? "destructive"
												: "outline"
										}
										size="sm"
										onClick={() => {
											setEditSets((prev) =>
												prev.map((item, i) =>
													i === index
														? {
																...item,
																toDelete:
																	!item.toDelete,
																// When marking as delete, clear the selected types.
																selectedTypes:
																	!item.toDelete
																		? []
																		: item.selectedTypes,
														  }
														: item
												)
											);
										}}
									>
										{editSet.toDelete
											? "Undo Delete"
											: "Delete Set"}
									</Button>
								</div>
								<div className="flex flex-wrap gap-2 mt-2">
									{[
										"poster",
										"backdrop",
										"seasonPoster",
										"titlecard",
									].map((type) => {
										// Determine if the current set already has this type selected.
										const isSelected =
											editSet.selectedTypes.includes(
												type
											);
										// Disable the badge if the current set is marked for deletion or if another set (other than the current) already selected this type.
										const isTypeDisabled =
											editSet.toDelete ||
											(!isSelected &&
												editSets.some(
													(item, j) =>
														j !== index &&
														item.selectedTypes.includes(
															type
														)
												));
										return (
											<Badge
												key={type}
												className={`flex items-center gap-2 transition duration-200
                            ${
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
															if (i !== index)
																return item;
															const newSelectedTypes =
																item.selectedTypes.includes(
																	type
																)
																	? item.selectedTypes.filter(
																			(
																				t
																			) =>
																				t !==
																				type
																	  )
																	: [
																			...item.selectedTypes,
																			type,
																	  ];
															return {
																...item,
																selectedTypes:
																	newSelectedTypes,
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
													: type === "titlecard"
													? "Title Card"
													: type}
											</Badge>
										);
									})}
								</div>
								{/* Auto Download Badge */}
								<div className="flex flex-wrap gap-2 mt-2">
									<Badge
										className={`cursor-pointer transition duration-200
                    ${
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
																autoDownload:
																	!item.autoDownload,
														  }
														: item
												)
											);
										}}
									>
										{editSet.autoDownload
											? "Autodownload"
											: "No Autodownload"}
									</Badge>
								</div>
							</div>
						))}
					</div>

					{updateError && (
						<Small className="text-destructive mt-2">
							{updateError}
						</Small>
					)}
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => {
								setEditSets(
									savedSet.Sets.map((set) => ({
										id: set.ID,
										set: set.Set,
										selectedTypes: set.SelectedTypes,
										autoDownload: set.AutoDownload,
										toDelete: false,
									}))
								);
								setIsEditModalOpen(false);
							}}
						>
							Cancel
						</Button>
						<Button
							variant={allToDelete ? "destructive" : "default"}
							onClick={confirmEdit}
						>
							{allToDelete
								? editSets.length === 1
									? "Delete Set"
									: "Delete All"
								: "Save"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Delete Confirmation Modal */}
			<Dialog
				open={isDeleteModalOpen}
				onOpenChange={setIsDeleteModalOpen}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Confirm Delete</DialogTitle>
						<DialogDescription>
							Are you sure you want to delete all sets for "
							{savedSet.MediaItem.Title}"? This action cannot be
							undone.
						</DialogDescription>
					</DialogHeader>

					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setIsDeleteModalOpen(false)}
						>
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
