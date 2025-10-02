import { Plus, TriangleAlert, Wrench } from "lucide-react";

import { FaGithub } from "react-icons/fa";
import ReactMarkdown from "react-markdown";

export function ChangelogMarkdown({ children }: { children: string }) {
	return (
		<ReactMarkdown
			components={{
				a: ({ href, children, ...props }) => {
					const isGithub = href?.includes("github.com");
					return (
						<a
							href={href}
							target="_blank"
							rel="noopener noreferrer"
							className="inline-flex items-center gap-1 underline text-primary active:scale-95 hover:brightness-120"
							{...props}
						>
							{children}
							{isGithub && <FaGithub className="inline-block w-4 h-4" />}
						</a>
					);
				},
				h2: ({ ...props }) => {
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
					if (/breaking/i.test(text)) {
						return (
							<h3 className="text-lg font-semibold mt-4 mb-1 text-red-600 flex items-center gap-2">
								<TriangleAlert className="w-5 h-5" />
								{props.children}
							</h3>
						);
					}
					if (/added/i.test(text)) {
						return (
							<h3 className="text-lg font-semibold mt-4 mb-1 text-green-600 flex items-center gap-2">
								<Plus className="w-5 h-5" />
								{props.children}
							</h3>
						);
					}
					if (/fixed/i.test(text)) {
						return (
							<h3 className="text-lg font-semibold mt-4 mb-1 text-yellow-600 flex items-center gap-2">
								<Wrench className="w-5 h-5" />
								{props.children}
							</h3>
						);
					}
					return <h3 className="text-lg font-semibold mt-4 mb-1">{props.children}</h3>;
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
			{children}
		</ReactMarkdown>
	);
}
