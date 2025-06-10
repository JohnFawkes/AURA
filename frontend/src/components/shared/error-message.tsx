import { AlertCircle } from "lucide-react";

import { cn } from "@/lib/utils";

interface ErrorMessageProps {
	message: string;
	className?: string;
}

export function ErrorMessage({ message, className }: ErrorMessageProps) {
	if (!message) return null;

	return (
		<div className={cn("flex items-center justify-center mt-10", className)}>
			<div className="flex items-center gap-2 text-destructive">
				<AlertCircle className="h-4 w-4" />
				<p className="text-md font-medium">{message}</p>
			</div>
		</div>
	);
}
