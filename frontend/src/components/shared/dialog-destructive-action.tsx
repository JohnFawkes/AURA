"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";

interface ConfirmDestructiveDialogActionProps {
	children: React.ReactNode; // Button label
	onConfirm: () => void | Promise<void>;
	title?: string;
	description?: string;
	confirmText?: string;
	cancelText?: string;
	variant?: "destructive" | "outline" | "default" | "ghost" | "link";
	disabled?: boolean;
	hidden?: boolean;
	className?: string;
}

export function ConfirmDestructiveDialogActionButton({
	children,
	onConfirm,
	title = "Are you sure?",
	description = "This action cannot be undone. Are you sure you want to continue?",
	confirmText = "Yes, Continue",
	cancelText = "Cancel",
	variant = "ghost",
	disabled,
	hidden,
	className,
}: ConfirmDestructiveDialogActionProps) {
	const [open, setOpen] = useState(false);
	const [loading, setLoading] = useState(false);

	const handleConfirm = async () => {
		setLoading(true);
		await onConfirm();
		setLoading(false);
		setOpen(false);
	};

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button
					variant={variant}
					disabled={disabled}
					hidden={hidden}
					onClick={() => setOpen(true)}
					className={className || "cursor-pointer hover:text-primary active:scale-95 hover:brightness-120"}
				>
					{children}
				</Button>
			</DialogTrigger>
			<DialogContent className="border border-red-500">
				<DialogHeader>
					<DialogTitle>{title}</DialogTitle>
					<DialogDescription>{description}</DialogDescription>
				</DialogHeader>
				<DialogFooter>
					<DialogClose asChild>
						<Button
							variant="outline"
							className="hover:text-primary active:scale-95 hover:brightness-120"
							disabled={loading}
						>
							{cancelText}
						</Button>
					</DialogClose>
					<Button
						variant="ghost"
						className="text-destructive border-1 shadow-none hover:text-red-500 cursor-pointer"
						onClick={handleConfirm}
						disabled={loading}
					>
						{confirmText}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
