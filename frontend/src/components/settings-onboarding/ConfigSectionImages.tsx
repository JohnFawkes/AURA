"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { postClearTempImagesFolder } from "@/services/settings-onboarding/api-images-actions";
import { toast } from "sonner";

import React, { useEffect, useRef } from "react";

import { ConfirmDestructiveDialogActionButton } from "@/components/shared/dialog-destructive-action";
import { PopoverHelp } from "@/components/shared/popover-help";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigImages } from "@/types/config/config-app";

const EPISODE_NAMING_CONVENTION_OPTIONS = ["match", "static"];

interface ConfigSectionImagesProps {
	value: AppConfigImages;
	editing: boolean;
	dirtyFields?: {
		CacheImages?: { Enabled?: boolean };
		SaveImagesLocally?: {
			Enabled?: boolean;
			Path?: boolean;
			EpisodeNamingConvention?: boolean;
		};
	};
	onChange: <K extends keyof AppConfigImages, F extends keyof AppConfigImages[K]>(
		group: K,
		field: F,
		value: AppConfigImages[K][F]
	) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigImages, string>>) => void;
	mediaServerType?: string;
}

export const ConfigSectionImages: React.FC<ConfigSectionImagesProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
	mediaServerType,
}) => {
	const prevErrorsRef = useRef<string>("{}");

	const clearTempImagesFolder = async () => {
		try {
			const response = await postClearTempImagesFolder();
			if (response.status === "error") {
				toast.error(response.error?.message || "Failed to clear temp images folder");
				return;
			}
			toast.success(response.data || "Temp images folder cleared successfully");
		} catch (error) {
			const errorResponse = ReturnErrorMessage<void>(error);
			toast.error(errorResponse.error?.message || "An unexpected error occurred");
		}
	};

	const errors = React.useMemo<Partial<Record<keyof AppConfigImages, string>>>(() => {
		const errs: Partial<Record<keyof AppConfigImages, string>> = {};

		// If Media Server Type is Plex, validate SaveImagesLocally.EpisodeNamingConvention
		if (mediaServerType && mediaServerType === "Plex" && value.SaveImagesLocally.Enabled) {
			if (!value.SaveImagesLocally.EpisodeNamingConvention) {
				errs.SaveImagesLocally = "Episode naming convention is required.";
			} else {
				if (!EPISODE_NAMING_CONVENTION_OPTIONS.includes(value.SaveImagesLocally.EpisodeNamingConvention)) {
					errs.SaveImagesLocally = `Episode naming convention must be one of: ${EPISODE_NAMING_CONVENTION_OPTIONS.join(", ")}.`;
				}
			}
		}

		return errs;
	}, [mediaServerType, value.SaveImagesLocally.Enabled, value.SaveImagesLocally.EpisodeNamingConvention]);

	// Emit errors upward
	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	return (
		<Card className="p-5">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold text-blue-500">Images</h2>
				<ConfirmDestructiveDialogActionButton
					hidden={editing}
					onConfirm={clearTempImagesFolder}
					title="Clear Temp Images Folder?"
					description="This will permanently delete all temporary images. Are you sure you want to continue?"
					confirmText="Yes, Clear Images"
					cancelText="Cancel"
					className="text-destructive border-1 shadow-none hover:text-red-500 cursor-pointer"
					variant="ghost"
				>
					Clear Temp Images
				</ConfirmDestructiveDialogActionButton>
			</div>
			{/* Cache Images */}
			<div
				className={cn(
					"flex items-center justify-between border rounded-md p-3 transition",
					"border-muted",
					dirtyFields.CacheImages?.Enabled && "border-amber-500"
				)}
			>
				<Label className="mr-2">Cache Images</Label>
				<div className="flex items-center gap-2">
					<Switch
						disabled={!editing}
						checked={value.CacheImages.Enabled}
						onCheckedChange={(v) => onChange("CacheImages", "Enabled", v)}
					/>
					{editing && (
						<PopoverHelp ariaLabel="help-images-cache">
							<p>Store downloaded artwork locally to reduce external requests and speed repeat access.</p>
						</PopoverHelp>
					)}
				</div>
			</div>

			{/* Save Images Locally */}
			{mediaServerType === "Plex" && (
				<div
					className={cn(
						"border rounded-md p-3 transition",
						"border-muted",
						dirtyFields.SaveImagesLocally?.Enabled && "border-amber-500"
					)}
				>
					<div className="flex items-center justify-between mb-2">
						<Label className="mr-2">Save Images Locally</Label>
						<div className="flex items-center gap-2">
							<Switch
								disabled={!editing}
								checked={!!value.SaveImagesLocally.Enabled}
								onCheckedChange={(v) => onChange("SaveImagesLocally", "Enabled", v)}
							/>
							{editing && (
								<PopoverHelp ariaLabel="help-images-save-next-to-content">
									<p>
										Save images to a local folder on the server. This is useful for not relying on
										your Media Server database. Make sure the path is accessible by the Aura server.
									</p>
								</PopoverHelp>
							)}
						</div>
					</div>

					{value.SaveImagesLocally.Enabled && (
						<div
							className={cn(
								"mt-2",
								dirtyFields.SaveImagesLocally?.Enabled && "border border-amber-500 rounded-md p-2"
							)}
						>
							<div className="flex items-center justify-between mb-2">
								<Label className="mr-2">Path</Label>
								{editing && (
									<PopoverHelp ariaLabel="help-images-save-path">
										<p>
											Enter the local folder path where images should be saved. This must be
											accessible by the Aura server. Leave this blank if you want to save images
											next to the content.
										</p>
									</PopoverHelp>
								)}
							</div>
							<Input
								type="text"
								disabled={!editing}
								value={value.SaveImagesLocally.Path || ""}
								onChange={(e) => onChange("SaveImagesLocally", "Path", e.target.value)}
								className={cn(
									"w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50 transition",
									dirtyFields.SaveImagesLocally?.Path && "border-amber-500"
								)}
								placeholder="/path/to/images"
							/>
						</div>
					)}

					{mediaServerType === "Plex" && value.SaveImagesLocally.Enabled && (
						<div className={cn("space-y-1 mt-4")}>
							<div className="flex items-center justify-between">
								<Label>Episode Naming Convention</Label>
								{editing && (
									<PopoverHelp ariaLabel="help-media-server-episode-naming-convention">
										<div className="space-y-3">
											<div>
												<p className="font-medium mb-1">Episode Naming Convention</p>
												<p className="text-muted-foreground">
													How Plex episode files are named.
												</p>
											</div>
											<ul className="space-y-1">
												<li className="flex items-center gap-2">
													<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono ">
														match
													</span>
													<span>Some Episode Title S01E01.jpg</span>
												</li>
												<li className="flex items-center gap-2">
													<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono">
														static
													</span>
													<span>S01E01.jpg</span>
												</li>
											</ul>
											<p className="text-muted-foreground">Used for file naming logic.</p>
										</div>
									</PopoverHelp>
								)}
							</div>
							<Select
								disabled={!editing}
								value={value.SaveImagesLocally.EpisodeNamingConvention || ""}
								onValueChange={(v) => onChange("SaveImagesLocally", "EpisodeNamingConvention", v)}
							>
								<SelectTrigger
									id="media-server-episode-naming-convention-trigger"
									className={cn(
										"w-full",
										dirtyFields.SaveImagesLocally?.EpisodeNamingConvention && "border-amber-500"
									)}
								>
									<SelectValue placeholder="Select convention..." />
								</SelectTrigger>
								<SelectContent>
									{EPISODE_NAMING_CONVENTION_OPTIONS.map((o) => (
										<SelectItem key={o} value={o}>
											{o}
										</SelectItem>
									))}
									<SelectScrollUpButton />
									<SelectScrollDownButton />
								</SelectContent>
							</Select>
						</div>
					)}
				</div>
			)}
		</Card>
	);
};
