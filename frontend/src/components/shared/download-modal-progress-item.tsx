import { Check, LoaderIcon, TriangleAlert, X } from "lucide-react";

export interface ProgressItemProps {
	status: string;
	label: string;
	total?: number | null;
	failed?: number | null;
}

export const DownloadModalProgressItem = ({ status, label, total, failed }: ProgressItemProps) => {
	// Status checks
	const isFinished = status.startsWith("Finished") || status.startsWith("Added");
	const isFailed = status.startsWith("Failed");
	const isInProgress = !isFinished && !isFailed;

	// Determine icon and color
	const getIcon = () => {
		if (isFinished) return <Check className="mr-1 h-4 w-4" />;
		if (isInProgress) return <LoaderIcon className="mr-1 h-4 w-4 animate-spin" />;

		if (!total) return <X className="mr-1 h-4 w-4 text-destructive" />;

		if (failed && failed !== 0 && failed === total) {
			return <X className="mr-1 h-4 w-4 text-destructive" />;
		}

		return <TriangleAlert className={"mr-1 h-4 w-4 text-yellow-500"} />;
	};

	return (
		<div className="flex justify-between text-sm text-muted-foreground">
			<div className="flex items-center">
				{getIcon()}
				{isInProgress && <span>{status}</span>}
				{isFinished && <span className="text-green-500">{label}</span>}
				{isFailed && !total && <span className="text-destructive">{label}</span>}
				{isFailed && total && failed && total === failed && <span className="text-destructive">{label}</span>}
				{isFailed && total && failed && total !== failed && (
					<span className="text-yellow-500">
						{" "}
						{label} ({failed}/{total}){" "}
					</span>
				)}
			</div>
		</div>
	);
};
