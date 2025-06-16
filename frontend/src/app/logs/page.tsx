"use client";

import { fetchLogContents } from "@/services/api.settings";
import { ReturnErrorMessage } from "@/services/api.shared";
import { ArrowLeft, Download } from "lucide-react";

import { useEffect, useState } from "react";

import { useRouter } from "next/navigation";

import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";

export default function LogsPage() {
	const router = useRouter();
	const [logs, setLogs] = useState<string>("");
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<APIResponse<string> | null>(null);
	const [isMounted, setIsMounted] = useState(false);

	useEffect(() => {
		if (isMounted) return;
		setIsMounted(true);
		if (typeof window !== "undefined") {
			document.title = "Aura | Logs";
		}

		const fetchLogs = async () => {
			try {
				setLoading(true);
				const response = await fetchLogContents();

				if (response.status === "error") {
					setError(response);
					setLogs("");
					return;
				}

				setLogs(response.data || "");
				setError(null);
			} catch (error) {
				setError(ReturnErrorMessage<string>(error));
				setLogs("");
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
					<ErrorMessage error={error} />
				</div>
			) : (
				<Card className="shadow-lg">
					<CardHeader>
						<h1 className="text-2xl font-bold text-center">Application Logs</h1>
					</CardHeader>
					<CardContent className="md:p-6">
						<Textarea
							value={logs}
							readOnly
							className="font-mono text-sm bg-muted/50 h-[50vh] md:h-[65vh] overflow-auto w-full"
						/>
					</CardContent>
					<CardFooter className="flex flex-col sm:flex-row justify-center gap-3">
						<Button variant="outline" className="w-full sm:w-auto" onClick={() => router.push("/settings")}>
							<ArrowLeft className="mr-2 h-4 w-4" /> Back to Settings
						</Button>
						<Button variant="default" className="w-full sm:w-auto" onClick={handleDownloadLogs}>
							<Download className="mr-2 h-4 w-4" /> Download Logs
						</Button>
					</CardFooter>
				</Card>
			)}
		</div>
	);
}
