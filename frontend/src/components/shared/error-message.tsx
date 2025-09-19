"use client";

import { AlertCircle, AlertTriangle, ChevronDown } from "lucide-react";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { H3, H4, Lead } from "@/components/ui/typography";

import { cn } from "@/lib/cn";

import { APIResponse } from "@/types/api/api-response";

interface ErrorMessageProps<T> {
	error: APIResponse<T> | null;
	className?: string;
	isWarning?: boolean;
}

export function ErrorMessage<T>({ error, className, isWarning }: ErrorMessageProps<T>) {
	const [isExpanded, setIsExpanded] = useState(false);

	if (!error?.error) return null;

	const textColorClass = isWarning ? "text-yellow-500" : "text-destructive";
	const backgroundColorClass = isWarning ? "bg-yellow-500/10" : "bg-destructive/10";

	// Helper for pretty JSON
	const getPrettyDetails = () => {
		return JSON.stringify(
			Object.fromEntries(
				error?.error && typeof error.error.Details === "object" && error.error.Details !== null
					? Object.entries(error.error.Details).map(([key, value]) => {
							if (
								typeof value === "string" &&
								(value.trim().startsWith("{") || value.trim().startsWith("["))
							) {
								try {
									return [key, JSON.parse(value)];
								} catch {
									return [key, value];
								}
							}
							return [key, value];
						})
					: []
			),
			null,
			2
		);
	};

	return (
		<div className={cn("flex flex-col items-center justify-center mt-10 w-full max-w-md mx-auto", className)}>
			<div className={cn("w-full rounded-lg p-4 text-center", backgroundColorClass)}>
				<div className="flex items-center justify-center gap-2 mb-1">
					{isWarning ? (
						<AlertTriangle className="h-5 w-5 text-yellow-500" />
					) : (
						<AlertCircle className="h-5 w-5 text-destructive" />
					)}
					<H3 className={cn("text-lg", textColorClass)}>{isWarning ? "Warning" : "Error"}</H3>
				</div>

				<H4 className={cn("text-md", textColorClass)}>{error.error.Message}</H4>

				<Lead className="text-sm text-muted-foreground mt-2">{error.error.HelpText}</Lead>

				{(!!error.error.Details ||
					(!!error.error.Function && error.error.Function !== "" && error.error.Function !== "Unknown") ||
					(!!error.error.LineNumber && error.error.LineNumber !== 0)) && (
					<Button
						variant="outline"
						onClick={() => setIsExpanded(!isExpanded)}
						className="flex items-center gap-1 mx-auto mt-3 text-xs text-muted-foreground/80 hover:text-muted-foreground transition-colors"
					>
						<ChevronDown
							className={cn("h-4 w-4 transition-transform duration-200", isExpanded ? "rotate-180" : "")}
						/>
						{isExpanded ? "Hide details" : "Show details"}
					</Button>
				)}

				{isExpanded && (
					<div className="mt-3 pt-3 border-t border-border/50">
						<div className="text-xs text-left text-muted-foreground/80">
							{error.error.Function &&
								error.error.Function !== "" &&
								error.error.Function !== "Unknown" && (
									<Lead className="text-sm text-muted-foreground mb-1 break-words whitespace-pre-line">
										Function: {error.error.Function}
									</Lead>
								)}

							{error.error.LineNumber && error.error.LineNumber !== 0 && (
								<Lead className="text-sm text-muted-foreground mb-1">
									Line Number: {error.error.LineNumber}
								</Lead>
							)}

							{error.elapsed && error.elapsed !== "0" && (
								<Lead className="text-sm text-muted-foreground mb-1">
									Elapsed Time: {error.elapsed}
								</Lead>
							)}

							{error.error.Details && (
								<>
									<Lead className="text-sm text-muted-foreground">Details: </Lead>
									{typeof error.error.Details === "string" ? (
										<pre className="bg-muted/30 rounded p-2 mt-1 whitespace-pre-wrap break-words overflow-x-auto max-h-64 text-xs">
											{error.error.Details}
										</pre>
									) : (
										<>
											<Textarea
												rows={Math.max(6, getPrettyDetails().split("\n").length)}
												className="mt-2 w-full resize-none font-mono bg-muted/30"
												value={getPrettyDetails()}
												readOnly
											/>
										</>
									)}
								</>
							)}
						</div>
						<Button
							variant="secondary"
							className="mt-3 w-full"
							onClick={() => {
								window.location.href = "/logs";
							}}
						>
							Go to Logs
						</Button>
					</div>
				)}
			</div>
		</div>
	);
}
