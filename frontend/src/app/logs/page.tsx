"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { fetchLogContents } from "@/services/api.settings";
import {
	Card,
	CardContent,
	CardHeader,
	CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import Loader from "@/components/ui/loader";
import ErrorMessage from "@/components/ui/error-message";
import { Download, ArrowLeft } from "lucide-react";

export default function LogsPage() {
	const router = useRouter();
	const [logs, setLogs] = useState<string>("");
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<string>("");

	useEffect(() => {
		const fetchLogs = async () => {
			try {
				const resp = await fetchLogContents();
				if (!resp) throw new Error("Failed to fetch logs");
				if (resp.status !== "success") throw new Error(resp.message);

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
				setLoading(false);
			}
		};

		fetchLogs();
	}, []);

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
