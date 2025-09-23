import { HelpCircle } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";
import { useUserPreferencesStore } from "@/lib/stores/global-user-preferences";

import { DEFAULT_IMAGE_TYPE_OPTIONS } from "@/types/ui-options";

export function UserPreferencesCard() {
	const defaultImageTypes = useUserPreferencesStore((state) => state.defaultImageTypes);
	const setDefaultImageTypes = useUserPreferencesStore((state) => state.setDefaultImageTypes);
	const showOnlyDefaultImages = useUserPreferencesStore((state) => state.showOnlyDefaultImages);
	const setShowOnlyDefaultImages = useUserPreferencesStore((state) => state.setShowOnlyDefaultImages);

	return (
		<Card className="mt-4 p-5 space-y-1 border border-muted">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">User Preferences</h2>
			</div>
			<div className="border rounded-md p-3 mt-3 space-y-2">
				<div className="flex items-center justify-between">
					<Label>Default Image Types</Label>
					<Popover>
						<PopoverTrigger asChild>
							<Button
								variant="outline"
								className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
								aria-label="help-default-image-types"
							>
								<HelpCircle className="h-4 w-4" />
							</Button>
						</PopoverTrigger>
						<PopoverContent side="top" align="end" sideOffset={6} className="w-64 text-xs leading-snug">
							<p className="mb-2">
								Select which image types you want autochecked for each download. This will let you avoid
								unchecking them manually for each download.
							</p>
							<p className="text-[10px] text-muted-foreground">Click a badge to toggle it on or off.</p>
						</PopoverContent>
					</Popover>
				</div>
				<div className="flex flex-wrap gap-2 mt-3">
					{DEFAULT_IMAGE_TYPE_OPTIONS.map((type) => (
						<Badge
							key={type}
							className={cn(
								"cursor-pointer text-sm px-3 py-1 font-normal transition",
								defaultImageTypes.includes(type)
									? "bg-primary text-primary-foreground"
									: "bg-muted text-muted-foreground border"
							)}
							variant={defaultImageTypes.includes(type) ? "default" : "outline"}
							onClick={() => {
								if (defaultImageTypes.includes(type)) {
									// Only allow removal if more than one type is selected
									if (defaultImageTypes.length > 1) {
										setDefaultImageTypes(defaultImageTypes.filter((t) => t !== type));
									}
								} else {
									setDefaultImageTypes([...defaultImageTypes, type]);
								}
							}}
							style={
								defaultImageTypes.includes(type) && defaultImageTypes.length === 1
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
						<Label>Only Show Default Image Types</Label>
						<Switch
							checked={showOnlyDefaultImages}
							onCheckedChange={() => setShowOnlyDefaultImages(!showOnlyDefaultImages)}
						/>
					</div>
					<Popover>
						<PopoverTrigger asChild>
							<Button
								variant="outline"
								className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
								aria-label="help-show-only-default-image-types"
							>
								<HelpCircle className="h-4 w-4" />
							</Button>
						</PopoverTrigger>
						<PopoverContent side="top" align="end" sideOffset={6} className="w-64 text-xs leading-snug">
							<p className="mb-2">
								When enabled, only poster sets that have at least one of your selected default image
								types will be shown.
							</p>
						</PopoverContent>
					</Popover>
				</div>
				<div className="mt-2"></div>
			</div>
		</Card>
	);
}
