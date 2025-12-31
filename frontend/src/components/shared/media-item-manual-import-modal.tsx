"use client";

import { patchDownloadPosterFileAndUpdateMediaServer } from "@/services/mediaserver/api-mediaserver-download-and-update";
import yaml from "js-yaml";
import { User } from "lucide-react";

import { useEffect, useState } from "react";

import Link from "next/link";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import { Progress } from "@/components/ui/progress";
import { Textarea } from "@/components/ui/textarea";

import { cn } from "@/lib/cn";

import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterFile, PosterSet } from "@/types/media-and-posters/poster-sets";

export type MediaItemManualImportModalProps = {
	mediaItem: MediaItem;
	isOpen: boolean;
	onClose: () => void;
};

type ImportTaskStatus = "pending" | "in-progress" | "completed" | "failed";
type ImportTask = {
	id: string;
	status: ImportTaskStatus;
	label: string;
	PosterFile: PosterFile;
	error?: string;
};

type ImportProgress = {
	currentText: string;
	totalPlanned: number;
	tasks: ImportTask[];
};

const newId = () => globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(16).slice(2)}`;

const getRefreshCounts = (state: ImportProgress) => {
	const relevant = state.tasks;
	const done = relevant.filter((t) => t.status === "completed" || t.status === "failed").length;

	const total = Math.max(1, state.totalPlanned || relevant.length || 1);
	return { done, total };
};

const getImportPercentComplete = (progress: ImportProgress) => {
	const { done, total } = getRefreshCounts(progress);
	return Math.min(100, Math.round((done / total) * 100));
};

export function ManualImportModal({ mediaItem, isOpen, onClose }: MediaItemManualImportModalProps) {
	// States for YAML text input
	const [yamlText, setYamlText] = useState("");
	const [isValidYaml, setIsValidYaml] = useState(false);

	// States for PreImport
	const [setAuthor, setSetAuthor] = useState<string | null>(null);
	const [setId, setSetId] = useState<string | null>(null);
	const [validPosterSet, setValidPosterSet] = useState<PosterSet | null>(null);
	const [numberOfPostersInSet, setNumberOfPostersInSet] = useState<number | null>(null);
	const [numberOfBackdropsInSet, setNumberOfBackdropsInSet] = useState<number | null>(null);
	const [numberOfSeasonPostersInSet, setNumberOfSeasonPostersInSet] = useState<number | null>(null);
	const [numberOfTitlecardsInSet, setNumberOfTitlecardsInSet] = useState<number | null>(null);

	// States for Import Progress
	const [importProgress, setImportProgress] = useState<ImportProgress>({
		currentText: "",
		totalPlanned: 0,
		tasks: [],
	});

	// States for Error Handling
	const [error, setError] = useState<string | null>(null);

	const resetImportProgress = () => {
		setImportProgress({ currentText: "", totalPlanned: 0, tasks: [] });
	};

	useEffect(() => {
		// When closing, clear progress so the next open starts clean
		if (!isOpen) {
			resetImportProgress();
		}
	}, [isOpen]);

	const setImportText = (text: string) => {
		setImportProgress((prev) => ({ ...prev, currentText: text }));
	};

	const updateImportTask = (taskId: string, updater: (t: ImportTask) => ImportTask) => {
		setImportProgress((prev) => ({
			...prev,
			tasks: prev.tasks.map((t) => (t.id === taskId ? updater(t) : t)),
		}));
	};

	const handleValidateYaml = () => {
		try {
			if (!yamlText || yamlText.trim() === "") {
				throw new Error("YAML input is empty.");
			}

			// Parse YAML
			const parsedData: MediuxYaml = yaml.load(yamlText) as MediuxYaml;

			// Validate basic structure
			if (typeof parsedData !== "object" || parsedData === null) {
				throw new Error("YAML does not represent a valid object.");
			}

			// Extract author and setId from comments
			const authorMatch = yamlText.match(/Set by ([\w\d_-]+) on MediUX/i);
			const setIdMatch = yamlText.match(/mediux\.pro\/sets\/(\d+)/i);
			const author = authorMatch ? authorMatch[1] : null;
			const setId = setIdMatch ? setIdMatch[1] : null;

			setSetAuthor(author);
			setSetId(setId);

			// Build the poster set using your utility
			const posterSet = yamlToPosterSet(parsedData, mediaItem, mediaItem.Type as "movie" | "show", setId, author);

			// Count posters, backdrops, etc. from the posterSet object
			let posterCount = posterSet.Poster ? 1 : 0;
			let backdropCount = posterSet.Backdrop ? 1 : 0;
			let seasonPosterCount = posterSet.SeasonPosters ? posterSet.SeasonPosters.length : 0;
			let titlecardCount = posterSet.TitleCards ? posterSet.TitleCards.length : 0;

			setValidPosterSet(posterSet);
			setNumberOfPostersInSet(posterCount);
			setNumberOfBackdropsInSet(backdropCount);
			setNumberOfSeasonPostersInSet(seasonPosterCount);
			setNumberOfTitlecardsInSet(titlecardCount);

			setIsValidYaml(true);
		} catch (error) {
			setIsValidYaml(false);
			setError("Error: " + (error instanceof Error ? error.message : String(error)));
		}
	};

	const buildImportTasks = (posterSet: PosterSet): ImportTask[] => {
		const tasks: ImportTask[] = [];
		if (posterSet.Poster) {
			tasks.push({
				id: newId(),
				status: "pending",
				label: `Downloading Poster for '${mediaItem.Title}'`,
				PosterFile: posterSet.Poster,
			});
		}
		if (posterSet.Backdrop) {
			tasks.push({
				id: newId(),
				status: "pending",
				label: `Downloading Backdrop for '${mediaItem.Title}'`,
				PosterFile: posterSet.Backdrop,
			});
		}
		posterSet.SeasonPosters?.forEach((sp, index) => {
			tasks.push({
				id: newId(),
				status: "pending",
				label: `Downloading Season ${sp.Season?.Number} Poster for '${mediaItem.Title}'`,
				PosterFile: posterSet.SeasonPosters![index],
			});
		});
		posterSet.TitleCards?.forEach((ep, index) => {
			const episodeTitle =
				mediaItem.Series?.Seasons.find((s) =>
					s.Episodes.find(
						(e) =>
							e.SeasonNumber === ep.Episode?.SeasonNumber && e.EpisodeNumber === ep.Episode?.EpisodeNumber
					)
				)?.Episodes.find(
					(e) => e.SeasonNumber === ep.Episode?.SeasonNumber && e.EpisodeNumber === ep.Episode?.EpisodeNumber
				)?.Title || "";

			tasks.push({
				id: newId(),
				status: "pending",
				label: `Downloading "${episodeTitle}" (S${ep.Episode?.SeasonNumber}E${ep.Episode?.EpisodeNumber?.toString().padStart(2, "0")}) Titlecard`,
				PosterFile: posterSet.TitleCards![index],
			});
		});
		return tasks;
	};

	const runImportTask = async (t: ImportTask) => {
		updateImportTask(t.id, (task) => ({ ...task, status: "in-progress" }));
		try {
			setImportText(t.label);
			const response = await patchDownloadPosterFileAndUpdateMediaServer(t.PosterFile, mediaItem, t.label);
			if (response.status === "error") {
				throw new Error(response.error?.message || "Unknown error");
			}

			updateImportTask(t.id, (task) => ({ ...task, status: "completed" }));
		} catch (error) {
			updateImportTask(t.id, (task) => ({
				...task,
				status: "failed",
				error: error instanceof Error ? error.message : String(error),
			}));
		}
	};

	const handleImport = async () => {
		const tasks = buildImportTasks(validPosterSet!);
		setImportProgress({ currentText: "Starting import...", totalPlanned: tasks.length, tasks });
		for (const task of tasks) {
			await runImportTask(task);
		}
		setImportText("Completed!");
	};

	const percent = getImportPercentComplete(importProgress);
	const { done, total } = getRefreshCounts(importProgress);
	const hasProgress = importProgress.totalPlanned > 0;
	const hasErrors = importProgress.tasks.some((t) => t.status === "failed");

	return (
		<Dialog open={isOpen} onOpenChange={onClose}>
			<DialogContent className={cn("max-h-[80vh] overflow-y-auto sm:max-w-[700px]", "border border-primary")}>
				<DialogHeader>
					<DialogTitle>Manual YAML Import for '{mediaItem.Title}'</DialogTitle>
					{setAuthor && (
						<div className="flex items-center justify-center sm:justify-start">
							<Avatar className="rounded-lg mr-1 w-4 h-4">
								<AvatarImage
									src={`/api/mediux/avatar-image?username=${setAuthor}`}
									className="w-4 h-4"
								/>
								<AvatarFallback className="">
									<User className="w-4 h-4" />
								</AvatarFallback>
							</Avatar>
							<Link href={`/user/${setAuthor}`} className="hover:underline">
								{setAuthor}
							</Link>
						</div>
					)}
					{setId && <DialogDescription>Set ID: {setId}</DialogDescription>}
				</DialogHeader>
				{!isValidYaml ? (
					<div>
						<label className="block mb-2 font-medium">MediUX YAML:</label>
						<Textarea
							value={yamlText}
							onChange={(e) => setYamlText(e.target.value)}
							rows={5}
							placeholder="Paste MediUX YAML here..."
						/>
						{error && <div className="text-destructive mt-2">{error}</div>}
					</div>
				) : (
					<div className="flex flex-wrap gap-2 mt-2">
						{numberOfPostersInSet !== null && (
							<Badge>
								Poster{numberOfPostersInSet !== 1 ? "s" : ""}: {numberOfPostersInSet}
							</Badge>
						)}
						{numberOfBackdropsInSet !== null && (
							<Badge>
								Backdrop{numberOfBackdropsInSet !== 1 ? "s" : ""}: {numberOfBackdropsInSet}
							</Badge>
						)}
						{numberOfSeasonPostersInSet !== null && (
							<Badge>
								Season Poster{numberOfSeasonPostersInSet !== 1 ? "s" : ""}: {numberOfSeasonPostersInSet}
							</Badge>
						)}
						{numberOfTitlecardsInSet !== null && (
							<Badge>
								Titlecard{numberOfTitlecardsInSet !== 1 ? "s" : ""}: {numberOfTitlecardsInSet}
							</Badge>
						)}
					</div>
				)}

				{/* Progress Bar */}
				{hasProgress ? (
					<div className="mt-3 min-h-[24px]">
						<div className="flex items-center justify-between gap-2">
							<div className="relative w-full min-w-0">
								<Progress
									value={percent}
									className={cn(
										"w-full rounded-md overflow-hidden",
										percent < 100 ? "h-5" : "h-3",
										percent === 100 && !hasErrors && "[&>div]:bg-green-500",
										percent === 100 && hasErrors && "[&>div]:bg-destructive",
										percent < 100 && "[&>div]:bg-primary"
									)}
								/>
								{percent < 100 && (
									<span
										className={cn(
											"absolute inset-0 flex items-center justify-center",
											"text-xs text-white pointer-events-none",
											"px-2 min-w-0"
										)}
										title={importProgress.currentText}
									>
										<span className="w-full min-w-0 truncate text-center">
											{importProgress.currentText}
										</span>
									</span>
								)}
							</div>

							<span className="text-sm text-muted-foreground min-w-[56px] text-right tabular-nums">
								{percent}%
							</span>
						</div>

						{done !== total && (
							<div className="mt-1 text-xs text-muted-foreground tabular-nums text-right">
								{done}/{total}
							</div>
						)}
					</div>
				) : null}

				<DialogFooter>
					<Button variant="outline" onClick={onClose}>
						Cancel
					</Button>
					{!isValidYaml && <Button onClick={handleValidateYaml}>Validate YAML</Button>}
					{isValidYaml && (
						<>
							<Button
								variant="destructive"
								onClick={() => {
									setIsValidYaml(false);
									setYamlText("");
									setSetAuthor(null);
									setSetId(null);
									setNumberOfPostersInSet(null);
									setNumberOfBackdropsInSet(null);
									setNumberOfSeasonPostersInSet(null);
									setNumberOfTitlecardsInSet(null);
									setError(null);
								}}
							>
								Clear
							</Button>
							<Button onClick={handleImport}>Import</Button>
						</>
					)}
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}

export function yamlToPosterSet(
	yamlData: MediuxYaml,
	mediaItem: MediaItem,
	type: "movie" | "show",
	setId: string | null,
	setAuthor: string | null
): PosterSet {
	const posterSet: PosterSet = {
		ID: setId || "",
		Title: mediaItem.Title,
		Type: type,
		User: { Name: setAuthor ?? "" },
		DateCreated: new Date().toISOString(),
		DateUpdated: new Date().toISOString(),
		Status: "",
	};

	function extractAssetId(url?: string): string {
		if (!url) return "";
		const parts = url.split("/");
		return parts[parts.length - 1];
	}

	// Helper to create PosterFile
	const makePosterFile = (
		src: string,
		fileType: string,
		seasonNumber?: number,
		episodeNumber?: number
	): PosterFile => {
		const assetId = extractAssetId(src);
		return {
			ID: "---" + assetId,
			Type: fileType,
			Modified: new Date().toISOString(),
			FileSize: 0,
			Src: "---" + assetId,
			Blurhash: "",
			Movie:
				type === "movie"
					? {
							ID: mediaItem.TMDB_ID,
							Title: mediaItem.Title,
							Status: "",
							Tagline: "",
							Slug: "",
							DateUpdated: "",
							TVbdID: "",
							ImdbID: "",
							TraktID: "",
							ReleaseDate: "",
							MediaItem: mediaItem,
						}
					: undefined,
			Show:
				type === "show"
					? {
							ID: mediaItem.TMDB_ID,
							Title: mediaItem.Title,
							MediaItem: mediaItem,
						}
					: undefined,
			Season:
				type === "show"
					? {
							Number: seasonNumber!,
						}
					: undefined,
			Episode:
				type === "show"
					? {
							Title:
								mediaItem.Series?.Seasons.find((s) =>
									s.Episodes.find(
										(e) => e.SeasonNumber === seasonNumber && e.EpisodeNumber === episodeNumber
									)
								)?.Episodes.find(
									(e) => e.SeasonNumber === seasonNumber && e.EpisodeNumber === episodeNumber
								)?.Title || "",
							SeasonNumber: seasonNumber!,
							EpisodeNumber: episodeNumber!,
						}
					: undefined,
		};
	};

	if (type === "movie") {
		// Only keep the item that matches the mediaItem (e.g., by TMDB ID)
		const matchKey = mediaItem.TMDB_ID;
		const item = yamlData[matchKey];
		if (!item) throw new Error("No matching item found in YAML for this media item.");
		if (item.url_poster) posterSet.Poster = makePosterFile(item.url_poster, "poster");
		if (item.url_background) posterSet.Backdrop = makePosterFile(item.url_background, "backdrop");
	} else if (type === "show") {
		// For shows, just use the first item in the YAML
		const firstKey = Object.keys(yamlData)[0];
		const item = yamlData[firstKey];

		if (item.url_poster) posterSet.Poster = makePosterFile(item.url_poster, "poster");
		if (item.url_background) posterSet.Backdrop = makePosterFile(item.url_background, "backdrop");
		if (item.seasons) {
			posterSet.SeasonPosters = [];
			for (const [seasonNum, seasonData] of Object.entries(item.seasons)) {
				const seasonPosters: PosterFile[] = [];
				if (
					seasonData.url_poster &&
					mediaItem.Series?.Seasons.find((s) => s.SeasonNumber.toString() === seasonNum)
				) {
					seasonPosters.push(makePosterFile(seasonData.url_poster, "seasonPoster", parseInt(seasonNum)));
				}
				const titleCards: PosterFile[] = [];
				for (const [episodeNum, episodeData] of Object.entries(seasonData.episodes || {})) {
					if (
						episodeData.url_poster &&
						mediaItem.Series?.Seasons.find((s) =>
							s.Episodes.find(
								(e) =>
									e.SeasonNumber.toString() === seasonNum && e.EpisodeNumber.toString() === episodeNum
							)
						)
					) {
						titleCards.push(
							makePosterFile(
								episodeData.url_poster,
								"titlecard",
								parseInt(seasonNum),
								parseInt(episodeNum)
							)
						);
					}
				}
				posterSet.SeasonPosters.push(...seasonPosters);
				posterSet.TitleCards = posterSet.TitleCards || [];
				posterSet.TitleCards.push(...titleCards);
			}
		}
	}

	return posterSet;
}

export type MediuxYaml = {
	[key: string]: MediuxYamlEntry;
};

export type MediuxYamlEntry = {
	url_poster?: string;
	url_background?: string;
	seasons?: {
		[seasonNumber: string]: MediuxYamlSeasonEntry;
	};
};

export type MediuxYamlSeasonEntry = {
	url_poster?: string;
	episodes?: {
		[episodeNumber: string]: MediuxYamlEpisodeEntry;
	};
};

export type MediuxYamlEpisodeEntry = {
	url_poster?: string;
};
