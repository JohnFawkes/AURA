import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogOverlay,
	DialogPortal,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";
import { Download } from "lucide-react";
import Link from "next/link";
import { Button } from "./button";

const PosterSetModal: React.FC<{
	posterSet: PosterSet;
	mediaItem: MediaItem;
}> = ({ posterSet, mediaItem }) => {
	return (
		<Dialog>
			<DialogTrigger asChild>
				<button className="btn">
					<Download className="mr-2 h-4 w-4" />
				</button>
			</DialogTrigger>
			<DialogPortal>
				<DialogOverlay />
				<DialogContent className="sm:max-w-[425px]">
					<DialogHeader>
						<DialogTitle>
							Download Poster Set - {mediaItem.Title}
						</DialogTitle>
						<DialogDescription>
							<div className="flex flex-col ">
								<span className="text-sm text-muted-foreground">
									Set by: {posterSet.User.Name}
								</span>
								<Link
									href={`https://mediux.pro/sets/${posterSet.ID}`}
									className="hover:underline"
									target="_blank"
									rel="noopener noreferrer"
								>
									Set ID: {posterSet.ID}
								</Link>
							</div>
						</DialogDescription>
					</DialogHeader>
					<div className="flex justify-end">
						<DialogClose asChild>
							<div className="flex gap-2">
								<Button variant="destructive">Cancel</Button>
								<Button>Download</Button>
							</div>
						</DialogClose>
					</div>
				</DialogContent>
			</DialogPortal>
		</Dialog>
	);
};

export default PosterSetModal;
