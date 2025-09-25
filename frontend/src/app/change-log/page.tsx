"use client";

import { useEffect, useState } from "react";
import { FaGithub } from "react-icons/fa";
import ReactMarkdown from "react-markdown";

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
				<ReactMarkdown
					components={{
						a: ({ href, children, ...props }) => {
							const isGithub = href?.includes("github.com");
							return (
								<a
									href={href}
									target="_blank"
									rel="noopener noreferrer"
									className="inline-flex items-center gap-1 underline text-primary hover:text-primary/80"
									{...props}
								>
									{children}
									{isGithub && <FaGithub className="inline-block w-4 h-4" />}
								</a>
							);
						},
						h2: ({ ...props }) => {
							// Match: ## [0.9.24] - 2025-09-25
							const headingText = Array.isArray(props.children)
								? props.children.map(String).join("")
								: String(props.children ?? "");
							const match = /(\[([^\]]+)\])\s*-\s*(\d{4}-\d{2}-\d{2})/.exec(headingText);
							if (match) {
								return (
									<div className="flex items-center gap-2 mt-6 mb-2">
										<span className="text-primary font-bold text-lg">{match[1]}</span>
										<span className="text-muted-foreground text-sm">{match[3]}</span>
									</div>
								);
							}
							return <h2 className="text-xl font-semibold mt-6 mb-2 border-b pb-1" {...props} />;
						},
						h3: ({ ...props }) => {
							const text = String(props.children);
							let color = "";
							if (/added/i.test(text)) color = "text-green-600";
							else if (/fixed/i.test(text)) color = "text-yellow-600";
							return <h3 className={`text-lg font-semibold mt-4 mb-1 ${color}`}>{props.children}</h3>;
						},
						ul: (props) => <ul className="text-md list-disc ml-6 mb-2" {...props} />,
						ol: (props) => <ol className="text-md list-decimal ml-6 mb-2" {...props} />,
						li: (props) => <li className="text-md mb-1" {...props} />,
						code: ({
							inline,
							className,
							children,
							...props
						}: React.HTMLAttributes<HTMLElement> & {
							inline?: boolean;
							className?: string;
							children?: React.ReactNode;
						}) =>
							inline ? (
								<code className="bg-zinc-100 dark:bg-zinc-800 px-1 rounded text-sm" {...props}>
									{children}
								</code>
							) : (
								<pre className="bg-zinc-100 dark:bg-zinc-800 p-3 rounded mb-4 overflow-x-auto">
									<code className={className} {...props}>
										{children}
									</code>
								</pre>
							),
						blockquote: (props) => (
							<blockquote
								className="border-l-4 border-primary pl-4 italic text-muted-foreground my-4"
								{...props}
							/>
						),
						p: (props) => <p className="mb-2" {...props} />,
						hr: () => <hr className="my-6 border-muted-foreground/20" />,
					}}
				>
					{content}
				</ReactMarkdown>
			</div>
		</div>
	);
}
