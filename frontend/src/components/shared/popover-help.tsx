import { HelpCircle } from "lucide-react";

import { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

interface PopoverHelpProps {
	children: ReactNode;
	trigger?: ReactNode;
	ariaLabel: string;
	side?: "top" | "right" | "bottom" | "left";
	align?: "start" | "center" | "end";
	sideOffset?: number;
	className?: string;
}

export function PopoverHelp({
	children,
	trigger,
	ariaLabel,
	side = "bottom",
	align = "center",
	sideOffset = 8,
	className = "w-72 text-sm leading-snug",
}: PopoverHelpProps) {
	return (
		<Popover>
			<PopoverTrigger asChild>
				{trigger ? (
					trigger
				) : (
					<Button
						variant="outline"
						className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition hover:cursor-help active:scale-95 hover:brightness-120 p-0"
						aria-label={ariaLabel}
					>
						<HelpCircle className="w-4 h-4" />
					</Button>
				)}
			</PopoverTrigger>
			<PopoverContent side={side} align={align} sideOffset={sideOffset} className={className}>
				{children}
			</PopoverContent>
		</Popover>
	);
}
