"use client";

import { posterSetToFormItem } from "@/helper/download-modal/posterSetToFormItem";
import { formatDownloadSize } from "@/helper/formatDownloadSize";
import { postAddItemToDB } from "@/services/api.db";
import { fetchMediaServerItemContent, patchDownloadPosterFileAndUpdateMediaServer } from "@/services/api.mediaserver";
import { zodResolver } from "@hookform/resolvers/zod";
import { Download, Loader, TriangleAlert, X } from "lucide-react";
import { z } from "zod";

import { useEffect, useMemo, useRef, useState } from "react";
import { ControllerRenderProps, useForm, useWatch } from "react-hook-form";

import Link from "next/link";

import { AssetImage } from "@/components/shared/asset-image";
import DownloadModalPopover from "@/components/shared/download-modal-popover";
import { DownloadModalProgressItem } from "@/components/shared/download-modal-progress-item";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
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

import { log } from "@/lib/logger";

import { DBSavedItem } from "@/types/databaseSavedSet";
import { MediaItem } from "@/types/mediaItem";
import { PosterFile, PosterSet } from "@/types/posterSets";

export interface FormItemDisplay {
	MediaItemRatingKey: string;
	MediaItemTitle: string;
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
	// Number of Files Failed
	seasonPosterFailed?: number;
	titlecardFailed?: number;
};

// Main progress type
type DownloadProgress = {
	// Shared progress bar state
	value: number;
	color: string;

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
			z.object({
				types: z.array(z.enum(Object.keys(AssetTypes) as [AssetType, ...AssetType[]])),
				autodownload: z.boolean().optional(),
				futureUpdatesOnly: z.boolean().optional(),
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
	const [isMounted, setIsMounted] = useState(false);

	// Download Progress
	const [progressValues, setProgressValues] = useState<DownloadProgress>({
		value: 0,
		color: "",
		itemProgress: {},
		warningMessages: {},
	});
	// Add this with your other state declarations
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

	// Function - Reset Progress Values
	const resetProgressValues = () => {
		setProgressValues({
			value: 0,
			color: "",
			itemProgress: {},
			warningMessages: {},
		});
	};

	// Function - Close Modal
	const handleClose = () => {
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
		return items;
	}, [posterSets]);

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
						types: computeAssetTypes(item),
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
						types: computeAssetTypes(item),
						autodownload: item.Set.Type === "show" ? autoDownloadDefault : false,
						addToDBOnly: false,
						source: item.Set.Type === "movie" || item.Set.Type === "collection" ? item.Set.Type : undefined,
					};
					return acc;
				},
				{} as z.infer<typeof formSchema>["selectedOptionsByItem"]
			),
		});
	}, [formItems, form, autoDownloadDefault]);

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
		log("Props:", {
			setType,
			setTitle,
			setAuthor,
			setID,
			posterSets,
		});
		log("Form Items:", formItems);
		log("Form Values:", form);
		log("Watch Selected Types:", watchSelectedOptions);
		log("Progress Values:", progressValues);
		log("Duplicates:", duplicates);
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

		return (
			<FormItem key={`${field.name}-${assetType}`} className="flex flex-row items-start space-x-2 space-y-0">
				<FormControl>
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
			</FormItem>
		);
	};

	const renderFormItem = (item: FormItemDisplay) => {
		const isDuplicate = duplicates[item.MediaItemRatingKey];
		// Calculate disabled state
		const isDisabled = Boolean(
			isDuplicate && isDuplicate.selectedType && isDuplicate.selectedType !== item.Set.Type
		);

		return (
			<FormField
				key={`${item.MediaItemRatingKey}-${item.Set.Type}`}
				control={form.control}
				name={`selectedOptionsByItem.${item.MediaItemRatingKey}`}
				render={({ field }) => (
					<div className="rounded-md border p-4 rounded mb-4">
						<FormLabel className="text-md font-normal mb-4">
							{item.MediaItemTitle}
							{setType === "boxset" &&
								isDuplicate &&
								isDuplicate.selectedType !== "" &&
								isDuplicate.selectedType !== item.Set.Type && (
									<Popover modal={true}>
										<PopoverTrigger>
											<TriangleAlert className="h-4 w-4 text-yellow-500 cursor-help" />
										</PopoverTrigger>
										<PopoverContent className="w-60">
											<div className="text-sm text-yellow-500">
												This item is selected in the{" "}
												{isDuplicate.selectedType === "movie" ? "Movie Set" : "Collection Set"}.
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
		dbItem: DBSavedItem;
	} => {
		return {
			dbItem: {
				MediaItemID: mediaItem.RatingKey,
				MediaItem: mediaItem,
				PosterSetID: item.Set.ID,
				PosterSet: item.Set,
				LastDownloaded: "",
				SelectedTypes: options.types,
				AutoDownload: options.autodownload || false,
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
			if (response.status === "error") {
				updateItemProgress(mediaItem.RatingKey, posterFileType, `Failed to download ${fileName}`);
				throw new Error(response.error?.Message || "Unknown error");
			} else {
				updateItemProgress(mediaItem.RatingKey, posterFileType, `Finished downloading ${fileName}`);
			}
			updateProgressValue(progressRef.current + progressIncrementRef.current);
		} catch (error) {
			if (posterFileType === "seasonPoster" || posterFileType === "titlecard") {
				setProgressValues((prev) => ({
					...prev,
					progressColor: "yellow",
					// warningMessages: [
					// 	...prev.warningMessages,
					// 	`${fileName} - ${error instanceof Error ? error.message : "Unknown error"}`,
					// ],
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
					progressColor: "yellow",
					// warningMessages: [
					// 	...prev.warningMessages,
					// 	`${fileName} - ${error instanceof Error ? error.message : "Unknown error"}`,
					// ],
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

		try {
			setIsMounted(true);
			setButtonTexts({
				cancel: "Cancel",
				download: "Downloading...",
			});
			progressDownloadRef.current = 0;
			resetProgressValues();

			// Your download logic here
			log("Form submitted with data:", data);
			log("Progress Values:", progressValues);
			log("Selected Types:", { watchSelectedOptions });

			updateProgressValue(1); // Reset progress to 0 at the start

			// File Count + 1 for every item
			progressIncrementRef.current = 95 / (selectedSizes.fileCount + formItems.length);

			// Sort formItems by MediaItemTitle for consistent order
			const sortedFormItems = formItems.sort((a, b) => a.MediaItemTitle.localeCompare(b.MediaItemTitle));
			log("Sorted Form Items:", sortedFormItems);

			// Go through each item in the formItems
			for (const item of sortedFormItems) {
				log("Processing Item:", item);

				const selectedOptions = data.selectedOptionsByItem[item.MediaItemRatingKey];
				log("Selected Types for Item:", selectedOptions);

				// If no types are selected and not set to "Add to DB Only", skip this item
				if (selectedOptions.types.length === 0 && !selectedOptions.addToDBOnly) {
					log(`Skipping ${item.MediaItemTitle} - Nothing to do here.`);
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
					log(`No MediaItem found for ${item.MediaItemTitle}. Skipping.`);
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
						`Skipping ${item.MediaItemTitle} in ${item.SetID} - Duplicate type selected: ${duplicates[item.MediaItemRatingKey].selectedType}`
					);
					continue;
				}

				// Get the latest media item from the server
				const latestMediaItemResp = await fetchMediaServerItemContent(
					item.MediaItemRatingKey,
					currentMediaItem.LibraryTitle
				);
				if (latestMediaItemResp.status === "error") {
					setProgressValues((prev) => ({
						...prev,
						progressColor: "red",
						// warningMessages: [
						// 	...prev.warningMessages,
						// 	`Error fetching latest media item for ${item.MediaItemTitle}. Skipping.`,
						// ],
						warningMessages: {
							...prev.warningMessages,
							[item.MediaItemTitle]: {
								posterFile: null,
								posterFileType: null,
								fileName: null,
								mediaItem: null,
								message: `Error fetching latest media item for ${item.MediaItemTitle}: ${latestMediaItemResp.error}`,
							},
						},
					}));
					continue;
				}
				if (!latestMediaItemResp.data) {
					setProgressValues((prev) => ({
						...prev,
						progressColor: "red",
						// warningMessages: [
						// 	...prev.warningMessages,
						// 	`No media item found for ${item.MediaItemTitle}. Skipping.`,
						// ],
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
				const latestMediaItem = latestMediaItemResp.data;

				// If the item is set to "Add to DB Only", create the DB item and skip download
				const createdSavedItem = createDBItem(item, selectedOptions, latestMediaItem);
				if (selectedOptions.addToDBOnly) {
					updateItemProgress(item.MediaItemRatingKey, "addToDB", "Adding");
					log(`Adding ${item.MediaItemTitle} to DB only.`);
					log("DB Item Created:", createdSavedItem.dbItem);
					const addToDBResp = await postAddItemToDB(createdSavedItem.dbItem);
					if (addToDBResp.status === "error") {
						log(`Error adding ${item.MediaItemTitle} to DB:`, addToDBResp.error);
						setProgressValues((prev) => ({
							...prev,
							progressColor: "red",
							// warningMessages: [...prev.warningMessages, `Error adding ${item.MediaItemTitle} to DB.`],
							warningMessages: {
								...prev.warningMessages,
								[item.MediaItemTitle]: {
									posterFile: null,
									posterFileType: null,
									fileName: null,
									mediaItem: null,
									message: `Error adding ${item.MediaItemTitle} to DB: ${addToDBResp.error}`,
								},
							},
						}));
						updateItemProgress(item.MediaItemRatingKey, "addToDB", "Failed to add to DB");
					} else {
						log(`Successfully added ${item.MediaItemTitle} to DB.`);
						updateItemProgress(item.MediaItemRatingKey, "addToDB", "Added to DB");
					}
					updateProgressValue(progressRef.current + progressIncrementRef.current);
					updateItemProgress(item.MediaItemRatingKey, "addToDB", "Added to DB");
					continue;
				}

				const selectedTypes = selectedOptions.types.sort((a, b) => {
					const order = ["poster", "backdrop", "seasonPoster", "specialSeasonPoster", "titlecard"];
					return order.indexOf(a) - order.indexOf(b);
				});

				// If no types are selected, skip this item
				if (selectedTypes.length === 0) {
					log(`Skipping ${item.MediaItemTitle} - No types selected.`);
					continue;
				}

				log(`Selected Types for ${item.MediaItemTitle}:`, selectedTypes);

				for (const type of selectedTypes) {
					switch (type) {
						case "poster":
							if (item.Set.Poster) {
								log(`Downloading Poster for ${item.MediaItemTitle}`);
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
								log(`Downloading Backdrop for ${item.MediaItemTitle}`);
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
								log(`Downloading Season Posters for ${item.MediaItemTitle}`);
								const sortedSeasonPosters = item.Set.SeasonPosters.filter(
									(sp) => sp.Season?.Number !== 0
								).sort((a, b) => (a.Season?.Number || 0) - (b.Season?.Number || 0));
								for (const sp of sortedSeasonPosters) {
									const seasonExists = latestMediaItem.Series?.Seasons.some(
										(season) => season.SeasonNumber === sp.Season?.Number
									);
									if (!seasonExists) {
										log(
											`Skipping Season ${sp.Season?.Number} for ${item.MediaItemTitle} - Season does not exist in latest media item.`
										);
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
								log(`Downloading Special Season Posters for ${item.MediaItemTitle}`);
								const specialSeasonPosters = item.Set.SeasonPosters.filter(
									(sp) => sp.Season?.Number === 0
								);
								for (const sp of specialSeasonPosters) {
									const spExists = latestMediaItem.Series?.Seasons.some(
										(season) => season.SeasonNumber === 0
									);
									if (!spExists) {
										log(
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
								log(`Downloading Title Cards for ${item.MediaItemTitle}`);
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
												`Skipping Title Card for S${tc.Episode.SeasonNumber}E${tc.Episode.EpisodeNumber} - Episode does not exist in latest media item.`
											);
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
							log(`Unknown type ${type} for ${item.MediaItemTitle}. Skipping.`);
							continue;
					}
				}
				// Add the item to the database after downloading
				updateItemProgress(item.MediaItemRatingKey, "addToDB", "Adding");
				log(`Adding ${item.MediaItemTitle} to DB only.`);
				log("DB Item Created:", createdSavedItem.dbItem);
				const addToDBResp = await postAddItemToDB(createdSavedItem.dbItem);
				if (addToDBResp.status === "error") {
					log(`Error adding ${item.MediaItemTitle} to DB:`, addToDBResp.error);
					setProgressValues((prev) => ({
						...prev,
						progressColor: "red",
						// warningMessages: [...prev.warningMessages, `Error adding ${item.MediaItemTitle} to DB.`],
						warningMessages: {
							...prev.warningMessages,
							[item.MediaItemTitle]: {
								posterFile: null,
								posterFileType: null,
								fileName: null,
								mediaItem: null,
								message: `Error adding ${item.MediaItemTitle} to DB: ${addToDBResp.error}`,
							},
						},
					}));
					updateItemProgress(item.MediaItemRatingKey, "addToDB", "Failed to add to DB");
				} else {
					log(`Successfully added ${item.MediaItemTitle} to DB.`);
					updateItemProgress(item.MediaItemRatingKey, "addToDB", "Added to DB");
				}
				updateProgressValue(progressRef.current + progressIncrementRef.current);
				updateItemProgress(item.MediaItemRatingKey, "addToDB", "Added to DB");
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
			log("Download Error:", error);
			setProgressValues((prev) => ({
				...prev,
				progressColor: "red",
				// warningMessages: [...prev.warningMessages, "An error occurred while downloading."],
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
				<Download className="mr-2 h-5 w-5 sm:h-7 sm:w-7 cursor-pointer" />
			</DialogTrigger>

			<DialogPortal>
				<DialogOverlay />
				<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[500px] ">
					<DialogHeader>
						<DialogTitle onClick={LOG_VALUES}>{setTitle}</DialogTitle>
						<DialogDescription>{setAuthor}</DialogDescription>
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
										!watchSelectedOptions[item.MediaItemRatingKey].types.length &&
										!watchSelectedOptions[item.MediaItemRatingKey].addToDBOnly
								) && (
									<div className="text-sm text-destructive">
										No image types selected for download. Please select at least one image type to
										download.
									</div>
								)}

								{/* Total Size of Selected Types */}
								{
									// Only show if not all items are set to "Add to DB Only"
									selectedSizes.fileCount > 0 &&
									formItems.some(
										(item) => !watchSelectedOptions[item.MediaItemRatingKey].addToDBOnly
									) ? (
										<div>
											<div className="text-sm text-muted-foreground">
												Number of Images: {selectedSizes.fileCount}
											</div>
											<div className="text-sm text-muted-foreground">
												Total Download Size: {formatDownloadSize(selectedSizes.downloadSize)}
											</div>
										</div>
									) : (
										// If all items are set to "Add to DB Only", show a message
										formItems.some(
											(item) => watchSelectedOptions[item.MediaItemRatingKey].addToDBOnly
										) && (
											<div className="text-sm text-muted-foreground">
												Will add{" "}
												{
													formItems.filter(
														(item) =>
															watchSelectedOptions[item.MediaItemRatingKey].addToDBOnly ||
															watchSelectedOptions[item.MediaItemRatingKey].types.length >
																0
													).length
												}{" "}
												items to the database without downloading any images.
											</div>
										)
									)
								}

								{/* Progress Bar */}
								{progressValues.value > 0 && (
									<div className="w-full">
										<div className="flex items-center justify-between">
											<Progress
												value={progressValues.value}
												className={`flex-1 
												rounded-md ${progressValues.value < 100 ? "animate-pulse" : ""}
												${progressValues.color ? `[&>div]:bg-${progressValues.color}-500` : ""}`}
											/>
											<span className="ml-2 text-sm text-muted-foreground">
												{Math.round(progressValues.value)}%
											</span>
										</div>

										{Object.entries(progressValues.itemProgress)
											.sort(([a], [b]) => {
												const itemA = formItems.find((item) => item.MediaItemRatingKey === a);
												const itemB = formItems.find((item) => item.MediaItemRatingKey === b);
												return (itemA?.MediaItemTitle || "").localeCompare(
													itemB?.MediaItemTitle || ""
												);
											})
											.map(([itemKey, progress]) => {
												// Find the corresponding form item to get totals
												const formItem = formItems.find(
													(item) => item.MediaItemRatingKey === itemKey
												);

												return (
													<div key={itemKey} className="my-2">
														<Lead className="text-md text-muted-foreground mb-1">
															{formItem?.MediaItemTitle}
														</Lead>
														{progress.poster && (
															<DownloadModalProgressItem
																key={`${itemKey}-poster`}
																status={progress.poster}
																label="Poster"
															/>
														)}
														{progress.backdrop && (
															<DownloadModalProgressItem
																key={`${itemKey}-backdrop`}
																status={progress.backdrop}
																label="Backdrop"
															/>
														)}
														{progress.seasonPoster && (
															<DownloadModalProgressItem
																key={`${itemKey}-seasonPoster`}
																status={progress.seasonPoster}
																label="Season Poster"
																total={
																	formItem?.Set.SeasonPosters?.filter(
																		(sp) => sp.Season?.Number !== 0
																	).length
																}
																failed={progress.seasonPosterFailed || 0}
															/>
														)}
														{progress.specialSeasonPoster && (
															<DownloadModalProgressItem
																key={`${itemKey}-specialSeasonPoster`}
																status={progress.specialSeasonPoster}
																label="Special Season Poster"
															/>
														)}
														{progress.titlecard && (
															<DownloadModalProgressItem
																key={`${itemKey}-titlecard`}
																status={progress.titlecard}
																label="Title Card"
																total={
																	formItem?.Set.TitleCards?.filter(
																		(tc) =>
																			tc.Episode?.SeasonNumber !== undefined &&
																			tc.Episode?.EpisodeNumber !== undefined &&
																			tc.Episode.SeasonNumber !== 0
																	).length
																}
																failed={progress.titlecardFailed || 0}
															/>
														)}
														{progress.addToDB && (
															<DownloadModalProgressItem
																key={`${itemKey}-addToDB`}
																status={progress.addToDB}
																label="Add to DB"
															/>
														)}
													</div>
												);
											})}
									</div>
								)}
								{/* Warning Messages */}
								{Object.keys(progressValues.warningMessages).length > 0 && (
									<div className="my-2">
										<Accordion type="single" collapsible defaultValue="warnings">
											<AccordionItem value="warnings">
												<AccordionTrigger className="text-destructive">
													Failed Downloads (
													{Object.keys(progressValues.warningMessages).length})
												</AccordionTrigger>
												<AccordionContent>
													<div className="flex flex-col space-y-2">
														{Object.entries(progressValues.warningMessages).map(
															([key, item]) => (
																<div
																	key={`warning-${key}-${Math.random()}`}
																	className="flex items-center text-destructive"
																>
																	<X
																		className="mr-1 h-4 w-4"
																		onClick={() => {
																			// Retry Download if there is a posterFile, posterFileType, fileName and mediaItem
																			if (
																				item.posterFile &&
																				item.posterFileType &&
																				item.fileName &&
																				item.mediaItem
																			) {
																				log(
																					"Retrying Download for:",
																					item.posterFile,
																					item.posterFileType,
																					item.fileName,
																					item.mediaItem
																				);
																				// Remove this warningMessage
																				setProgressValues((prev) => {
																					// eslint-disable-next-line @typescript-eslint/no-unused-vars
																					const { [key]: _, ...rest } =
																						prev.warningMessages;
																					return {
																						...prev,
																						warningMessages: rest,
																					};
																				});
																				progressDownloadRef.current -= 1;
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
																	/>
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
												className="cursor-pointer hover:bg-destructive/90"
												variant="destructive"
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
													watchSelectedOptions[item.MediaItemRatingKey].types.length > 0 ||
													watchSelectedOptions[item.MediaItemRatingKey].addToDBOnly
											) && (
												<Button
													className="cursor-pointer hover:text-white"
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
													{buttonTexts.download.startsWith("Downloading") ? (
														<>
															<Loader className="mr-2 h-4 w-4 animate-spin" />
															{buttonTexts.download}
														</>
													) : (
														<>
															<Download className="mr-2 h-4 w-4" />
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
												className="cursor-pointer"
												variant="secondary"
												onClick={() => {
													form.reset();
													setSelectedSizes({
														fileCount: 0,
														downloadSize: 0,
													});
													setProgressValues({
														value: 0,
														color: "blue",
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
