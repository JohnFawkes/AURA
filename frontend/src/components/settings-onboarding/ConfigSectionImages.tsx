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
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigImages } from "@/types/config/config-app";

interface ConfigSectionImagesProps {
	value: AppConfigImages;
	editing: boolean;
	dirtyFields?: {
		CacheImages?: { Enabled?: boolean };
		SaveImageLocally?: { Enabled?: boolean; Path?: boolean };
	};
	onChange: <K extends keyof AppConfigImages, F extends keyof AppConfigImages[K]>(
		group: K,
		field: F,
		value: AppConfigImages[K][F]
	) => void;
	errorsUpdate?: (errors: Record<string, string>) => void;
}

export const ConfigSectionImages: React.FC<ConfigSectionImagesProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("{}");

	useEffect(() => {
		const errors: Record<string, string> = {};
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errorsUpdate]);

	const clearTempImagesFolder = async () => {
		try {
			const response = await postClearTempImagesFolder();

			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to clear temp images folder");
				return;
			}

			toast.success(response.data || "Temp images folder cleared successfully");
		} catch (error) {
			const errorResponse = ReturnErrorMessage<void>(error);
			toast.error(errorResponse.error?.Message || "An unexpected error occurred");
		}
	};

	return (
		<Card className="p-5 space-y-1">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Images</h2>
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
			<div
				className={cn(
					"border rounded-md p-3 transition mb-4",
					"border-muted",
					dirtyFields.SaveImageLocally?.Enabled && "border-amber-500"
				)}
			>
				<div className="flex items-center justify-between mb-2">
					<Label className="mr-2">Save Images Locally</Label>
					<div className="flex items-center gap-2">
						<Switch
							disabled={!editing}
							checked={value.SaveImageLocally.Enabled}
							onCheckedChange={(v) => onChange("SaveImageLocally", "Enabled", v)}
						/>
						{editing && (
							<PopoverHelp ariaLabel="help-images-save-next-to-content">
								<p>
									Save images to a local folder on the server. This is useful for not relying on your
									Media Server database. Make sure the path is accessible by the Aura server.
								</p>
							</PopoverHelp>
						)}
					</div>
				</div>

				{value.SaveImageLocally.Enabled && (
					<div
						className={cn(
							"",
							dirtyFields.SaveImageLocally?.Path && "border border-amber-500 rounded-md p-2"
						)}
					>
						<div className="flex items-center justify-between mb-2">
							<Label className="mr-2">Path</Label>
							{editing && (
								<PopoverHelp ariaLabel="help-images-save-path">
									<p>
										Enter the local folder path where images should be saved. This must be
										accessible by the Aura server. Leave this blank if you want to save images next
										to the content.
									</p>
								</PopoverHelp>
							)}
						</div>
						<Input
							type="text"
							disabled={!editing}
							value={value.SaveImageLocally.Path}
							onChange={(e) => onChange("SaveImageLocally", "Path", e.target.value)}
							className={cn(
								"w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50 transition",
								dirtyFields.SaveImageLocally?.Path && "border-amber-500"
							)}
							placeholder="/path/to/images"
						/>
					</div>
				)}
			</div>
		</Card>
	);
};
