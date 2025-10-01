"use client";

import { useEffect, useState } from "react";

import { ChangelogMarkdown } from "@/components/shared/changelog-markdown";

export default function Changelog() {
	const [content, setContent] = useState("");

	useEffect(() => {
		fetch("/CHANGELOG.md")
			.then((res) => res.text())
			.then(setContent);
	}, []);

	return (
		<div className="min-h-screen py-8 px-4 flex justify-center">
			<div className="w-full max-w-2xl rounded-lg shadow-md p-6">
				<h1 className="text-3xl font-bold mb-6 text-center">Change Log</h1>
				<ChangelogMarkdown>{content}</ChangelogMarkdown>
			</div>
		</div>
	);
}
