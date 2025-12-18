"use client";

import { posterSetToFormItem } from "@/helper/download-modal/poster-set-to-form-item";
import { formatDownloadSize } from "@/helper/format-download-size";
import { postAddItemToDB } from "@/services/database/api-db-item-add";
import { postAddToQueue } from "@/services/download-queue/api-queue-add";
import { patchDownloadPosterFileAndUpdateMediaServer } from "@/services/mediaserver/api-mediaserver-download-and-update";
import { fetchMediaServerItemContent } from "@/services/mediaserver/api-mediaserver-fetch-item-content";
import { zodResolver } from "@hookform/resolvers/zod";
import {
	Check,
	CircleAlert,
	Database,
	DatabaseZap,
	Download,
	ListEnd,
	Loader,
	OctagonMinus,
	RefreshCcw,
	TriangleAlert,
	User,
	X,
} from "lucide-react";
import { z } from "zod";

import { Fragment, useEffect, useMemo, useRef, useState } from "react";
import React from "react";
import { ControllerRenderProps, useForm, useWatch } from "react-hook-form";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import DownloadModalPopover from "@/components/shared/download-modal-popover";
import { PopoverHelp } from "@/components/shared/popover-help";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogOverlay,
	DialogPortal,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel } from "@/components/ui/form";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Progress } from "@/components/ui/progress";
import { Lead } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useUserPreferencesStore } from "@/lib/stores/global-user-preferences";

import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterFile, PosterSet } from "@/types/media-and-posters/poster-sets";

export interface FormItemDisplay {
	MediaItemRatingKey: string;
	MediaItemTitle: string;
	MediaItem: MediaItem;
	SetID: string;
	Set: PosterSet;
}

const AssetTypes = {
	poster: "Poster",
	backdrop: "Backdrop",
	seasonPoster: "Season Posters",
	specialSeasonPoster: "Special Season Poster",
	titlecard: "Titlecards",
} as const;

type AssetType = keyof typeof AssetTypes;
type SourceType = "show" | "movie" | "collection";

interface AssetTypeFormValues {
	types: AssetType[];
	autodownload?: boolean;
	addToDBOnly?: boolean;
	source?: SourceType;
}

type TaskStatus = "pending" | "in-progress" | "completed" | "failed" | "skipped";

type DownloadTaskPayload = {
	kind: "download";
	itemKey: string;
	itemTitle: string;
	posterFile: PosterFile;
	posterFileType: AssetType;
	fileName: string;
	mediaItem: MediaItem;
};

type AddToDBTaskPayload = {
	kind: "addToDB";
	itemKey: string;
	itemTitle: string;
	dbItem: DBMediaItemWithPosterSets;
};

type AddToQueueTaskPayload = {
	kind: "addToQueue";
	itemKey: string;
	itemTitle: string;
	dbItem: DBMediaItemWithPosterSets;
};

// Non-Retryable “record only” task (e.g. fetch latest media item failed)
type NoteTaskPayload = {
	kind: "note";
	itemKey: string;
	itemTitle: string;
};

type TaskPayload = DownloadTaskPayload | AddToDBTaskPayload | AddToQueueTaskPayload | NoteTaskPayload;

type Task = {
	id: string;
	status: TaskStatus;
	label: string;
	attempts: number;
	payload: TaskPayload;
	error?: string;
};

type ItemProgress = {
	itemKey: string;
	title: string;
	tasks: Task[];
};

type DownloadProgress = {
	currentText: string;
	totalPlanned: number;
	items: Record<string, ItemProgress>;
};

// helper for stable ids
const newId = () => globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(16).slice(2)}`;

// derive progress counts from tasks (uses totalPlanned so total doesn't grow as tasks are added)
const getOverallCounts = (state: DownloadProgress) => {
	const allTasks = Object.values(state.items).flatMap((i) => i.tasks);

	// Notes shouldn't affect progress totals
	const relevant = allTasks.filter((t) => t.payload.kind !== "note");

	// Treat skipped as "done" so we always end at 100%
	const done = relevant.filter(
		(t) => t.status === "completed" || t.status === "failed" || t.status === "skipped"
	).length;

	// Use the planned total when present; fallback for safety
	const total = Math.max(1, state.totalPlanned || relevant.length || 1);

	return { done, total };
};

// derive progress (0..100) from tasks
const getOverallProgress = (state: DownloadProgress) => {
	const { done, total } = getOverallCounts(state);
	return Math.round((done / total) * 100);
};

// derive “errors per item” for the accordion
const getErrorsByItem = (state: DownloadProgress) =>
	Object.values(state.items)
		.map((i) => ({
			itemKey: i.itemKey,
			title: i.title,
			errors: i.tasks.filter((t) => t.status === "failed"),
		}))
		.filter((x) => x.errors.length > 0);

const formSchema = z
	.object({
		selectedOptionsByItem: z.record(
			z.string(),
			z.object({
				types: z.array(z.enum(Object.keys(AssetTypes) as [AssetType, ...AssetType[]])),
				autodownload: z.boolean().optional(),
				source: z.enum(["movie", "collection"]).optional(),
				addToDBOnly: z.boolean().optional(),
			})
		),
	})
	.refine(
		(data) =>
			Object.values(data.selectedOptionsByItem).some(
				(item) => Array.isArray(item.types) && item.types.length > 0
			),
		{
			message: "Please select at least one image type to download.",
			path: ["selectedOptionsByItem.message"],
		}
	);

export interface DownloadModalProps {
	setType: "show" | "movie" | "boxset" | "collection" | "set";
	setTitle: string;
	setAuthor: string;
	setID: string;
	posterSets: PosterSet[];
	autoDownloadDefault?: boolean;
	onMediaItemChange?: (item: MediaItem) => void;
}
interface DuplicateMap {
	[mediaItemKey: string]: {
		options: Array<{
			id: string;
			type: "movie" | "collection";
			image?: PosterFile;
		}>;
		selectedType?: "movie" | "collection" | "";
	};
}

const findDuplicateMediaItems = (items: FormItemDisplay[]): DuplicateMap => {
	return items.reduce((acc: DuplicateMap, item) => {
		if (!acc[item.MediaItemRatingKey]) {
			acc[item.MediaItemRatingKey] = {
				options: [
					{
						id: item.SetID,
						type: item.Set.Type as "movie" | "collection",
						image: item.Set.Poster || item.Set.Backdrop,
					},
				],
			};
		} else {
			acc[item.MediaItemRatingKey].options.push({
				id: item.SetID,
				type: item.Set.Type as "movie" | "collection",
				image: item.Set.Poster || item.Set.Backdrop,
			});
		}
		return acc;
	}, {});
};

const DownloadModal: React.FC<DownloadModalProps> = ({
	setType,
	setTitle,
	setAuthor,
	setID,
	posterSets,
	autoDownloadDefault = true, // Default to false if not provided
	onMediaItemChange,
}) => {
	const router = useRouter();
	const [isMounted, setIsMounted] = useState(false);

	// Download Progress
	const [progress, setProgress] = useState<DownloadProgress>({
		currentText: "",
		totalPlanned: 0,
		items: {},
	});

	// State - Modal Button Texts
	const [buttonTexts, setButtonTexts] = useState({
		cancel: "Cancel",
		download: "Download",
	});

	// State - Selected Types - Files Size and Download Size
	const [selectedSizes, setSelectedSizes] = useState({
		fileCount: 0,
		downloadSize: 0,
	});

	// State - Duplicate Media Items
	const [duplicates, setDuplicates] = useState<DuplicateMap>({});

	// State - Add to Queue Only
	const [addToQueueOnly, setAddToQueueOnly] = useState(false);

	// User Preferences
	const { downloadDefaults } = useUserPreferencesStore();

	// Media Store
	const { setMediaItem } = useMediaStore();

	// Search Query Store
	const { setSearchQuery } = useSearchQueryStore();

	// Cancel Ref
	const cancelRef = useRef(false);

	// Function - Reset Progress Values
	const resetProgress = () => {
		setProgress({
			currentText: "",
			totalPlanned: 0,
			items: {},
		});
	};

	// Function - Close Modal
	const handleClose = () => {
		cancelRef.current = true;
		setIsMounted(false);
		resetProgress();
		setButtonTexts({
			cancel: "Cancel",
			download: "Download",
		});
		form.reset();
	};

	// Function - Handle Link Click
	const getMediuxBaseUrl = () => {
		if (setType === "boxset") {
			return `https://mediux.io/boxset/${setID}`;
		} else {
			return `https://mediux.io/${setType}-set/${setID}`;
		}
	};

	// useMemo - set formItems based on posterSets
	const formItems: FormItemDisplay[] = useMemo(() => {
		const items: FormItemDisplay[] = [];
		posterSets
			.sort((a, b) => a.Title.localeCompare(b.Title))
			.forEach((set) => {
				const getItems = posterSetToFormItem(set);
				if (getItems) {
					items.push(...getItems);
				}
			});

		// Make sure that the main item always appears first
		if (setType === "boxset" || setType === "collection") {
			items.sort((a, b) => {
				if (a.Set.ID === setID) return -1;
				if (b.Set.ID === setID) return 1;
				return 0;
			});
		}

		return items;
	}, [posterSets, setID, setType]);

	// Compute Asset Types based on what the form item has
	const computeAssetTypes = (item: FormItemDisplay): AssetType[] => {
		if (!item.Set) return [];

		const setHasPoster = item.Set.Poster;
		const setHasBackdrop = item.Set.Backdrop;
		const setHasSeasonPosters = item.Set.SeasonPosters?.some(
			(sp) => sp.Season?.Number !== 0 && sp.Show?.MediaItem.RatingKey
		);
		const setHasSpecialSeasonPosters = item.Set.SeasonPosters?.some(
			(sp) => sp.Season?.Number === 0 && sp.Show?.MediaItem.RatingKey
		);
		const setHasTitleCards =
			item.Set.TitleCards &&
			item.Set.TitleCards.length > 0 &&
			item.Set.TitleCards.some((tc) => tc.Show?.MediaItem.RatingKey);
		const types: (AssetType | null)[] = [
			setHasPoster ? "poster" : null,
			setHasBackdrop ? "backdrop" : null,
			setHasSeasonPosters ? "seasonPoster" : null,
			setHasSpecialSeasonPosters ? "specialSeasonPoster" : null,
			setHasTitleCards ? "titlecard" : null,
		];

		return types.filter((type): type is AssetType => type !== null);
	};

	// Define Form
	const form = useForm<z.infer<typeof formSchema>>({
		resolver: zodResolver(formSchema),
		mode: "onChange",
		defaultValues: {
			selectedOptionsByItem: formItems.reduce(
				(acc, item) => {
					acc[item.MediaItemRatingKey] = {
						types: computeAssetTypes(item).filter((type) => downloadDefaults.includes(type)),
						autodownload: item.Set.Type === "show" ? autoDownloadDefault : false,
						addToDBOnly: false,
						source: item.Set.Type === "movie" || item.Set.Type === "collection" ? item.Set.Type : undefined,
					};
					return acc;
				},
				{} as z.infer<typeof formSchema>["selectedOptionsByItem"]
			),
		},
	});

	const watchSelectedOptions = useWatch({
		control: form.control,
		name: "selectedOptionsByItem",
	});

	// Reset form on mount
	useEffect(() => {
		form.reset({
			selectedOptionsByItem: formItems.reduce(
				(acc, item) => {
					acc[item.MediaItemRatingKey] = {
						types: computeAssetTypes(item).filter((type) => downloadDefaults.includes(type)),
						autodownload: item.Set.Type === "show" ? autoDownloadDefault : false,
						addToDBOnly: false,
						source: item.Set.Type === "movie" || item.Set.Type === "collection" ? item.Set.Type : undefined,
					};
					return acc;
				},
				{} as z.infer<typeof formSchema>["selectedOptionsByItem"]
			),
		});
	}, [formItems, form, autoDownloadDefault, downloadDefaults]);

	useEffect(() => {
		// If all the form items are set to "Add to Database Only", change button text
		if (Object.values(watchSelectedOptions).every((option) => option.addToDBOnly || option.types.length === 0)) {
			setButtonTexts((prev) => ({
				...prev,
				download: "Add to Database",
			}));
		} else if (addToQueueOnly) {
			setButtonTexts((prev) => ({
				...prev,
				download: "Add to Queue",
			}));
		} else {
			setButtonTexts((prev) => ({
				...prev,
				download: "Download",
			}));
		}
	}, [addToQueueOnly, watchSelectedOptions]);

	useEffect(() => {
		const dups = findDuplicateMediaItems(formItems);
		setDuplicates(
			Object.fromEntries(
				Object.entries(dups)
					.filter(([, value]) => value.options.length > 1)
					.map(([key, value]) => {
						// Set initial selection to "movie" if available
						const hasMovie = value.options.some((opt) => opt.type === "movie");
						return [
							key,
							{
								...value,
								selectedType: hasMovie ? "movie" : "collection",
							},
						];
					})
			)
		);

		// Update form values for duplicates
		Object.entries(dups)
			.filter(([, value]) => value.options.length > 1)
			.forEach(([key, value]) => {
				const hasMovie = value.options.some((opt) => opt.type === "movie");

				// Uncheck collection if movie is available
				if (hasMovie) {
					form.setValue(`selectedOptionsByItem.${key}`, {
						...form.getValues(`selectedOptionsByItem.${key}`),
						source: "movie",
					});
				}
			});
	}, [formItems, form]);

	useEffect(() => {
		const calculateSizes = () => {
			let totalFiles = 0;
			let totalSize = 0;

			// Iterate through each selected item
			Object.entries(watchSelectedOptions).forEach(([ratingKey, selection]) => {
				// Find the corresponding form item to access PosterFile sizes
				const formItem = formItems.find((item) => item.MediaItemRatingKey === ratingKey);

				if (!formItem || !selection.types || selection.addToDBOnly) return;

				// Count files and sum sizes for each selected type
				selection.types.forEach((type) => {
					switch (type) {
						case "poster":
							if (formItem.Set.Poster?.FileSize) {
								totalSize += formItem.Set.Poster.FileSize;
								totalFiles += 1;
							}
							break;
						case "backdrop":
							if (formItem.Set.Backdrop?.FileSize) {
								totalSize += formItem.Set.Backdrop.FileSize;
								totalFiles += 1;
							}
							break;
						case "seasonPoster":
							formItem.Set.SeasonPosters?.forEach((sp) => {
								if (sp.Season?.Number !== 0 && sp.FileSize) {
									totalSize += sp.FileSize;
									totalFiles += 1;
								}
							});
							break;
						case "specialSeasonPoster":
							formItem.Set.SeasonPosters?.forEach((sp) => {
								if (sp.Season?.Number === 0 && sp.FileSize) {
									totalSize += sp.FileSize;
									totalFiles += 1;
								}
							});
							break;
						case "titlecard":
							formItem.Set.TitleCards?.forEach((tc) => {
								if (
									tc.FileSize &&
									tc.Episode?.SeasonNumber !== undefined &&
									tc.Episode?.EpisodeNumber !== undefined
								) {
									if (
										tc.Episode.SeasonNumber === 0 &&
										!selection.types.includes("specialSeasonPoster")
									) {
										return; // Skip if it's a special season poster
									}
									totalSize += tc.FileSize;
									totalFiles += 1;
								}
							});
							break;
					}
				});
			});

			setSelectedSizes({
				fileCount: totalFiles,
				downloadSize: totalSize,
			});
		};

		calculateSizes();
	}, [formItems, watchSelectedOptions]);

	const LOG_VALUES = () => {
		log("INFO", "Download Modal", "Debug Info", "Logging props values:", {
			setType,
			setTitle,
			setAuthor,
			setID,
			posterSets,
		});
		log("INFO", "Download Modal", "Debug Info", "Logging form items:", formItems);
		log("INFO", "Download Modal", "Debug Info", "Logging form values:", form);
		log("INFO", "Download Modal", "Debug Info", "Logging watch selected types:", watchSelectedOptions);
		log("INFO", "Download Modal", "Debug Info", "Logging progress:", progress);
		log("INFO", "Download Modal", "Debug Info", "Logging duplicates:", duplicates);
	};

	// --- Progress/Task Helpers ---
	const setCurrentText = (text: string) => {
		setProgress((prev) => ({
			...prev,
			currentText: text,
		}));
	};

	const upsertItem = (itemKey: string, title: string) => {
		setProgress((prev) => {
			if (prev.items[itemKey]) return prev;
			return {
				...prev,
				items: {
					...prev.items,
					[itemKey]: { itemKey, title, tasks: [] },
				},
			};
		});
	};

	const addTask = (itemKey: string, title: string, task: Task) => {
		setProgress((prev) => {
			const existing = prev.items[itemKey] ?? { itemKey, title, tasks: [] as Task[] };
			return {
				...prev,
				items: {
					...prev.items,
					[itemKey]: {
						...existing,
						title: existing.title || title,
						tasks: [...existing.tasks, task],
					},
				},
			};
		});
	};

	const updateTask = (taskId: string, updater: (t: Task) => Task) => {
		setProgress((prev) => {
			let changed = false;
			const nextItems: Record<string, ItemProgress> = {};

			for (const [itemKey, item] of Object.entries(prev.items)) {
				const idx = item.tasks.findIndex((t) => t.id === taskId);
				if (idx === -1) {
					nextItems[itemKey] = item;
					continue;
				}

				const nextTasks = [...item.tasks];
				nextTasks[idx] = updater(nextTasks[idx]);
				nextItems[itemKey] = { ...item, tasks: nextTasks };
				changed = true;
			}

			return changed ? { ...prev, items: nextItems } : prev;
		});
	};

	const findTask = (state: DownloadProgress, taskId: string): Task | undefined => {
		for (const item of Object.values(state.items)) {
			const t = item.tasks.find((x) => x.id === taskId);
			if (t) return t;
		}
		return undefined;
	};

	const getAssetStatus = (itemKey: string, assetType: AssetType): TaskStatus | undefined => {
		const tasks = progress.items[itemKey]?.tasks ?? [];
		const relevant = tasks.filter((t) => t.payload.kind === "download" && t.payload.posterFileType === assetType);
		return relevant.length ? relevant[relevant.length - 1].status : undefined;
	};

	const runDownloadTask = async (taskId: string, payload: DownloadTaskPayload): Promise<boolean> => {
		updateTask(taskId, (t) => ({
			...t,
			status: "in-progress",
			attempts: (t.attempts ?? 0) + 1,
			error: undefined,
		}));
		setCurrentText(`Downloading ${payload.fileName} for "${payload.itemTitle}"`);

		try {
			const response = await patchDownloadPosterFileAndUpdateMediaServer(
				payload.posterFile,
				payload.mediaItem,
				payload.fileName
			);

			if (response.status === "error") {
				throw new Error(response.error?.message || "Unknown error");
			}

			updateTask(taskId, (t) => ({ ...t, status: "completed" }));
			return true;
		} catch (error) {
			const message = `"${payload.itemTitle}" - ${payload.fileName} - ${error instanceof Error ? error.message : "Unknown error"}`;
			updateTask(taskId, (t) => ({ ...t, status: "failed", error: message }));
			return false;
		}
	};

	const runAddToDBTask = async (taskId: string, payload: AddToDBTaskPayload): Promise<boolean> => {
		updateTask(taskId, (t) => ({
			...t,
			status: "in-progress",
			attempts: (t.attempts ?? 0) + 1,
			error: undefined,
		}));
		setCurrentText(`Adding "${payload.itemTitle}" to DB`);

		try {
			const resp = await postAddItemToDB(payload.dbItem);
			if (resp.status === "error") {
				throw new Error(resp.error?.message || (typeof resp.error === "string" ? resp.error : "Unknown error"));
			}
			updateTask(taskId, (t) => ({ ...t, status: "completed" }));
			return true;
		} catch (error) {
			updateTask(taskId, (t) => ({
				...t,
				status: "failed",
				error: `"${payload.itemTitle}" - Add to DB failed - ${error instanceof Error ? error.message : "Unknown error"}`,
			}));
			return false;
		}
	};

	const runAddToQueueTask = async (taskId: string, payload: AddToQueueTaskPayload): Promise<boolean> => {
		updateTask(taskId, (t) => ({
			...t,
			status: "in-progress",
			attempts: (t.attempts ?? 0) + 1,
			error: undefined,
		}));
		setCurrentText(`Adding "${payload.itemTitle}" to queue`);

		try {
			const resp = await postAddToQueue(payload.dbItem);
			if (resp.status === "error") {
				throw new Error(resp.error?.message || (typeof resp.error === "string" ? resp.error : "Unknown error"));
			}
			updateTask(taskId, (t) => ({ ...t, status: "completed" }));
			return true;
		} catch (error) {
			updateTask(taskId, (t) => ({
				...t,
				status: "failed",
				error: `"${payload.itemTitle}" - Add to queue failed - ${error instanceof Error ? error.message : "Unknown error"}`,
			}));
			return false;
		}
	};

	const retryTask = async (taskId: string) => {
		const t = findTask(progress, taskId);
		if (!t) return;

		// NOTE: "note" tasks are not retryable
		if (t.payload.kind === "note") return;

		if (t.payload.kind === "download") {
			await runDownloadTask(taskId, t.payload);
		} else if (t.payload.kind === "addToDB") {
			await runAddToDBTask(taskId, t.payload);
		} else if (t.payload.kind === "addToQueue") {
			await runAddToQueueTask(taskId, t.payload);
		}

		setButtonTexts((prev) => ({ ...prev, download: "Download Again" }));
	};

	const renderFormItemAssetType = (
		field: ControllerRenderProps<
			{ selectedOptionsByItem: Record<string, AssetTypeFormValues> },
			`selectedOptionsByItem.${string}`
		>,
		assetType: AssetType,
		item: FormItemDisplay
	) => {
		const types = field.value?.types || [];
		const isDuplicate = duplicates[item.MediaItemRatingKey];

		// Calculate checked state
		const isChecked = types.includes(assetType) && (!isDuplicate || isDuplicate.selectedType === item.Set.Type);

		// Calculate disabled state
		const isDisabled = Boolean(
			isDuplicate && isDuplicate.selectedType && isDuplicate.selectedType !== item.Set.Type
		);

		// Check download status from task state
		const status = getAssetStatus(item.MediaItemRatingKey, assetType);
		const isDownloaded = status === "completed";
		const isFailed = status === "failed";
		const isLoading = status === "in-progress";

		// Check if this assetType is already downloaded in another set (using DBSavedSets)
		let isDownloadedInAnotherSet = false;
		if (!isDownloaded && !isLoading && item.MediaItem?.DBSavedSets) {
			isDownloadedInAnotherSet = item.MediaItem.DBSavedSets.some((set) => {
				const found =
					set.PosterSetID !== item.Set.ID &&
					Array.isArray(set.SelectedTypes) &&
					set.SelectedTypes.some((typeStr) =>
						typeStr
							.split(",")
							.map((t) => t.trim())
							.includes(assetType)
					);
				return found;
			});
		}

		return (
			<FormItem key={`${field.name}-${assetType}`} className="flex flex-row items-start space-x-2">
				<FormControl className="mt-1">
					<Checkbox
						checked={isChecked}
						disabled={isDisabled}
						onCheckedChange={(checked) => {
							const newTypes = checked
								? [...types, assetType]
								: types.filter((type: string) => type !== assetType);

							// Update duplicates tracking
							if (checked && isDuplicate) {
								setDuplicates((prev) => ({
									...prev,
									[item.MediaItemRatingKey]: {
										...prev[item.MediaItemRatingKey],
										selectedType: item.Set.Type as "movie" | "collection",
									},
								}));
							} else if (!checked && isDuplicate) {
								setDuplicates((prev) => ({
									...prev,
									[item.MediaItemRatingKey]: {
										...prev[item.MediaItemRatingKey],
										selectedType: "",
									},
								}));
							}

							field.onChange({
								...(field.value ?? {}),
								types: newTypes,
								autodownload: newTypes.length === 0 ? false : field.value?.autodownload,
								addToDBOnly: newTypes.length === 0 ? false : field.value?.addToDBOnly,
								source:
									item.Set.Type === "movie" || item.Set.Type === "collection"
										? item.Set.Type
										: undefined,
							});
						}}
						className="h-5 w-5 sm:h-4 sm:w-4 cursor-pointer"
					/>
				</FormControl>
				<FormLabel className="text-md font-normal cursor-pointer">
					{assetType.charAt(0).toUpperCase() + assetType.slice(1).replace(/([A-Z])/g, " $1")}
				</FormLabel>
				{isDownloaded ? (
					<Check className="h-4 w-4 text-green-500 mt-1" strokeWidth={3} />
				) : isFailed ? (
					<X className="h-4 w-4 text-destructive mt-1" strokeWidth={3} />
				) : isLoading ? (
					<Loader className="h-4 w-4 text-yellow-500 mt-1 animate-spin" />
				) : isDownloadedInAnotherSet ? (
					<PopoverHelp
						ariaLabel="Type already downloaded in another set"
						side="right"
						className="max-w-xs"
						trigger={<TriangleAlert className="h-4 w-4 mt-1 text-yellow-500 cursor-help" />}
					>
						<div className="flex items-center">
							<CircleAlert className="h-5 w-5 text-yellow-500 mr-2" />
							<span className="text-xs">
								{AssetTypes[assetType]} {AssetTypes[assetType].endsWith("s") ? "have" : "has"} already
								been downloaded in set{" "}
								{
									item.MediaItem?.DBSavedSets?.find(
										(set) =>
											Array.isArray(set.SelectedTypes) &&
											set.SelectedTypes.some((typeStr) =>
												typeStr
													.split(",")
													.map((t) => t.trim())
													.includes(assetType)
											)
									)?.PosterSetID
								}{" "}
								by user{" "}
								{
									item.MediaItem?.DBSavedSets?.find(
										(set) =>
											Array.isArray(set.SelectedTypes) &&
											set.SelectedTypes.some((typeStr) =>
												typeStr
													.split(",")
													.map((t) => t.trim())
													.includes(assetType)
											)
									)?.PosterSetUser
								}
								.
							</span>
						</div>
					</PopoverHelp>
				) : null}
			</FormItem>
		);
	};

	const renderFormItem = (item: FormItemDisplay) => {
		const isDuplicate = duplicates[item.MediaItemRatingKey];
		// Calculate disabled state
		const isDisabled = Boolean(
			isDuplicate && isDuplicate.selectedType && isDuplicate.selectedType !== item.Set.Type
		);

		// Calculate whether the item is already in the database
		const isInDatabase = item.MediaItem && item.MediaItem.ExistInDatabase;

		// Calculate whether the item is already in the database and has this set saved
		const isInDatabaseWithSet =
			isInDatabase &&
			item.MediaItem.DBSavedSets &&
			item.MediaItem.DBSavedSets.some((set) => set.PosterSetID === item.Set.ID);

		// Calculate whether the item has any error tasks
		const itemProgress = progress.items[item.MediaItemRatingKey];
		const hasErrorTasks = itemProgress ? itemProgress.tasks.some((t) => t.status === "failed") : false;

		return (
			<FormField
				key={`${item.MediaItemRatingKey}-${item.Set.Type}`}
				control={form.control}
				name={`selectedOptionsByItem.${item.MediaItemRatingKey}`}
				render={({ field }) => (
					<div
						className={cn("rounded-md border p-4 rounded-lg mb-4", {
							"border-green-500": isInDatabaseWithSet,
							"border-yellow-500": isInDatabase && !isInDatabaseWithSet,
							"border-destructive": hasErrorTasks,
						})}
					>
						<FormLabel
							className="text-md font-normal mb-4"
							onDoubleClick={() => {
								setMediaItem(item.MediaItem);
							}}
						>
							<span className="flex items-center justify-between w-full">
								{item.MediaItemTitle}
								{item.MediaItem && item.MediaItem.ExistInDatabase && item.MediaItem.DBSavedSets && (
									<Popover modal={true}>
										<PopoverTrigger>
											<Database
												className={cn(
													"h-4 w-4 ml-2 cursor-pointer",
													isInDatabaseWithSet ? "text-green-500" : "text-yellow-500"
												)}
											/>
										</PopoverTrigger>
										<PopoverContent
											className={cn(
												"max-w-[400px] rounded-lg shadow-lg border-2 p-2 flex flex-col items-center justify-center",
												isInDatabaseWithSet ? "border-green-800" : "border-yellow-800"
											)}
										>
											<div className="flex items-center mb-2">
												<CircleAlert className="h-5 w-5 text-yellow-500 mr-2" />
												<span className="text-sm text-muted-foreground">
													This media item already exists in your database
												</span>
											</div>
											<div className="text-xs text-muted-foreground mb-2">
												You have previously saved it in the following sets
											</div>
											<ul className="space-y-2">
												{item.MediaItem.DBSavedSets.map((set) => (
													<li
														key={set.PosterSetID}
														className="flex items-center rounded-md px-2 py-1 shadow-sm"
													>
														<Button
															variant="outline"
															className={cn(
																"flex items-center transition-colors rounded-md px-2 py-1 cursor-pointer text-sm",
																set.PosterSetID.toString() === item.SetID.toString()
																	? "text-green-600  hover:bg-green-100  hover:text-green-600"
																	: "text-yellow-600 hover:bg-yellow-100 hover:text-yellow-700"
															)}
															aria-label={`View saved set ${set.PosterSetID} ${set.PosterSetUser ? `by ${set.PosterSetUser}` : ""}`}
															onClick={(e) => {
																e.stopPropagation();
																setSearchQuery(
																	`${item.MediaItem.Title} Y:${item.MediaItem.Year}: ID:${item.MediaItem.TMDB_ID}: L:${item.MediaItem.LibraryTitle}:`
																);
																router.push("/saved-sets");
															}}
														>
															Set ID: {set.PosterSetID}
															{set.PosterSetUser ? ` by ${set.PosterSetUser}` : ""}
														</Button>
													</li>
												))}
											</ul>
										</PopoverContent>
									</Popover>
								)}
								{setType === "boxset" &&
									isDuplicate &&
									isDuplicate.selectedType !== "" &&
									isDuplicate.selectedType !== item.Set.Type && (
										<Popover modal={true}>
											<PopoverTrigger>
												<TriangleAlert className="h-4 w-4 text-yellow-500 cursor-help" />
											</PopoverTrigger>
											<PopoverContent className="max-w-[400px] w-60">
												<div className="text-sm text-yellow-500">
													This item is selected in the{" "}
													{isDuplicate.selectedType === "movie"
														? "Movie Set"
														: "Collection Set"}
													.
												</div>
												{(() => {
													const img = isDuplicate.options.find(
														(o) => o.type === isDuplicate.selectedType
													)?.image;
													return img && <AssetImage className="mt-2" image={img} />;
												})()}
											</PopoverContent>
										</Popover>
									)}
							</span>
						</FormLabel>

						<div className="space-y-2">
							{computeAssetTypes(item).map((assetType) => (
								<div key={assetType}>
									{renderFormItemAssetType(
										field as ControllerRenderProps<
											{
												selectedOptionsByItem: Record<string, AssetTypeFormValues>;
											},
											`selectedOptionsByItem.${string}`
										>,
										assetType,
										item
									)}
								</div>
							))}
							<FormLabel className={`text-md font-normal` + (isDisabled ? " text-gray-500" : "")}>
								Download Options
							</FormLabel>
							{(item.Set.Type === "movie" || item.Set.Type === "collection") && (
								<FormItem className="flex items-center space-x-2">
									<FormControl>
										<Checkbox
											checked={
												isDisabled || field.value?.types.length === 0
													? false
													: field.value?.addToDBOnly || false
											}
											disabled={isDisabled || field.value?.types.length === 0}
											onCheckedChange={(checked) => {
												field.onChange({
													...(field.value ?? {}),
													addToDBOnly: checked,
												});
											}}
											className="h-5 w-5 sm:h-4 sm:w-4 cursor-pointer"
										/>
									</FormControl>
									<FormLabel className="text-md font-normal cursor-pointer">
										Add to Database Only
									</FormLabel>

									<DownloadModalPopover type="add-to-db-only" />
								</FormItem>
							)}
							{item.Set.Type === "show" && (
								<>
									<FormItem className="flex items-center space-x-2">
										<FormControl>
											<Checkbox
												checked={field.value?.addToDBOnly || false}
												onCheckedChange={(checked) => {
													field.onChange({
														...(field.value ?? {}),
														addToDBOnly: checked,
													});
												}}
												className="h-5 w-5 sm:h-4 sm:w-4 cursor-pointer"
											/>
										</FormControl>
										<FormLabel className="text-md font-normal cursor-pointer">
											Future Updates Only
										</FormLabel>
										<DownloadModalPopover type="future-updated-only" />
									</FormItem>
									<FormItem className="flex items-center space-x-2">
										<FormControl>
											<Checkbox
												checked={field.value?.autodownload || false}
												onCheckedChange={(checked) => {
													field.onChange({
														...(field.value ?? {}),
														autodownload: checked,
													});
												}}
												className="h-5 w-5 sm:h-4 sm:w-4 cursor-pointer"
											/>
										</FormControl>
										<FormLabel className="text-md font-normal cursor-pointer">
											Auto Download
										</FormLabel>
										<DownloadModalPopover type="autodownload" />
									</FormItem>
								</>
							)}
						</div>
					</div>
				)}
			/>
		);
	};

	const createDBItem = (
		item: FormItemDisplay,
		options: z.infer<typeof formSchema>["selectedOptionsByItem"][string],
		mediaItem: MediaItem
	): {
		dbItem: DBMediaItemWithPosterSets;
	} => {
		return {
			dbItem: {
				TMDB_ID: mediaItem.TMDB_ID,
				LibraryTitle: mediaItem.LibraryTitle,
				MediaItem: mediaItem as MediaItem,
				PosterSets: [
					{
						PosterSetID: item.Set.ID,
						PosterSet: item.Set,
						SelectedTypes: options.types,
						AutoDownload: options.autodownload || false,
						LastDownloaded: "",
						ToDelete: false,
					},
				],
			},
		};
	};

	// Compute how many tasks we *expect* to run before starting
	const computePlannedTotal = (data: z.infer<typeof formSchema>) => {
		let total = 0;

		for (const item of formItems) {
			const selected = data.selectedOptionsByItem[item.MediaItemRatingKey];
			if (!selected) continue;

			// Skip duplicates that aren't selected
			if (
				(item.Set.Type === "movie" || item.Set.Type === "collection") &&
				duplicates[item.MediaItemRatingKey]?.selectedType &&
				duplicates[item.MediaItemRatingKey].selectedType !== item.Set.Type
			) {
				continue;
			}

			// Nothing selected and not DB-only => no tasks
			if ((selected.types?.length ?? 0) === 0 && !selected.addToDBOnly) continue;

			// DB-only => just one task
			if (selected.addToDBOnly) {
				total += 1;
				continue;
			}

			// Add-to-queue-only => just one task (no downloads, no add-to-db)
			if (addToQueueOnly) {
				total += 1;
				continue;
			}

			// Download tasks
			for (const type of selected.types ?? []) {
				switch (type) {
					case "poster":
						if (item.Set.Poster) total += 1;
						break;
					case "backdrop":
						if (item.Set.Backdrop) total += 1;
						break;
					case "seasonPoster":
						total += item.Set.SeasonPosters?.filter((sp) => sp.Season?.Number !== 0).length ?? 0;
						break;
					case "specialSeasonPoster":
						total += item.Set.SeasonPosters?.filter((sp) => sp.Season?.Number === 0).length ?? 0;
						break;
					case "titlecard":
						total +=
							item.Set.TitleCards?.filter(
								(tc) =>
									tc.Episode?.SeasonNumber !== undefined && tc.Episode?.EpisodeNumber !== undefined
							).length ?? 0;
						break;
				}
			}

			// After downloads, you *attempt* add-to-db (or mark it skipped) => count it as a planned task
			if ((selected.types?.length ?? 0) > 0) total += 1;
		}

		return total;
	};

	const onSubmit = async (data: z.infer<typeof formSchema>) => {
		if (isMounted) return;
		cancelRef.current = false;

		try {
			setIsMounted(true);
			setButtonTexts({
				cancel: "Cancel",
				download: "Starting...",
			});
			resetProgress();
			// Reset + set the planned total up front (fixed total)
			const plannedTotal = computePlannedTotal(data);
			setProgress({
				currentText: "Starting...",
				totalPlanned: plannedTotal,
				items: {},
			});

			// Your download logic here
			log("INFO", "Download Modal", "Debug Info", "Form submitted with data:", data);
			log("INFO", "Download Modal", "Debug Info", "Progress:", progress);
			log("INFO", "Download Modal", "Debug Info", "Selected Types:", { watchSelectedOptions });
			log("INFO", "Download Modal", "Debug Info", "Add to Queue Only:", addToQueueOnly);

			// Sort formItems by MediaItemTitle for consistent order
			const sortedFormItems = formItems;
			log("INFO", "Download Modal", "Debug Info", "Sorted Form Items:", sortedFormItems);

			// Go through each item in the formItems
			for (let idx = 0; idx < sortedFormItems.length; idx++) {
				if (cancelRef.current) return; // Exit if cancelled
				const item = sortedFormItems[idx];
				log("INFO", "Download Modal", "Debug Info", "Processing Item:", item);

				const selectedOptions = data.selectedOptionsByItem[item.MediaItemRatingKey];
				log("INFO", "Download Modal", "Debug Info", "Selected Types for Item:", selectedOptions);

				// If no types are selected and not set to "Add to DB Only", skip this item
				if (selectedOptions.types.length === 0 && !selectedOptions.addToDBOnly) {
					log(
						"INFO",
						"Download Modal",
						"Debug Info",
						`Skipping ${item.MediaItemTitle} - Nothing to do here.`
					);
					continue;
				}

				upsertItem(item.MediaItemRatingKey, item.MediaItemTitle);

				let currentMediaItem: MediaItem | undefined;
				if (item.Set.Type === "movie" || item.Set.Type === "collection") {
					currentMediaItem = item.Set.Poster?.Movie?.MediaItem || item.Set.Backdrop?.Movie?.MediaItem;
				} else if (item.Set.Type === "show") {
					currentMediaItem =
						item.Set.Poster?.Show?.MediaItem ||
						item.Set.Backdrop?.Show?.MediaItem ||
						item.Set.SeasonPosters?.find((poster) => poster.Show?.MediaItem)?.Show?.MediaItem ||
						item.Set.TitleCards?.find((card) => card.Show?.MediaItem)?.Show?.MediaItem;
				}
				if (!currentMediaItem) {
					const noteId = newId();
					addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
						id: noteId,
						status: "failed",
						label: "Resolve media item",
						attempts: 1,
						error: `No MediaItem found for ${item.MediaItemTitle}.`,
						payload: { kind: "note", itemKey: item.MediaItemRatingKey, itemTitle: item.MediaItemTitle },
					});
					continue;
				}

				// If the item set is a movie or collection, check for duplicates
				// If this duplicate type is not selected, skip it
				if (
					(item.Set.Type === "movie" || item.Set.Type === "collection") &&
					duplicates[item.MediaItemRatingKey] &&
					duplicates[item.MediaItemRatingKey].selectedType &&
					duplicates[item.MediaItemRatingKey].selectedType !== item.Set.Type
				) {
					log(
						"INFO",
						"Download Modal",
						"Debug Info",
						`Skipping ${item.MediaItemTitle} in ${item.SetID} - Duplicate type selected: ${duplicates[item.MediaItemRatingKey].selectedType}`
					);
					continue;
				}

				// Get the latest media item from the server
				const latestMediaItemResp = await fetchMediaServerItemContent(
					item.MediaItemRatingKey,
					currentMediaItem.LibraryTitle,
					"mediaitem"
				);
				if (latestMediaItemResp.status === "error" || !latestMediaItemResp.data) {
					const noteId = newId();
					const msg =
						latestMediaItemResp.status === "error"
							? latestMediaItemResp.error?.help ||
								latestMediaItemResp.error?.detail ||
								latestMediaItemResp.error?.message ||
								"Unknown error"
							: "No media item found.";

					addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
						id: noteId,
						status: "failed",
						label: "Fetch latest media item",
						attempts: 1,
						error: `Error fetching latest media item for ${item.MediaItemTitle}: ${msg}`,
						payload: { kind: "note", itemKey: item.MediaItemRatingKey, itemTitle: item.MediaItemTitle },
					});
					continue;
				}

				const latestMediaItem = latestMediaItemResp.data.mediaItem;
				const createdSavedItem = createDBItem(item, selectedOptions, latestMediaItem);

				// Add to DB only
				if (selectedOptions.addToDBOnly) {
					const taskId = newId();
					addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
						id: taskId,
						status: "pending",
						label: `Add "${item.MediaItemTitle}" to DB`,
						attempts: 0,
						payload: {
							kind: "addToDB",
							itemKey: item.MediaItemRatingKey,
							itemTitle: item.MediaItemTitle,
							dbItem: createdSavedItem.dbItem,
						},
					});

					await runAddToDBTask(taskId, {
						kind: "addToDB",
						itemKey: item.MediaItemRatingKey,
						itemTitle: item.MediaItemTitle,
						dbItem: createdSavedItem.dbItem,
					});

					if (onMediaItemChange) {
						latestMediaItem.ExistInDatabase = true;
						onMediaItemChange(latestMediaItem);
					}
					continue;
				}

				const selectedTypes = selectedOptions.types.sort((a, b) => {
					const order = ["poster", "backdrop", "seasonPoster", "specialSeasonPoster", "titlecard"];
					return order.indexOf(a) - order.indexOf(b);
				});

				// If no types are selected, skip this item
				if (selectedTypes.length === 0) {
					log("INFO", "Download Modal", "Debug Info", `Skipping ${item.MediaItemTitle} - No types selected.`);
					continue;
				}

				log(
					"INFO",
					"Download Modal",
					"Debug Info",
					`Selected Types for ${item.MediaItemTitle}:`,
					selectedTypes
				);

				// Add to queue only (for items that actually have types selected)
				if (addToQueueOnly) {
					const taskId = newId();
					addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
						id: taskId,
						status: "pending",
						label: `Add "${item.MediaItemTitle}" to queue`,
						attempts: 0,
						payload: {
							kind: "addToQueue",
							itemKey: item.MediaItemRatingKey,
							itemTitle: item.MediaItemTitle,
							dbItem: createdSavedItem.dbItem,
						},
					});

					await runAddToQueueTask(taskId, {
						kind: "addToQueue",
						itemKey: item.MediaItemRatingKey,
						itemTitle: item.MediaItemTitle,
						dbItem: createdSavedItem.dbItem,
					});
					continue;
				}

				// Track whether THIS item had at least one successful download
				let downloadedAtLeastOneForItem = false;

				for (const type of selectedTypes) {
					if (cancelRef.current) return; // Exit if cancelled
					switch (type) {
						case "poster":
							if (item.Set.Poster) {
								const taskId = newId();
								const payload: DownloadTaskPayload = {
									kind: "download",
									itemKey: item.MediaItemRatingKey,
									itemTitle: item.MediaItemTitle,
									posterFile: item.Set.Poster,
									posterFileType: "poster",
									fileName: "Poster",
									mediaItem: latestMediaItem,
								};

								addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
									id: taskId,
									status: "pending",
									label: payload.fileName,
									attempts: 0,
									payload,
								});

								const okPoster = await runDownloadTask(taskId, payload);
								downloadedAtLeastOneForItem = downloadedAtLeastOneForItem || okPoster;
							}
							break;

						case "backdrop":
							if (item.Set.Backdrop) {
								const taskId = newId();
								const payload: DownloadTaskPayload = {
									kind: "download",
									itemKey: item.MediaItemRatingKey,
									itemTitle: item.MediaItemTitle,
									posterFile: item.Set.Backdrop,
									posterFileType: "backdrop",
									fileName: "Backdrop",
									mediaItem: latestMediaItem,
								};

								addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
									id: taskId,
									status: "pending",
									label: payload.fileName,
									attempts: 0,
									payload,
								});

								const okBackdrop = await runDownloadTask(taskId, payload);
								downloadedAtLeastOneForItem = downloadedAtLeastOneForItem || okBackdrop;
							}
							break;

						case "seasonPoster":
							if (item.Set.SeasonPosters) {
								const sorted = item.Set.SeasonPosters.filter((sp) => sp.Season?.Number !== 0).sort(
									(a, b) => (a.Season?.Number || 0) - (b.Season?.Number || 0)
								);

								for (const sp of sorted) {
									const seasonNum = sp.Season?.Number;
									const seasonExists = latestMediaItem.Series?.Seasons.some(
										(season) => season.SeasonNumber === seasonNum
									);

									const fileName = `Season ${seasonNum?.toString().padStart(2, "0")} Poster`;

									const taskId = newId();
									const payload: DownloadTaskPayload = {
										kind: "download",
										itemKey: item.MediaItemRatingKey,
										itemTitle: item.MediaItemTitle,
										posterFile: sp,
										posterFileType: "seasonPoster",
										fileName,
										mediaItem: latestMediaItem,
									};

									addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
										id: taskId,
										status: "pending",
										label: fileName,
										attempts: 0,
										payload,
									});

									if (!seasonExists) {
										updateTask(taskId, (t) => ({
											...t,
											status: "skipped",
										}));
										continue;
									}

									const okSeason = await runDownloadTask(taskId, payload);
									downloadedAtLeastOneForItem = downloadedAtLeastOneForItem || okSeason;
								}
							}
							break;

						case "specialSeasonPoster":
							if (item.Set.SeasonPosters) {
								const specials = item.Set.SeasonPosters.filter((sp) => sp.Season?.Number === 0);

								for (const sp of specials) {
									const exists = latestMediaItem.Series?.Seasons.some(
										(season) => season.SeasonNumber === 0
									);

									const taskId = newId();
									const payload: DownloadTaskPayload = {
										kind: "download",
										itemKey: item.MediaItemRatingKey,
										itemTitle: item.MediaItemTitle,
										posterFile: sp,
										posterFileType: "specialSeasonPoster",
										fileName: `"Special Season Poster"`,
										mediaItem: latestMediaItem,
									};

									addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
										id: taskId,
										status: "pending",
										label: payload.fileName,
										attempts: 0,
										payload,
									});

									if (!exists) {
										updateTask(taskId, (t) => ({
											...t,
											status: "skipped",
										}));
										continue;
									}

									const okSpecialSeason = await runDownloadTask(taskId, payload);
									downloadedAtLeastOneForItem = downloadedAtLeastOneForItem || okSpecialSeason;
								}
							}
							break;

						case "titlecard":
							if (item.Set.TitleCards) {
								const sorted = item.Set.TitleCards.filter(
									(tc) =>
										tc.Episode?.SeasonNumber !== undefined &&
										tc.Episode?.EpisodeNumber !== undefined
								).sort((a, b) => {
									const sa = a.Episode?.SeasonNumber || 0;
									const sb = b.Episode?.SeasonNumber || 0;
									const ea = a.Episode?.EpisodeNumber || 0;
									const eb = b.Episode?.EpisodeNumber || 0;
									return sa - sb || ea - eb;
								});

								for (const tc of sorted) {
									const season = tc.Episode!.SeasonNumber;
									const episodeNum = tc.Episode!.EpisodeNumber;

									const episodeExists = latestMediaItem.Series?.Seasons.flatMap(
										(s) => s.Episodes || []
									).some((e) => e.SeasonNumber === season && e.EpisodeNumber === episodeNum);

									const fileName = `"S${season.toString().padStart(2, "0")}E${episodeNum
										.toString()
										.padStart(2, "0")} Title Card`;

									const taskId = newId();
									const payload: DownloadTaskPayload = {
										kind: "download",
										itemKey: item.MediaItemRatingKey,
										itemTitle: item.MediaItemTitle,
										posterFile: tc,
										posterFileType: "titlecard",
										fileName,
										mediaItem: latestMediaItem,
									};

									addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
										id: taskId,
										status: "pending",
										label: fileName,
										attempts: 0,
										payload,
									});

									if (!episodeExists) {
										updateTask(taskId, (t) => ({
											...t,
											status: "skipped",
										}));
										continue;
									}

									const okTitleCard = await runDownloadTask(taskId, payload);
									downloadedAtLeastOneForItem = downloadedAtLeastOneForItem || okTitleCard;
								}
							}
							break;
					}
				}

				// Only add to DB if at least one download succeeded
				if (!downloadedAtLeastOneForItem) {
					const taskId = newId();
					addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
						id: taskId,
						status: "skipped",
						label: `Add "${item.MediaItemTitle}" to DB (skipped: no successful downloads)`,
						attempts: 0,
						payload: {
							kind: "addToDB",
							itemKey: item.MediaItemRatingKey,
							itemTitle: item.MediaItemTitle,
							dbItem: createdSavedItem.dbItem,
						},
					});
					continue;
				}

				const addId = newId();
				addTask(item.MediaItemRatingKey, item.MediaItemTitle, {
					id: addId,
					status: "pending",
					label: `Add "${item.MediaItemTitle}" to DB`,
					attempts: 0,
					payload: {
						kind: "addToDB",
						itemKey: item.MediaItemRatingKey,
						itemTitle: item.MediaItemTitle,
						dbItem: createdSavedItem.dbItem,
					},
				});

				const ok = await runAddToDBTask(addId, {
					kind: "addToDB",
					itemKey: item.MediaItemRatingKey,
					itemTitle: item.MediaItemTitle,
					dbItem: createdSavedItem.dbItem,
				});

				if (ok && onMediaItemChange) {
					latestMediaItem.ExistInDatabase = true;
					onMediaItemChange(latestMediaItem);
				}
			}

			setCurrentText("Completed!");
			setButtonTexts({
				cancel: "Close",
				download: "Download Again",
			});
			setIsMounted(false);
		} catch (error) {
			const taskId = newId();
			addTask("general", "General", {
				id: taskId,
				status: "failed",
				label: "Unexpected error",
				attempts: 1,
				error: error instanceof Error ? error.message : "An unknown error occurred",
				payload: { kind: "note", itemKey: "general", itemTitle: "General" },
			});

			setButtonTexts({
				cancel: "Close",
				download: "Retry Download",
			});
			setIsMounted(false);
		}
	};

	return (
		<Dialog
			onOpenChange={(open) => {
				if (!open) {
					handleClose();
				}
			}}
		>
			<DialogTrigger asChild>
				<Download className="mr-2 h-5 w-5 sm:h-7 sm:w-7 cursor-pointer active:scale-95 hover:text-primary" />
			</DialogTrigger>

			<DialogPortal>
				<DialogOverlay />
				<DialogContent
					className={cn("z-50", "max-h-[80vh] overflow-y-auto", "sm:max-w-[700px]", "border border-primary")}
				>
					<DialogHeader>
						<DialogTitle onClick={LOG_VALUES}>{setTitle}</DialogTitle>
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
						<DialogDescription>
							<Link
								href={getMediuxBaseUrl()}
								className="hover:text-primary transition-colors text-sm text-muted-foreground"
								target="_blank"
								rel="noopener noreferrer"
							>
								{setType === "boxset" ? "Boxset" : "Set"} ID: {setID}
							</Link>
						</DialogDescription>
					</DialogHeader>

					{formItems.length > 0 ? (
						<Form {...form}>
							<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-2">
								{/* Form Items */}
								{formItems.map(renderFormItem)}

								{/* If the all items have no types selected, show a message */}
								{formItems.every(
									(item) =>
										!watchSelectedOptions?.[item.MediaItemRatingKey]?.types?.length &&
										!watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
								) && (
									<div className="text-sm text-destructive">
										No image types selected for download. Please select at least one image type to
										download.
									</div>
								)}

								{/* Add to Queue 
									Only show this button if at least one item has types selected for download and not set to Add to DB Only
								*/}
								{formItems.some(
									(item) =>
										watchSelectedOptions?.[item.MediaItemRatingKey]?.types?.length > 0 &&
										!watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
								) && (
									<FormItem className="flex items-center space-x-2 mb-4">
										<FormControl>
											<Checkbox
												checked={addToQueueOnly}
												onCheckedChange={(checked) => setAddToQueueOnly(checked ? true : false)}
											/>
										</FormControl>
										<FormLabel className="text-md font-normal cursor-pointer">
											Add to Queue
										</FormLabel>
										<DownloadModalPopover type="add-to-queue-only" />
									</FormItem>
								)}
								<div>
									{/* Number of Images and Total Download Size (only if there are images to download) */}
									{selectedSizes.fileCount > 0 &&
										formItems.some(
											(item) => !watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
										) && (
											<>
												<div className="text-sm text-muted-foreground">
													Number of Images: {selectedSizes.fileCount}
												</div>
												<div className="text-sm text-muted-foreground">
													Total Download Size: ~
													{formatDownloadSize(selectedSizes.downloadSize)}
												</div>
											</>
										)}

									{/* Always show the database-only message if any item is set to Add to DB Only or has types selected */}
									{formItems.some(
										(item) => watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
									) && (
										<div className="text-sm text-muted-foreground mt-1">
											* Will add{" "}
											{(() => {
												const titles = formItems
													.filter(
														(item) =>
															watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
													)
													.map((item) => item.MediaItemTitle);

												if (titles.length === 0) return "";
												if (titles.length === 1)
													return (
														<span className="font-medium text-yellow-500">{`'${titles[0]}' `}</span>
													);
												if (titles.length === 2)
													return (
														<>
															<span className="font-medium text-yellow-500">{`'${titles[0]}'`}</span>
															{" and "}
															<span className="font-medium text-yellow-500">{`'${titles[1]}' `}</span>
														</>
													);

												return (
													<>
														{titles.slice(0, -1).map((title, idx) => (
															<Fragment key={title}>
																<span className="font-medium text-yellow-500">{`'${title}'`}</span>
																{idx < titles.length - 2 ? ", " : ""}
															</Fragment>
														))}
														{" and "}
														<span className="font-medium text-yellow-500">{`'${titles[titles.length - 1]}' `}</span>
													</>
												);
											})()}
											to database without downloading any images.
										</div>
									)}
								</div>

								{/* Progress Bar */}
								{Object.values(progress.items).some((i) => i.tasks.length > 0) &&
									(() => {
										const overall = getOverallProgress(progress);
										const errors = getErrorsByItem(progress);
										const hasErrors = errors.length > 0;

										return (
											<div className="w-full">
												<div className="flex items-center justify-between w-full">
													<div className="relative w-full min-w-0">
														<Progress
															value={overall}
															className={cn(
																"w-full rounded-md overflow-hidden h-3",
																overall < 100 && "animate-pulse h-5",
																overall === 100 && !hasErrors && "[&>div]:bg-green-500",
																overall === 100 && hasErrors && "[&>div]:bg-destructive"
															)}
														/>

														{overall < 100 && (
															<span
																className={cn(
																	"absolute inset-0 flex items-center justify-center",
																	"text-xs text-white pointer-events-none mt-0.5",
																	"px-2 min-w-0"
																)}
																title={progress.currentText}
															>
																<span className="w-full min-w-0 truncate text-center">
																	{progress.currentText}
																</span>
															</span>
														)}
													</div>

													<span className="ml-1 text-sm text-muted-foreground min-w-[30px] text-right">
														{overall}%
													</span>
												</div>
											</div>
										);
									})()}

								{/* Errors */}
								{getErrorsByItem(progress).length > 0 && (
									<div className="my-2">
										{(() => {
											const grouped = getErrorsByItem(progress);
											const errorCount = grouped.reduce((acc, g) => acc + g.errors.length, 0);

											return (
												<Accordion type="single" collapsible>
													<AccordionItem value="errors">
														<AccordionTrigger className="text-destructive">
															Errors ({errorCount})
														</AccordionTrigger>
														<AccordionContent>
															<div className="flex flex-col space-y-2">
																{grouped.map((g) => (
																	<div
																		key={g.itemKey}
																		className="flex flex-col space-y-2"
																	>
																		{g.errors.map((t) => (
																			<div
																				key={t.id}
																				className="flex items-center text-destructive"
																			>
																				{t.payload.kind !== "note" ? (
																					<button
																						type="button"
																						className="mr-2 h-4 w-4 text-yellow-500 cursor-pointer"
																						onClick={() => retryTask(t.id)}
																						aria-label="Retry"
																					>
																						<RefreshCcw className="h-4 w-4" />
																					</button>
																				) : (
																					<span className="mr-1 h-4 w-4">
																						<X className="h-4 w-4" />
																					</span>
																				)}
																				<span>{t.error || t.label}</span>
																			</div>
																		))}
																	</div>
																))}
															</div>
														</AccordionContent>
													</AccordionItem>
												</Accordion>
											);
										})()}
									</div>
								)}

								<DialogFooter>
									<div className="flex space-x-4 justify-end w-full">
										{/* Cancel button to close the modal */}
										<DialogClose asChild>
											<Button
												className="text-destructive border-1 shadow-none hover:text-red-500 cursor-pointer"
												variant="ghost"
												onClick={() => {
													handleClose();
												}}
											>
												{buttonTexts.cancel}
											</Button>
										</DialogClose>

										{/* Download button to display download info */}
										{
											// Only show if at least one item has types selected or is set to "Add to DB Only"
											formItems.some(
												(item) =>
													watchSelectedOptions?.[item.MediaItemRatingKey]?.types?.length >
														0 ||
													watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
											) && (
												<Button
													variant={"outline"}
													className="cursor-pointer hover:text-primary hover:brightness-120 transition-colors"
													disabled={
														// Disable if no items are selected and nothing is set to "Add to DB Only"
														(selectedSizes.fileCount === 0 &&
															formItems.every(
																(item) =>
																	!watchSelectedOptions[item.MediaItemRatingKey].types
																		.length &&
																	!watchSelectedOptions[item.MediaItemRatingKey]
																		.addToDBOnly
															)) ||
														isMounted
													}
												>
													{(() => {
														const isBusy =
															buttonTexts.download.startsWith("Starting") ||
															buttonTexts.download.startsWith("Adding") ||
															buttonTexts.download.startsWith("Downloading");

														const { done, total } = getOverallCounts(progress);

														if (isBusy) {
															return (
																<>
																	<Loader className="h-4 w-4 animate-spin" />
																	{progress.totalPlanned > 0 && done !== total && (
																		<span className="ml-2 text-muted-foreground tabular-nums">
																			{done}/{total}
																		</span>
																	)}
																</>
															);
														}

														return (
															<>
																{buttonTexts.download === "Download" ||
																buttonTexts.download === "Download Again" ||
																buttonTexts.download === "Retry Download" ? (
																	<Download className="h-4 w-4" />
																) : buttonTexts.download === "Add to Queue" ? (
																	<ListEnd className="h-4 w-4" />
																) : buttonTexts.download === "Add to Database" ? (
																	<DatabaseZap className="h-4 w-4" />
																) : null}
																{buttonTexts.download}
															</>
														);
													})()}
												</Button>
											)
										}

										{/* Reset Form button: If the all items have no types selected and nothing is set to "Add to DB Only" */}
										{formItems.every(
											(item) =>
												!watchSelectedOptions?.[item.MediaItemRatingKey]?.types?.length &&
												!watchSelectedOptions?.[item.MediaItemRatingKey]?.addToDBOnly
										) && (
											<Button
												className="cursor-pointer hover:text-primary hover:brightness-120 transition-colors"
												variant="secondary"
												onClick={() => {
													form.reset();
													setSelectedSizes({ fileCount: 0, downloadSize: 0 });
													resetProgress();
													setDuplicates({});
												}}
											>
												<OctagonMinus className="h-4 w-4" />
												Reset Form
											</Button>
										)}
									</div>
								</DialogFooter>
							</form>
						</Form>
					) : (
						<div className="p-4 text-center">
							<Lead className="text-md text-muted-foreground mb-2">No items found in this set.</Lead>
							<p className="text-sm text-muted-foreground">
								Please ensure the set has items with images available for download.
							</p>
						</div>
					)}
				</DialogContent>
			</DialogPortal>
		</Dialog>
	);
};

export default DownloadModal;
