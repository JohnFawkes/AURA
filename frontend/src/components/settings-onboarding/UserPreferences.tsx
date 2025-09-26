import { PopoverHelp } from "@/components/shared/popover-help";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";
import { useUserPreferencesStore } from "@/lib/stores/global-user-preferences";

import { DOWNLOAD_DEFAULT_TYPE_OPTIONS } from "@/types/ui-options";

export function UserPreferencesCard() {
	// Download Defaults from User Preferences Store
	const downloadDefaultTypes = useUserPreferencesStore((state) => state.downloadDefaults);
	const setDownloadDefaultTypes = useUserPreferencesStore((state) => state.setDownloadDefaults);
	const showOnlyDownloadDefaults = useUserPreferencesStore((state) => state.showOnlyDownloadDefaults);
	const setShowOnlyDownloadDefaults = useUserPreferencesStore((state) => state.setShowOnlyDownloadDefaults);

	return (
		<Card className="mt-4 p-5 space-y-1 border border-muted">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">User Preferences</h2>
			</div>
			<div className="border rounded-md p-3 mt-3 space-y-2">
				<div className="flex items-center justify-between">
					<Label>Download Defaults</Label>
					<PopoverHelp ariaLabel="help-default-image-types">
						<p className="mb-2">
							Select which image types you want auto-checked for each download. This will let you avoid
							unchecking them manually for each download.
						</p>
						<p className="text-muted-foreground">Click a badge to toggle it on or off.</p>
					</PopoverHelp>
				</div>
				<div className="flex flex-wrap gap-2 mt-3">
					{DOWNLOAD_DEFAULT_TYPE_OPTIONS.map((type) => (
						<Badge
							key={type}
							className={cn(
								"cursor-pointer text-sm px-3 py-1 font-normal transition",
								downloadDefaultTypes.includes(type)
									? "bg-primary text-primary-foreground active:scale-95 hover:brightness-120"
									: "bg-muted text-muted-foreground border hover:text-accent-foreground"
							)}
							variant={downloadDefaultTypes.includes(type) ? "default" : "outline"}
							onClick={() => {
								if (downloadDefaultTypes.includes(type)) {
									// Only allow removal if more than one type is selected
									if (downloadDefaultTypes.length > 1) {
										setDownloadDefaultTypes(downloadDefaultTypes.filter((t) => t !== type));
									}
								} else {
									setDownloadDefaultTypes([...downloadDefaultTypes, type]);
								}
							}}
							style={
								downloadDefaultTypes.includes(type) && downloadDefaultTypes.length === 1
									? { opacity: 0.5, pointerEvents: "none" }
									: undefined
							}
						>
							{type.charAt(0).toUpperCase() + type.slice(1).replace(/([A-Z])/g, " $1")}
						</Badge>
					))}
				</div>
				<div className="flex items-center justify-between mt-3">
					<div className="flex items-center gap-5">
						<Label>Only Show Download Defaults</Label>
						<Switch
							checked={showOnlyDownloadDefaults}
							onCheckedChange={() => setShowOnlyDownloadDefaults(!showOnlyDownloadDefaults)}
						/>
					</div>
					<PopoverHelp ariaLabel="help-filter-image-types">
						<p className="mb-2">
							If checked, only sets that contain at least one of the selected image types will be shown.
						</p>
						<p className="text-muted-foreground">
							This is global setting that will be applied to all media items and user sets. You can always
							change this setting here or in the Filters section of the Media Item Page. Section.
						</p>
					</PopoverHelp>
				</div>
				<div className="mt-2"></div>
			</div>
		</Card>
	);
}
