import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export interface DownloadModalPopoverProps {
	type: "autodownload" | "future-updated-only" | "add-to-db-only";
}

const downloadModalPopoverHelpText = {
	autodownload:
		"Auto Download will check periodically for new updates to this set. This is helpful if you want to download and apply titlecard	updates from future updates to this set.",
	"future-updated-only":
		"Future Updates Only will not download anything right now. This is helpful if you have already downloaded the set and just want to future updates to be applied.",
	"add-to-db-only":
		"Add to Database Only will not download anything. It will only add the set to your database. This is helpful for movies that you have already processed and just want to add the set to your database.",
};

const DownloadModalPopover: React.FC<DownloadModalPopoverProps> = ({ type }) => {
	const helpText = downloadModalPopoverHelpText[type as keyof typeof downloadModalPopoverHelpText];
	return (
		<div className="ml-auto">
			<Popover modal={true}>
				<PopoverTrigger className="cursor-pointer">
					<span className="text-gray-500 dark:text-gray-400 cursor-pointer">?</span>
				</PopoverTrigger>
				<PopoverContent className="w-60">{helpText}</PopoverContent>
			</Popover>
		</div>
	);
};

export default DownloadModalPopover;
