import { RefreshCcw as RefreshIcon } from "lucide-react";

import { Button } from "@/components/ui/button";

import { cn } from "@/lib/cn";

interface RefreshButtonProps {
	onClick: () => void;
	text?: string;
	className?: string;
}

export function RefreshButton({ onClick, text = "Refresh", className }: RefreshButtonProps) {
	return (
		<Button
			variant="ghost"
			size="sm"
			className={cn(
				`fixed z-100 right-3 cursor-pointer
				bottom-10 sm:bottom-15
				rounded-full shadow-lg
				transition-all duration-300
				border-1 border-primary
				bg-primary/50 
				hover:text-black hover:!bg-primary
				`,
				className
			)}
			onClick={onClick}
			aria-label="refresh"
		>
			<RefreshIcon className="h-3 w-3 mr-1" />
			<span className="text-xs hidden sm:inline">{text}</span>
		</Button>
	);
}
