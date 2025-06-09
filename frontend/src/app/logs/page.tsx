"use client";

import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
} from "@/components/ui/card";
import { ErrorMessage } from "@/components/ui/error-message";
import Loader from "@/components/ui/loader";
import { Textarea } from "@/components/ui/textarea";
import { log } from "@/lib/logger";
import { fetchLogContents } from "@/services/api.settings";
import { ArrowLeft, Download } from "lucide-react";
import { useRouter } from "next/navigation";
import React, { useEffect, useState } from "react";

export default function LogsPage() {
	const router = useRouter();
	const [logs, setLogs] = useState<string>("");
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<string>("");
	const [isMounted, setIsMounted] = useState(false);

	useEffect(() => {
		if (isMounted) return;
		setIsMounted(true);
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = "Aura | Logs";
		}
		const fetchLogs = async () => {
			try {
				const resp = await fetchLogContents();

				if (!resp) {
					throw new Error("Failed to fetch logs");
				}
				if (resp.status !== "success") {
					throw new Error(resp.message);
				}

				const logContents = resp.data;
				if (!logContents) {
					throw new Error("No log contents found");
				}

				setLogs(logContents);
			} catch (error) {
				setError(
					error instanceof Error
						? error.message
						: "An unknown error occurred"
				);
			} finally {
				log("LogsPage - Fetching logs completed");
				setLoading(false);
			}
		};

		fetchLogs();
	}, [isMounted]);

	const handleDownloadLogs = () => {
		const blob = new Blob([logs], {
			type: "text/plain",
		});
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "logs.txt";
		a.click();
		URL.revokeObjectURL(url);
	};

	return (
		<div className="container mx-auto p-6 max-w-6xl">
			{loading ? (
				<div className="flex justify-center items-center min-h-[70vh]">
					<Loader message="Loading logs..." />
				</div>
			) : error ? (
				<div className="mt-8">
					<ErrorMessage message={error} />
				</div>
			) : (
				<Card className="shadow-lg">
					<CardHeader>
						<h1 className="text-2xl font-bold text-center">
							Application Logs
						</h1>
					</CardHeader>
					<CardContent className="md:p-6">
						<Textarea
							value={logs}
							readOnly
							className="font-mono text-sm bg-muted/50 h-[50vh] md:h-[65vh] overflow-auto w-full"
						/>
					</CardContent>
					<CardFooter className="flex flex-col sm:flex-row justify-center gap-3">
						<Button
							variant="outline"
							className="w-full sm:w-auto"
							onClick={() => router.push("/settings")}
						>
							<ArrowLeft className="mr-2 h-4 w-4" /> Back to
							Settings
						</Button>
						<Button
							variant="default"
							className="w-full sm:w-auto"
							onClick={handleDownloadLogs}
						>
							<Download className="mr-2 h-4 w-4" /> Download Logs
						</Button>
					</CardFooter>
				</Card>
			)}
		</div>
	);
}
