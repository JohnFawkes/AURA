"use client";

import { useMemo } from "react";

import { ChangelogMarkdown } from "@/components/shared/changelog-markdown";
import { Button } from "@/components/ui/button";
import { Dialog, DialogClose, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

const funLabels = [
	"🎉 Awesome!",
	"🌟 Fantastic!",
	"🚀 Super!",
	"💡 Brilliant!",
	"✅ Excellent!",
	"✨ Wonderful!",
	"🏆 Splendid!",
	"🎊 Yay!",
	"👍 Right on!",
	"💃 Fab!",
	"👏 Great!",
	"😎 Nice!",
	"🍬 Sweet!",
	"🆒 Cool!",
	"🧩 Neat!",
	"🔔 All set!",
	"🏁 Let’s go!",
	"👌 Got it!",
	"🙏 Thanks!",
	"🥂 Cheers!",
];

export function ReleaseNotesDialog({
	open,
	onOpenChange,
	changelog,
}: {
	open: boolean;
	onOpenChange: (v: boolean) => void;
	changelog: string;
}) {
	// Pick a random label each time the dialog is rendered
	const randomLabel = useMemo(() => funLabels[Math.floor(Math.random() * funLabels.length)], []);
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-2xl border border-primary">
				<DialogHeader>
					<DialogTitle className="text-2xl font-bold">What's New?</DialogTitle>
				</DialogHeader>
				<div className="max-h-[60vh] overflow-y-auto pr-2">
					<ChangelogMarkdown>{changelog}</ChangelogMarkdown>
				</div>
				<DialogClose asChild>
					<Button
						variant={"outline"}
						className="cursor-pointer hover:text-primary hover:brightness-120 transition-colors"
					>
						{randomLabel}
					</Button>
				</DialogClose>
			</DialogContent>
		</Dialog>
	);
}
