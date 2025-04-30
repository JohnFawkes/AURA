"use client";
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
import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";
import { Download } from "lucide-react";
import Link from "next/link";
import { Button } from "./button";
import { useEffect, useMemo, useState } from "react";
import { Checkbox } from "./checkbox";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { log } from "@/lib/logger";
import { ClientMessage } from "@/types/clientMessage";
import { postSendSetToAPI } from "@/services/api.mediaserver";
import { Progress } from "./progress";

const PosterSetModal: React.FC<{
	posterSet: PosterSet;
	mediaItem: MediaItem;
}> = ({ posterSet, mediaItem }) => {
	// State for modal
	const [modalErrorMessage, setModalErrorMessage] = useState<string | null>(
		null
	);
	const [cancelButtonText, setCancelButtonText] = useState("Cancel");

	// Tracking selected checkboxes for what to download
	const [totalSelectedSize, setTotalSelectedSize] = useState("0 B");

	const [autoDownload, setAutoDownload] = useState(false);
	const handleAutoDownloadChange = () => {
		setAutoDownload((prev) => !prev);
	};

	// Download Progress
	const [progressColor, setProgressColor] = useState("default");
	const [progressValue, setProgressValue] = useState(0);
	const [progressText, setProgressText] = useState("");
	const [progressNextStep, setProgressNextStep] = useState("");
	const [progressWarningMessages, setProgressWarningMessages] = useState<
		{ fileName: string; message: string }[]
	>([]);

	// Handle when closing the modal
	const handleClose = () => {
		setModalErrorMessage(null);
		setCancelButtonText("Cancel");
		setAutoDownload(false);
		setProgressValue(0);
		setProgressText("");
		setProgressNextStep("");
		setProgressWarningMessages([]);
		form.reset();
	};

	// Function to format file size
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return "0 B";
		const k = 1024;
		const sizes = ["B", "KB", "MB", "GB", "TB"];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
	}

	const formSchema = z.object({
		items: z.array(z.string()).refine((value) => value.length > 0, {
			message: "You must select at least one asset type.",
		}),
	});

	const assetOptions = useMemo(() => {
		const typeSet = new Set(posterSet.Files.map((file) => file.Type));
		const assetLabels: Record<string, string> = {
			poster: "Poster",
			backdrop: "Backdrop",
			seasonPoster: "Season Poster",
			titlecard: "Titlecard",
		};
		return Object.entries(assetLabels)
			.filter(([key]) => typeSet.has(key))
			.map(([key, label]) => ({ id: key, label }));
	}, [posterSet]);

	const form = useForm<z.infer<typeof formSchema>>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			items: [],
		},
	});

	const items = form.watch("items");

	useEffect(() => {
		const selectedTypes = items || [];
		log("Poster Set Modal - Selected Types:", selectedTypes);

		const matchingFiles = posterSet.Files.filter((file) =>
			selectedTypes.includes(file.Type)
		);
		log("Poster Set Modal - Matching Files:", matchingFiles);

		const totalSizeBytes = matchingFiles.reduce(
			(acc, file) => acc + (parseInt(file.FileSize || "0") || 0),
			0
		);
		log(
			"Poster Set Modal - Total Size in Bytes:",
			totalSizeBytes,
			"Formatted Size:",
			formatFileSize(totalSizeBytes)
		);
		setTotalSelectedSize(formatFileSize(totalSizeBytes));
	}, [items, posterSet.Files]);

	// Function to handle form submission
	const onSubmit = async (data: z.infer<typeof formSchema>) => {
		log("Poster Set Modal - Form Submitted:", data);

		setModalErrorMessage(null);
		setProgressValue(0);
		setProgressText("");
		setProgressNextStep("");
		setProgressWarningMessages([]);
		setCancelButtonText("Cancel");

		try {
			const clientMessage: ClientMessage = {
				Set: posterSet,
				SelectedTypes: data.items,
				MediaItem: mediaItem,
				AutoDownload: autoDownload,
			};
			log("Poster Set Modal - Client Message:", clientMessage);

			const response = await postSendSetToAPI(clientMessage);
			if (!response || response.status !== "success") {
				throw new Error("Failed to start the task.");
			}

			setProgressColor("primary");
			setProgressValue(1);
			setProgressText("Task started successfully.");
			setProgressNextStep("Waiting for updates...");

			// Create a new SSE connection to the backend server
			const eventSource = new EventSource(
				`/api/mediaserver/update/set/${mediaItem.RatingKey}`
			);

			eventSource.onmessage = (event) => {
				try {
					// Parse the incoming data
					const data = JSON.parse(event.data);
					log("Poster Set Modal - SSE Data:", data);

					// Update progress bar and text
					if (data.response.status === "success") {
						setProgressValue(data.progress.value);
						setProgressText(data.progress.text);
						setProgressNextStep(data.progress.nextStep);
					} else if (data.response.status === "warning") {
						setProgressWarningMessages((prev) => [
							...prev,
							{
								fileName: data.progress.text,
								message: data.progress.nextStep,
							},
						]);
					} else if (data.response.status === "error") {
						throw new Error(
							data.response.message || "Unknown error"
						);
					} else if (data.response.status === "complete") {
						setProgressValue(100);
						if (progressWarningMessages.length > 0) {
							setProgressText(
								"Task completed with warnings. Check the logs for more details."
							);
						} else {
							setProgressText(data.response.message);
						}
						setProgressNextStep("");
						setCancelButtonText("Close");
						eventSource.close();
					}
				} catch (error) {
					log("Poster Set Modal - SSE Error:", error);
					setProgressColor("error");
					setModalErrorMessage(
						"An error occurred while processing updates. Check the logs for more details."
					);
					eventSource.close();
				}
			};

			eventSource.onerror = () => {
				eventSource.close();
				log("Poster Set Modal - Event Source connection error");
				setProgressColor("error");
				setModalErrorMessage(
					"An error occurred while connecting to the server. Check the logs for more details."
				);
			};
		} catch (error) {
			log("Poster Set Modal - Download Error:", error);
			setProgressColor("error");
			setModalErrorMessage(
				"An error occurred while processing the task. Check the logs for more details."
			);
			setProgressNextStep(
				error instanceof Error
					? error.message
					: "An unknown error occurred."
			);
		}
	};

	return (
		<Dialog>
			<DialogTrigger asChild>
				<button className="btn">
					<Download className="mr-2 h-4 w-4" />
				</button>
			</DialogTrigger>
			<DialogPortal>
				<DialogOverlay />
				<DialogContent className="sm:max-w-[425px]">
					<DialogHeader>
						<DialogTitle>
							Download Poster Set - {mediaItem.Title}
						</DialogTitle>
						<DialogDescription>
							<div className="flex flex-col ">
								<span className="text-sm text-muted-foreground">
									Set by: {posterSet.User.Name}
								</span>
								<Link
									href={`https://mediux.pro/sets/${posterSet.ID}`}
									className="hover:underline"
									target="_blank"
									rel="noopener noreferrer"
								>
									Set ID: {posterSet.ID}
								</Link>
							</div>
						</DialogDescription>
					</DialogHeader>

					<Form {...form}>
						<form
							onSubmit={form.handleSubmit(onSubmit)}
							className="space-y-4"
						>
							<FormField
								control={form.control}
								name="items"
								render={() => (
									<FormItem>
										<FormLabel className="text-base">
											Select Assets to Download
										</FormLabel>
										{assetOptions.map((item) => (
											<FormField
												key={item.id}
												control={form.control}
												name="items"
												render={({ field }) => (
													<FormItem className="flex flex-row items-start space-x-3 space-y-0">
														<FormControl>
															<Checkbox
																checked={field.value?.includes(
																	item.id
																)}
																onCheckedChange={(
																	checked
																) => {
																	return checked
																		? field.onChange(
																				[
																					...field.value,
																					item.id,
																				]
																		  )
																		: field.onChange(
																				field.value?.filter(
																					(
																						v
																					) =>
																						v !==
																						item.id
																				)
																		  );
																}}
															/>
														</FormControl>
														<FormLabel className="text-sm font-normal">
															{item.label}
														</FormLabel>
													</FormItem>
												)}
											/>
										))}
										<FormMessage />
									</FormItem>
								)}
							/>

							{/* Auto Download Check Box */}
							<FormItem className="space-y-1">
								<div className="flex items-center space-x-5">
									<FormControl>
										<Checkbox
											checked={autoDownload}
											onCheckedChange={
												handleAutoDownloadChange
											}
										/>
									</FormControl>
									<FormLabel className="text-sm font-normal">
										Auto Download
									</FormLabel>
								</div>
								<FormDescription className="text-xs text-muted-foreground">
									This will automatically download any new
									files added to this set.
								</FormDescription>
							</FormItem>

							{/* Total Size of Selected Types */}
							<div className="text-sm text-muted-foreground">
								Total Size: {totalSelectedSize}
							</div>

							{/* Progress Bar */}
							{progressValue > 0 && (
								<div className="w-full">
									<div className="flex items-center justify-between">
										<Progress
											value={progressValue}
											color={progressColor}
											className="flex-1"
										/>
										<span className="ml-2 text-sm text-muted-foreground">
											{progressValue}%
										</span>
									</div>
									<div className="flex justify-between text-sm text-muted-foreground ">
										<span>{progressText}</span>
									</div>
									<div className="text-sm text-muted-foreground">
										{progressNextStep}
									</div>
								</div>
							)}

							{/* Warning Messages */}
							{progressWarningMessages.length > 0 && (
								<div className="mt-2 space-y-4">
									<div className="text-sm text-destructive">
										Warnings:
									</div>
									{Object.entries(
										progressWarningMessages.reduce(
											(acc, warning) => {
												if (!acc[warning.message]) {
													acc[warning.message] = [];
												}
												acc[warning.message].push(
													warning.fileName
												);
												return acc;
											},
											{} as Record<string, string[]>
										)
									).map(([message, files], index) => (
										<div
											key={index}
											className="border border-destructive/50 rounded-md bg-destructive/10 p-4"
										>
											<div className="text-sm font-semibold text-destructive mb-2">
												{message}
											</div>
											<ul className="list-disc pl-5 text-sm text-muted-foreground">
												{files.map(
													(file, fileIndex) => (
														<li key={fileIndex}>
															{file}
														</li>
													)
												)}
											</ul>
										</div>
									))}
								</div>
							)}

							{/* Error Message */}
							{modalErrorMessage && (
								<div className="text-sm text-destructive">
									{modalErrorMessage}
								</div>
							)}

							<DialogFooter>
								<div className="flex space-x-4">
									{/* Cancel button to close the modal */}
									<DialogClose asChild>
										<Button
											className=""
											variant="destructive"
											onClick={() => {
												handleClose();
											}}
										>
											{cancelButtonText}
										</Button>
									</DialogClose>

									{/* Download button to display download info */}
									<Button
										className=""
										onClick={() => {
											onSubmit(form.getValues());
										}}
									>
										Download
									</Button>
								</div>
							</DialogFooter>
						</form>
					</Form>
				</DialogContent>
			</DialogPortal>
		</Dialog>
	);
};

export default PosterSetModal;
