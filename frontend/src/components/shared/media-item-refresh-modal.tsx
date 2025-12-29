"use client";

import { patchRefreshMediaServerItem } from "@/services/mediaserver/api-mediaserver-refresh-item";

import { useEffect, useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Progress } from "@/components/ui/progress";

import { cn } from "@/lib/cn";

import { MediaItem, MediaItemEpisode, MediaItemSeason } from "@/types/media-and-posters/media-item-and-library";

export interface MediaItemDetailsRefreshMetadataProps {
	mediaItem: MediaItem;
	isOpen: boolean;
	onClose: () => void;
}

type RefreshTaskStatus = "pending" | "in-progress" | "completed" | "failed" | "skipped";
type RefreshTaskKind = "refresh-item" | "refresh-season" | "refresh-episode";

type RefreshTask = {
	id: string;
	kind: RefreshTaskKind;
	status: RefreshTaskStatus;
	label: string;
	ratingKey?: string;
	error?: string;
};

type RefreshProgress = {
	currentText: string;
	totalPlanned: number;
	tasks: RefreshTask[];
};

const newId = () => globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(16).slice(2)}`;

const getRefreshCounts = (state: RefreshProgress) => {
	const relevant = state.tasks;
	const done = relevant.filter(
		(t) => t.status === "completed" || t.status === "failed" || t.status === "skipped"
	).length;

	const total = Math.max(1, state.totalPlanned || relevant.length || 1);
	return { done, total };
};

const getRefreshPercent = (state: RefreshProgress) => {
	const { done, total } = getRefreshCounts(state);
	return Math.min(100, Math.round((done / total) * 100));
};

export const RefreshMetadataModal = ({ mediaItem, isOpen, onClose }: MediaItemDetailsRefreshMetadataProps) => {
	const isShow = mediaItem.Type === "show";

	const [refreshProgress, setRefreshProgress] = useState<RefreshProgress>({
		currentText: "",
		totalPlanned: 0,
		tasks: [],
	});

	const seasons = useMemo(() => {
		if (!isShow) return [];
		const count = mediaItem.Series?.SeasonCount ?? 0;
		// Seasons are typically 1..N (skip Specials/0 here; adjust if you want season 0 too)
		return Array.from({ length: Math.max(0, count) }, (_, i) => i + 1);
	}, [isShow, mediaItem.Series?.SeasonCount]);

	// Map: seasonNumber -> episodeCount (from mediaItem.Series.Seasons)
	const seasonEpisodeCounts = useMemo(() => {
		const map = new Map<number, number>();
		if (!isShow) return map;

		const seasonObjs = mediaItem.Series?.Seasons ?? [];
		for (const s of seasonObjs) {
			if (!s?.SeasonNumber || s.SeasonNumber === 0) continue;
			const count = Array.isArray(s.Episodes) ? s.Episodes.length : 0;
			map.set(s.SeasonNumber, count);
		}

		return map;
	}, [isShow, mediaItem.Series?.Seasons]);

	const [selectedSeasons, setSelectedSeasons] = useState<Record<number, boolean>>({});

	useEffect(() => {
		if (!isOpen) return;

		const initial: Record<number, boolean> = {};
		for (const s of seasons) initial[s] = true;
		setSelectedSeasons(initial);
	}, [isOpen, seasons]);

	const selectedSeasonNumbers = useMemo(() => {
		if (!isShow) return [];
		return seasons.filter((s) => selectedSeasons[s]);
	}, [isShow, seasons, selectedSeasons]);

	const selectedEpisodeCount = useMemo(() => {
		if (!isShow) return 0;
		return selectedSeasonNumbers.reduce((acc, s) => acc + (seasonEpisodeCounts.get(s) ?? 0), 0);
	}, [isShow, selectedSeasonNumbers, seasonEpisodeCounts]);

	const resetRefreshProgress = () => {
		setRefreshProgress({ currentText: "", totalPlanned: 0, tasks: [] });
	};

	useEffect(() => {
		// When closing, clear progress so the next open starts clean
		if (!isOpen) resetRefreshProgress();
	}, [isOpen]);

	const updateRefreshTask = (taskId: string, updater: (t: RefreshTask) => RefreshTask) => {
		setRefreshProgress((prev) => ({
			...prev,
			tasks: prev.tasks.map((t) => (t.id === taskId ? updater(t) : t)),
		}));
	};

	const setRefreshText = (text: string) => {
		setRefreshProgress((prev) => ({ ...prev, currentText: text }));
	};

	const buildRefreshTasks = (): RefreshTask[] => {
		// Movie: single task
		if (!isShow) {
			return [
				{
					id: newId(),
					kind: "refresh-item",
					status: "pending",
					label: `Refresh "${mediaItem.Title}"`,
					ratingKey: mediaItem.RatingKey,
				},
			];
		}

		// Show: refresh item + each season + each episode in selected seasons
		const tasks: RefreshTask[] = [];

		// refresh show itself
		tasks.push({
			id: newId(),
			kind: "refresh-item",
			status: "pending",
			label: `Refresh "${mediaItem.Title}"`,
			ratingKey: mediaItem.RatingKey,
		});

		const seasonObjs = mediaItem.Series?.Seasons ?? [];
		for (const seasonNum of selectedSeasonNumbers) {
			const season = seasonObjs.find((s) => s?.SeasonNumber === seasonNum);

			// Season refresh
			tasks.push({
				id: newId(),
				kind: "refresh-season",
				status: "pending",
				label: `Refresh ${mediaItem.Title} - Season ${seasonNum}`,
				ratingKey: (season as MediaItemSeason)?.RatingKey,
			});

			// Episode refreshes
			const episodes = (season as MediaItemSeason)?.Episodes ?? [];
			for (const ep of episodes) {
				const epNum = (ep as MediaItemEpisode)?.EpisodeNumber;
				const epTitle = `"${ep.Title}" (S${seasonNum}E${epNum?.toString().padStart(2, "0")})`;
				tasks.push({
					id: newId(),
					kind: "refresh-episode",
					status: "pending",
					label: `Refresh ${epTitle}`,
					ratingKey: (ep as MediaItemEpisode)?.RatingKey,
				});
			}
		}

		return tasks;
	};

	const refreshByRatingKey = async (ratingKey: string, refreshTitle: string) => {
		try {
			const response = await patchRefreshMediaServerItem(
				{
					...mediaItem,
					RatingKey: ratingKey,
				},
				refreshTitle
			);
			if (response.status === "error") {
				throw new Error(response.error?.message || "Unknown error refreshing media server item");
			}
		} catch (error) {
			const message = error instanceof Error ? error.message : "Unknown error";
			throw new Error(message);
		}
	};

	const runRefreshTask = async (t: RefreshTask) => {
		// Missing ratingKey => skip (counts as done so progress completes)
		if (!t.ratingKey) {
			updateRefreshTask(t.id, (x) => ({
				...x,
				status: "skipped",
				error: "Missing ratingKey for refresh",
			}));
			return;
		}

		updateRefreshTask(t.id, (x) => ({ ...x, status: "in-progress", error: undefined }));
		setRefreshText(t.label);

		try {
			await refreshByRatingKey(t.ratingKey, t.label);
			updateRefreshTask(t.id, (x) => ({ ...x, status: "completed" }));
		} catch (e) {
			updateRefreshTask(t.id, (x) => ({
				...x,
				status: "failed",
				error: e instanceof Error ? e.message : "Unknown error",
			}));
		}
	};

	const handleRefresh = async () => {
		const tasks = buildRefreshTasks();
		setRefreshProgress({
			currentText: "Starting...",
			totalPlanned: tasks.length,
			tasks,
		});

		for (const t of tasks) {
			await runRefreshTask(t);
		}

		setRefreshText("Completed!");
	};

	const percent = getRefreshPercent(refreshProgress);
	const { done, total } = getRefreshCounts(refreshProgress);
	const hasProgress = refreshProgress.totalPlanned > 0;
	const hasErrors = refreshProgress.tasks.some((t) => t.status === "failed");
	const isAllSeasonsSelected = isShow && seasons.length > 0 && selectedSeasonNumbers.length === seasons.length;

	return (
		<Dialog
			open={isOpen}
			onOpenChange={(open) => {
				if (!open) onClose();
			}}
		>
			<DialogContent className={cn("max-h-[80vh] overflow-y-auto sm:max-w-[700px]", "border border-primary")}>
				<DialogHeader>
					<DialogTitle className="text-lg font-bold">Refresh Metadata</DialogTitle>

					{!isShow ? (
						<DialogDescription className="text-sm text-muted-foreground">
							Refresh metadata for <span className="font-semibold">'{mediaItem.Title}'</span>
						</DialogDescription>
					) : (
						<DialogDescription className="text-sm text-muted-foreground">
							Select which seasons to refresh for{" "}
							<span className="font-semibold">'{mediaItem.Title}'</span>
						</DialogDescription>
					)}
				</DialogHeader>

				{isShow && (
					<div className="mt-2 space-y-3">
						<div className="flex items-center justify-between gap-2">
							<div className="text-sm font-medium">Seasons</div>

							<div className="flex items-center gap-2">
								<Badge
									className={cn(
										"cursor-pointer transition duration-200 hover:brightness-110 active:scale-95",
										isAllSeasonsSelected
											? "bg-primary text-primary-foreground"
											: "bg-secondary text-secondary-foreground"
									)}
									onClick={() => {
										const allSelected: Record<number, boolean> = {};
										for (const s of seasons) allSelected[s] = true;
										setSelectedSeasons(allSelected);
									}}
								>
									Select All
								</Badge>

								<Badge
									className="cursor-pointer bg-secondary text-secondary-foreground hover:brightness-110 transition duration-200 active:scale-95"
									onClick={() => {
										setSelectedSeasons({});
									}}
								>
									Deselect All
								</Badge>
							</div>
						</div>

						<div className="flex flex-wrap gap-2 mt-2">
							{seasons.map((s) => {
								const epCount = seasonEpisodeCounts.get(s) ?? 0;
								const isSelected = !!selectedSeasons[s];

								return (
									<Badge
										key={s}
										className={`flex items-center transition duration-200 hover:brightness-120 active:scale-95 ${
											isSelected ? "cursor-pointer bg-primary" : "cursor-pointer bg-secondary"
										}`}
										onClick={() => {
											setSelectedSeasons((prev) => ({ ...prev, [s]: !isSelected }));
										}}
									>
										<span className="min-w-0">
											<span className={cn(isSelected ? "" : "text-muted-foreground")}>
												Season {s}
											</span>
											<span
												className={cn(
													"ml-2 text-xs tabular-nums",
													isSelected ? "" : "text-muted-foreground"
												)}
											>
												({epCount} ep{epCount === 1 ? "" : "s"})
											</span>
										</span>
									</Badge>
								);
							})}
						</div>
					</div>
				)}

				{isShow && (
					<div className="mt-1 text-xs text-muted-foreground">
						Selected: {selectedSeasonNumbers.length}/{seasons.length} season
						{seasons.length === 1 ? "" : "s"}
						{" • "}
						{selectedEpisodeCount} episode{selectedEpisodeCount === 1 ? "" : "s"}
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
										title={refreshProgress.currentText}
									>
										<span className="w-full min-w-0 truncate text-center">
											{refreshProgress.currentText}
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

				<div className="mt-4 flex justify-end space-x-2">
					<Button variant="outline" onClick={onClose}>
						Cancel
					</Button>

					<Button onClick={handleRefresh}>
						{isShow && selectedSeasonNumbers.length === 0 && "Refresh Show"}
						{!isShow && "Refresh"}
						{isShow &&
							selectedSeasonNumbers.length > 0 &&
							`Refresh Show & ${selectedSeasonNumbers.length} Season${selectedSeasonNumbers.length === 1 ? "" : "s"}`}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
};
