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
	Download,
	ListEnd,
	Loader,
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

// Type for individual item progress
type AssetProgress = {
	poster?: string;
	backdrop?: string;
	seasonPoster?: string;
	specialSeasonPoster?: string;
	titlecard?: string;
	addToDB?: string;
	addToQueue?: string;
	// Number of Files Failed
	seasonPosterFailed?: number;
	titlecardFailed?: number;
};

// Main progress type
type DownloadProgress = {
	// Shared progress bar state
	value: number;
	currentText: string;

	// Individual item progress, keyed by MediaItemRatingKey
	itemProgress: Record<string, AssetProgress>;

	// Shared warning messages
	//warningMessages: string[];
	warningMessages: Record<
		string,
		{
			posterFile: PosterFile | null;
			posterFileType: keyof AssetProgress | null;
			fileName: string | null;
			mediaItem: MediaItem | null;
			message: string;
		}
	>;
};

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
	const [progressValues, setProgressValues] = useState<DownloadProgress>({
		value: 0,
		currentText: "",
		itemProgress: {},
		warningMessages: {},
	});

	// Refs for progress values
	const progressRef = useRef(0);
	const progressIncrementRef = useRef(0);
	const progressDownloadRef = useRef(0);

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
	const resetProgressValues = () => {
		setProgressValues({
			value: 0,
			currentText: "",
			itemProgress: {},
			warningMessages: {},
		});
	};

	// Function - Close Modal
	const handleClose = () => {
		cancelRef.current = true;
		setIsMounted(false);
		resetProgressValues();
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
			setAddToQueueOnly(false);
		}

		if (addToQueueOnly) {
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
		log("INFO", "Download Modal", "Debug Info", "Logging progress values:", progressValues);
		log("INFO", "Download Modal", "Debug Info", "Logging duplicates:", duplicates);
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

		// Check download status from progressValues
		const progress = progressValues.itemProgress?.[item.MediaItemRatingKey]?.[assetType];
		const isDownloaded = progress?.startsWith("Downloaded");
		const isFailed = progress?.startsWith("Failed");
		const isLoading = progress && !isDownloaded && !isFailed;

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
													"h-4 w-4 ml-2 cursor-help",
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

	const updateProgressValue = (value: number) => {
		// Update the ref immediately
		progressRef.current = Math.min(value, 100);

		// Update the state
		setProgressValues((prev) => ({
			...prev,
			value: progressRef.current,
		}));
	};

	const updateItemProgress = (itemKey: string, assetType: keyof AssetProgress, status: string) => {
		setProgressValues((prev) => ({
			...prev,
			currentText: `${status}`,
			itemProgress: {
				...prev.itemProgress,
				[itemKey]: {
					...prev.itemProgress[itemKey],
					[assetType]: status,
				},
			},
		}));
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

	const downloadPosterFileAndUpdateMediaServer = async (
		posterFile: PosterFile,
		posterFileType: keyof AssetProgress,
		fileName: string,
		mediaItem: MediaItem
	) => {
		progressDownloadRef.current += 1;
		setButtonTexts((prev) => ({
			...prev,
			download: `Downloading ${progressDownloadRef.current}/${selectedSizes.fileCount}`,
		}));
		updateItemProgress(mediaItem.RatingKey, posterFileType, `Downloading ${fileName}`);
		try {
			const response = await patchDownloadPosterFileAndUpdateMediaServer(posterFile, mediaItem, fileName);
			updateProgressValue(progressRef.current + progressIncrementRef.current);
			if (response.status === "error") {
				updateItemProgress(mediaItem.RatingKey, posterFileType, `Failed to download ${fileName}`);
				throw new Error(response.error?.message || "Unknown error");
			} else {
				updateItemProgress(mediaItem.RatingKey, posterFileType, `Downloaded ${fileName}`);
			}
		} catch (error) {
			if (posterFileType === "seasonPoster" || posterFileType === "titlecard") {
				setProgressValues((prev) => ({
					...prev,
					warningMessages: {
						...prev.warningMessages,
						[mediaItem.Title]: {
							posterFile: posterFile,
							posterFileType: posterFileType,
							fileName: fileName,
							mediaItem: mediaItem,
							message: `${fileName} - ${error instanceof Error ? error.message : "Unknown error"}`,
						},
					},
					itemProgress: {
						...prev.itemProgress,
						[mediaItem.RatingKey]: {
							...prev.itemProgress[mediaItem.RatingKey],
							// Use specific property name based on the type
							...(posterFileType === "seasonPoster"
								? {
										seasonPosterFailed:
											(prev.itemProgress[mediaItem.RatingKey]?.seasonPosterFailed || 0) + 1,
									}
								: {
										titlecardFailed:
											(prev.itemProgress[mediaItem.RatingKey]?.titlecardFailed || 0) + 1,
									}),
						},
					},
				}));
			} else {
				setProgressValues((prev) => ({
					...prev,
					warningMessages: {
						...prev.warningMessages,
						[mediaItem.Title]: {
							posterFile: posterFile,
							posterFileType: posterFileType,
							fileName: fileName,
							mediaItem: mediaItem,
							message: `${fileName} - ${error instanceof Error ? error.message : "Unknown error"}`,
						},
					},
				}));
			}

			return;
		}
	};

	const onSubmit = async (data: z.infer<typeof formSchema>) => {
		if (isMounted) return;
		cancelRef.current = false;

		try {
			setIsMounted(true);
			setButtonTexts({
				cancel: "Cancel",
				download: "Downloading...",
			});
			progressDownloadRef.current = 0;
			resetProgressValues();

			// Your download logic here
			log("INFO", "Download Modal", "Debug Info", "Form submitted with data:", data);
			log("INFO", "Download Modal", "Debug Info", "Progress Values:", progressValues);
			log("INFO", "Download Modal", "Debug Info", "Selected Types:", { watchSelectedOptions });
			log("INFO", "Download Modal", "Debug Info", "Add to Queue Only:", addToQueueOnly);

			updateProgressValue(1); // Reset progress to 0 at the start

			// File Count + 1 for every item
			progressIncrementRef.current = 95 / (selectedSizes.fileCount + formItems.length);

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
					log(
						"INFO",
						"Download Modal",
						"Debug Info",
						`No MediaItem found for ${item.MediaItemTitle}. Skipping.`
					);
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
				if (latestMediaItemResp.status === "error") {
					setProgressValues((prev) => ({
						...prev,
						warningMessages: {
							...prev.warningMessages,
							[item.MediaItemTitle]: {
								posterFile: null,
								posterFileType: null,
								fileName: null,
								mediaItem: null,
								message: `Error fetching latest media item for ${item.MediaItemTitle}: ${
									latestMediaItemResp.error?.help ||
									latestMediaItemResp.error?.detail ||
									latestMediaItemResp.error?.message ||
									"Unknown error"
								}`,
							},
						},
					}));
					continue;
				}
				if (!latestMediaItemResp.data) {
					setProgressValues((prev) => ({
						...prev,
						warningMessages: {
							...prev.warningMessages,
							[item.MediaItemTitle]: {
								posterFile: null,
								posterFileType: null,
								fileName: null,
								mediaItem: null,
								message: `No media item found for ${item.MediaItemTitle}. Skipping.`,
							},
						},
					}));
					continue;
				}
				const latestMediaItem = latestMediaItemResp.data.mediaItem;

				// If the item is set to "Add to DB Only", create the DB item and skip download
				const createdSavedItem = createDBItem(item, selectedOptions, latestMediaItem);
				if (selectedOptions.addToDBOnly) {
					updateItemProgress(item.MediaItemRatingKey, "addToDB", `Adding ${item.MediaItemTitle} to DB`);
					log("INFO", "Download Modal", "Debug Info", `Adding ${item.MediaItemTitle} to DB only.`);
					log("INFO", "Download Modal", "Debug Info", "DB Item Created:", createdSavedItem.dbItem);
					const addToDBResp = await postAddItemToDB(createdSavedItem.dbItem);
					if (addToDBResp.status === "error") {
						log(
							"INFO",
							"Download Modal",
							"Debug Info",
							`Error adding ${item.MediaItemTitle} to DB:`,
							addToDBResp.error
						);
						setProgressValues((prev) => ({
							...prev,
							warningMessages: {
								...prev.warningMessages,
								[item.MediaItemTitle]: {
									posterFile: null,
									posterFileType: null,
									fileName: null,
									mediaItem: null,
									message: `Error adding ${item.MediaItemTitle} to DB: ${addToDBResp.error?.message || addToDBResp.error || "Unknown error"}`,
								},
							},
						}));
						updateItemProgress(item.MediaItemRatingKey, "addToDB", "Failed to add to DB");
					} else {
						log("INFO", "Download Modal", "Debug Info", `Successfully added ${item.MediaItemTitle} to DB.`);
					}
					updateProgressValue(progressRef.current + progressIncrementRef.current);
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

				if (addToQueueOnly) {
					const createdSavedItem = createDBItem(item, selectedOptions, latestMediaItem);
					log(
						"INFO",
						"Download Modal",
						"Debug Info",
						`Adding ${item.MediaItemTitle} to download queue only.`,
						{ createdSavedItem }
					);
					const addToQueueResp = await postAddToQueue(createdSavedItem.dbItem);
					if (addToQueueResp.status === "error") {
						log(
							"INFO",
							"Download Modal",
							"Debug Info",
							`Error adding ${item.MediaItemTitle} to download queue:`,
							addToQueueResp.error
						);
						setProgressValues((prev) => ({
							...prev,
							warningMessages: {
								...prev.warningMessages,
								[item.MediaItemTitle]: {
									posterFile: null,
									posterFileType: null,
									fileName: null,
									mediaItem: null,
									message: `Error adding ${item.MediaItemTitle} to download queue: ${addToQueueResp.error?.message || addToQueueResp.error || "Unknown error"}`,
								},
							},
						}));
						updateItemProgress(item.MediaItemRatingKey, "addToQueue", "Failed to add to queue");
					} else {
						log(
							"INFO",
							"Download Modal",
							"Debug Info",
							`Successfully added ${item.MediaItemTitle} to download queue.`
						);
						updateItemProgress(item.MediaItemRatingKey, "addToQueue", "Added to queue");
					}
					updateProgressValue(progressRef.current + progressIncrementRef.current);
					continue;
				}

				for (const type of selectedTypes) {
					if (cancelRef.current) return; // Exit if cancelled
					switch (type) {
						case "poster":
							if (item.Set.Poster) {
								log(
									"INFO",
									"Download Modal",
									"Debug Info",
									`Downloading Poster for ${item.MediaItemTitle}`
								);
								await downloadPosterFileAndUpdateMediaServer(
									item.Set.Poster,
									"poster",
									`${item.MediaItemTitle} - Poster`,
									latestMediaItem
								);
							}
							break;
						case "backdrop":
							if (item.Set.Backdrop) {
								log(
									"INFO",
									"Download Modal",
									"Debug Info",
									`Downloading Backdrop for ${item.MediaItemTitle}`
								);
								await downloadPosterFileAndUpdateMediaServer(
									item.Set.Backdrop,
									"backdrop",
									`${item.MediaItemTitle} - Backdrop`,
									latestMediaItem
								);
							}
							break;
						case "seasonPoster":
							if (item.Set.SeasonPosters) {
								log(
									"INFO",
									"Download Modal",
									"Debug Info",
									`Downloading Season Posters for ${item.MediaItemTitle}`
								);
								const sortedSeasonPosters = item.Set.SeasonPosters.filter(
									(sp) => sp.Season?.Number !== 0
								).sort((a, b) => (a.Season?.Number || 0) - (b.Season?.Number || 0));
								for (const sp of sortedSeasonPosters) {
									const seasonExists = latestMediaItem.Series?.Seasons.some(
										(season) => season.SeasonNumber === sp.Season?.Number
									);
									if (!seasonExists) {
										log(
											"INFO",
											"Download Modal",
											"Debug Info",
											`Skipping Season ${sp.Season?.Number} for ${item.MediaItemTitle} - Season does not exist in latest media item.`
										);
										setSelectedSizes((prev) => ({
											...prev,
											fileCount: prev.fileCount - 1,
											downloadSize: sp.FileSize
												? prev.downloadSize - sp.FileSize
												: prev.downloadSize,
										}));
										continue;
									}
									await downloadPosterFileAndUpdateMediaServer(
										sp,
										"seasonPoster",
										`${item.MediaItemTitle} - S${sp.Season?.Number?.toString().padStart(2, "0")} Poster`,
										latestMediaItem
									);
								}
							}
							break;
						case "specialSeasonPoster":
							if (item.Set.SeasonPosters) {
								log(
									"INFO",
									"Download Modal",
									"Debug Info",
									`Downloading Special Season Posters for ${item.MediaItemTitle}`
								);
								const specialSeasonPosters = item.Set.SeasonPosters.filter(
									(sp) => sp.Season?.Number === 0
								);
								for (const sp of specialSeasonPosters) {
									const spExists = latestMediaItem.Series?.Seasons.some(
										(season) => season.SeasonNumber === 0
									);
									if (!spExists) {
										log(
											"INFO",
											"Download Modal",
											"Debug Info",
											`Skipping Special Season Poster for ${item.MediaItemTitle} - Special Season does not exist in latest media item.`
										);
										continue;
									}
									await downloadPosterFileAndUpdateMediaServer(
										sp,
										"specialSeasonPoster",
										`${item.MediaItemTitle} - Special Season Poster`,
										latestMediaItem
									);
								}
							}
							break;
						case "titlecard":
							if (item.Set.TitleCards) {
								log(
									"INFO",
									"Download Modal",
									"Debug Info",
									`Downloading Title Cards for ${item.MediaItemTitle}`
								);
								const sortedTitleCards = item.Set.TitleCards.filter(
									(tc) =>
										tc.Episode?.SeasonNumber !== undefined &&
										tc.Episode?.EpisodeNumber !== undefined
								).sort((a, b) => {
									const seasonA = a.Episode?.SeasonNumber || 0;
									const seasonB = b.Episode?.SeasonNumber || 0;
									const episodeA = a.Episode?.EpisodeNumber || 0;
									const episodeB = b.Episode?.EpisodeNumber || 0;
									return seasonA - seasonB || episodeA - episodeB;
								});

								for (const tc of sortedTitleCards) {
									if (
										tc.FileSize &&
										tc.Episode?.SeasonNumber !== undefined &&
										tc.Episode?.EpisodeNumber !== undefined
									) {
										const episode = latestMediaItem.Series?.Seasons.flatMap(
											(s) => s.Episodes || []
										).find(
											(e) =>
												e.SeasonNumber === tc.Episode?.SeasonNumber &&
												e.EpisodeNumber === tc.Episode?.EpisodeNumber
										);
										if (!episode) {
											log(
												"INFO",
												"Download Modal",
												"Debug Info",
												`Skipping Title Card for S${tc.Episode.SeasonNumber}E${tc.Episode.EpisodeNumber} - Episode does not exist in latest media item.`
											);
											setSelectedSizes((prev) => ({
												...prev,
												fileCount: prev.fileCount - 1,
												downloadSize: prev.downloadSize - tc.FileSize!,
											}));
											continue;
										}
										await downloadPosterFileAndUpdateMediaServer(
											tc,
											"titlecard",
											`${item.MediaItemTitle} - S${tc.Episode.SeasonNumber.toString().padStart(
												2,
												"0"
											)}E${tc.Episode.EpisodeNumber.toString().padStart(2, "0")} Title Card`,
											latestMediaItem
										);
									}
								}
							}
							break;
						default:
							log(
								"INFO",
								"Download Modal",
								"Debug Info",
								`Unknown type ${type} for ${item.MediaItemTitle}. Skipping.`
							);
							continue;
					}
				}
				// Add the item to the database after downloading
				updateItemProgress(item.MediaItemRatingKey, "addToDB", `Adding ${item.MediaItemTitle} to DB`);
				log("INFO", "Download Modal", "Debug Info", `Adding ${item.MediaItemTitle} to DB only.`);
				log("INFO", "Download Modal", "Debug Info", "DB Item Created:", createdSavedItem.dbItem);
				const addToDBResp = await postAddItemToDB(createdSavedItem.dbItem);
				if (addToDBResp.status === "error") {
					log(
						"ERROR",
						"Download Modal",
						"Debug Info",
						`Error adding ${item.MediaItemTitle} to DB:`,
						addToDBResp.error
					);
					setProgressValues((prev) => ({
						...prev,
						warningMessages: {
							...prev.warningMessages,
							[item.MediaItemTitle]: {
								posterFile: null,
								posterFileType: null,
								fileName: null,
								mediaItem: null,
								message: `Error adding ${item.MediaItemTitle} to DB: ${addToDBResp.error?.message || addToDBResp.error || "Unknown error"}`,
							},
						},
					}));
					updateItemProgress(item.MediaItemRatingKey, "addToDB", "Failed to add to DB");
				} else {
					log("INFO", "Download Modal", "Debug Info", `Successfully added ${item.MediaItemTitle} to DB.`);
				}
				updateProgressValue(progressRef.current + progressIncrementRef.current);
				if (onMediaItemChange) {
					latestMediaItem.ExistInDatabase = true;
					onMediaItemChange(latestMediaItem);
				}
			}

			setButtonTexts({
				cancel: "Close",
				download: "Download Again",
			});
			updateProgressValue(100); // Set progress to 100% at the end
			setIsMounted(false);
		} catch (error) {
			log("ERROR", "Download Modal", "Debug Info", "Download Error:", error);
			setProgressValues((prev) => ({
				...prev,
				warningMessages: {
					...prev.warningMessages,
					general: {
						posterFile: null,
						posterFileType: null,
						fileName: null,
						mediaItem: null,
						message: error instanceof Error ? error.message : "An unknown error occurred",
					},
				},
			}));
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
								{progressValues.value > 0 && (
									<div className="w-full">
										<div className="flex items-center justify-between w-full">
											<div className="relative w-full">
												<Progress
													value={progressValues.value}
													className={cn(
														"w-full rounded-md overflow-hidden",
														progressValues.value < 100 && "animate-pulse h-5",
														progressValues.value === 100 && "[&>div]:bg-green-500 h-3",
														progressValues.value === 100 &&
															progressValues.warningMessages &&
															Object.keys(progressValues.warningMessages).length > 0
															? "[&>div]:bg-destructive"
															: ""
													)}
												/>
												{progressValues.value < 100 && (
													<span className="absolute inset-0 flex items-center justify-center text-xs text-white pointer-events-none w-full mt-0.5">
														{progressValues.currentText}
													</span>
												)}
											</div>
											<span className="ml-2 text-sm text-muted-foreground min-w-[40px] text-right">
												{Math.round(progressValues.value)}%
											</span>
										</div>
									</div>
								)}
								{/* Warning Messages */}
								{Object.keys(progressValues.warningMessages).length > 0 && (
									<div className="my-2">
										<Accordion type="single" collapsible defaultValue="warnings">
											<AccordionItem value="warnings">
												<AccordionTrigger className="text-destructive">
													Errors ({Object.keys(progressValues.warningMessages).length})
												</AccordionTrigger>
												<AccordionContent>
													<div className="flex flex-col space-y-2">
														{Object.entries(progressValues.warningMessages).map(
															([key, item]) => (
																<div
																	key={`warning-${key}-${Math.random()}`}
																	className="flex items-center text-destructive"
																>
																	<span
																		className="mr-1 h-4 w-4 relative group cursor-pointer"
																		onClick={() => {
																			// Retry Download if there is a posterFile, posterFileType, fileName and mediaItem
																			if (
																				item.posterFile &&
																				item.posterFileType &&
																				item.fileName &&
																				item.mediaItem
																			) {
																				log(
																					"INFO",
																					"Download Modal",
																					"Debug Info",
																					"Retrying Download for:",
																					{
																						posterFile: item.posterFile,
																						posterFileType:
																							item.posterFileType,
																						fileName: item.fileName,
																						mediaItem: item.mediaItem,
																					}
																				);
																				setProgressValues((prev) => {
																					// eslint-disable-next-line @typescript-eslint/no-unused-vars
																					const { [key]: _, ...rest } =
																						prev.warningMessages;
																					return {
																						...prev,
																						warningMessages: rest,
																					};
																				});
																				progressDownloadRef.current -=
																					progressIncrementRef.current;
																				updateProgressValue(
																					progressRef.current -
																						progressIncrementRef.current
																				);
																				// Retry the download
																				downloadPosterFileAndUpdateMediaServer(
																					item.posterFile,
																					item.posterFileType,
																					item.fileName,
																					item.mediaItem
																				);
																				setButtonTexts((prev) => ({
																					...prev,
																					download: `Download Again`,
																				}));
																			}
																		}}
																	>
																		<X className="absolute inset-0 h-4 w-4 transition-opacity duration-150 group-hover:opacity-0" />
																		<RefreshCcw className="absolute inset-0 h-4 w-4 opacity-0 transition-opacity duration-150 group-hover:opacity-100 text-yellow-500" />
																	</span>
																	<span>{item.message}</span>
																</div>
															)
														)}
													</div>
												</AccordionContent>
											</AccordionItem>
										</Accordion>
									</div>
								)}
								<DialogFooter>
									<div className="flex space-x-4">
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
													{buttonTexts.download.startsWith("Downloading") ||
													buttonTexts.download.startsWith("Adding") ? (
														<>
															<Loader className="mr-2 h-4 w-4 animate-spin" />
															{buttonTexts.download}
														</>
													) : (
														<>
															{buttonTexts.download === "Download" ? (
																<Download className="mr-2 h-4 w-4" />
															) : (
																<ListEnd className="mr-2 h-4 w-4" />
															)}
															{buttonTexts.download}
														</>
													)}
												</Button>
											)
										}

										{/* If the all items have no types selected and nothing is set to "Add to DB Only", show a button to reset the form */}
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
													setSelectedSizes({
														fileCount: 0,
														downloadSize: 0,
													});
													setProgressValues({
														value: 0,
														currentText: "",
														itemProgress: {},
														warningMessages: {},
													});
													setDuplicates({});
												}}
											>
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
