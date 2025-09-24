"use client";

import { HelpCircle } from "lucide-react";

import React, { useEffect, useRef } from "react";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

import { cn } from "@/lib/cn";

import { AppConfigTMDB } from "@/types/config/config-app";

interface ConfigSectionTMDBProps {
	value: AppConfigTMDB;
	editing: boolean;
	dirtyFields?: Partial<Record<keyof AppConfigTMDB, boolean>>;
	onChange: <K extends keyof AppConfigTMDB>(field: K, value: AppConfigTMDB[K]) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigTMDB, string>>) => void;
	onTest?: (cfg: AppConfigTMDB) => void;
}

export const ConfigSectionTMDB: React.FC<ConfigSectionTMDBProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
	onTest,
}) => {
	const prevErrorsRef = useRef<string>("");

	const errors = React.useMemo<Partial<Record<keyof AppConfigTMDB, string>>>(() => {
		return {};
	}, []);

	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	return (
		<Card
			hidden={true} // Hide TMDB settings for now since it's not used yet
			className="p-5 space-y-1"
		>
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">TMDB</h2>
				<Button
					variant="outline"
					disabled={editing || !!errors.ApiKey}
					hidden={editing}
					onClick={() => onTest?.(value)}
				>
					Test Key
				</Button>
			</div>

			<div
				className={cn(
					"space-y-1 border rounded-md p-3 transition",
					errors.ApiKey ? "border-red-500" : dirtyFields.ApiKey ? "border-amber-500" : "border-muted"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>API Key</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-tmdb-api-key">
							<p className="mb-2">The TMDB API key used for metadata lookups (v3 key).</p>
							<p className="text-muted-foreground">
								Get one at:{" "}
								<a
									href="https://www.themoviedb.org/settings/api"
									target="_blank"
									rel="noopener noreferrer"
									className="underline"
								>
									https://www.themoviedb.org/settings/api
								</a>
							</p>
						</PopoverHelp>
					)}
				</div>
				<Input
					disabled={!editing}
					placeholder="TMDB API key"
					value={value.ApiKey}
					onChange={(e) => onChange("ApiKey", e.target.value)}
					aria-invalid={!!errors.ApiKey}
					onBlur={() => {
						if (!errors.ApiKey) onTest?.(value);
					}}
				/>
				{errors.ApiKey && <p className="text-xs text-red-500">{errors.ApiKey}</p>}
			</div>
		</Card>
	);
};
