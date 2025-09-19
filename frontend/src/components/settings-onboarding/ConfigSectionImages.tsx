"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { postClearTempImagesFolder } from "@/services/settings-onboarding/api-images-actions";
import { toast } from "sonner";

import React, { useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigImages } from "@/types/config/config-app";

interface ConfigSectionImagesProps {
	value: AppConfigImages;
	editing: boolean;
	dirtyFields?: {
		CacheImages?: { Enabled?: boolean };
		SaveImageNextToContent?: { Enabled?: boolean };
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
				<Button
					variant="destructive"
					size="sm"
					hidden={editing}
					disabled={editing}
					onClick={() => {
						clearTempImagesFolder();
					}}
				>
					Clear Temp Images Folder
				</Button>
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
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-images-cache"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent side="top" align="end" sideOffset={6} className="w-64 text-xs leading-snug">
								<p>
									Store downloaded artwork locally to reduce external requests and speed repeat
									access.
								</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
			</div>

			{/* Save Images Next To Content */}
			<div
				className={cn(
					"flex items-center justify-between border rounded-md p-3 transition",
					"border-muted",
					dirtyFields.SaveImageNextToContent?.Enabled && "border-amber-500"
				)}
			>
				<Label className="mr-2">Save Images Next To Content</Label>
				<div className="flex items-center gap-2">
					<Switch
						disabled={!editing}
						checked={value.SaveImageNextToContent.Enabled}
						onCheckedChange={(v) => onChange("SaveImageNextToContent", "Enabled", v)}
					/>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-images-save-next-to-content"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-64 text-xs leading-snug"
							>
								<p>
									Write artwork files beside media items so external tools or servers can pick them up
									directly.
								</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
			</div>
		</Card>
	);
};
