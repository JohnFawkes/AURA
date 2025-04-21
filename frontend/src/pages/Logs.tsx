// filepath: /Users/moose/sync/projects/poster-setter/frontend/src/pages/Logs.tsx
import React, { useState, useEffect } from "react";
import { Box, Typography, Button, TextField } from "@mui/material";
import { fetchLogContents } from "../services/api.settings";
import { useNavigate } from "react-router-dom";

const Logs: React.FC = () => {
	const navigate = useNavigate();
	const [logs, setLogs] = useState<string>("");

	useEffect(() => {
		// Fetch logs from the server or file
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
				console.error("Error fetching logs:", error);
				setLogs("Failed to load logs.");
			}
		};

		fetchLogs();
	}, []);

	return (
		<Box
			sx={{
				width: "90%",
				margin: "0 auto",
				padding: 2,
				display: "flex",
				flexDirection: "column",
				alignItems: "center",
				minHeight: "100vh",
			}}
		>
			<Typography variant="h4" sx={{ marginBottom: 4 }}>
				Logs
			</Typography>
			<TextField
				value={logs}
				multiline
				fullWidth
				rows={20}
				variant="outlined"
			/>
			<Button
				variant="contained"
				color="primary"
				fullWidth
				sx={{ marginTop: 2 }}
				onClick={() => navigate("/settings")}
			>
				Back to Settings
			</Button>

			<Button
				variant="contained"
				color="secondary"
				fullWidth
				sx={{ marginTop: 2 }}
				onClick={() => {
					const blob = new Blob([logs], { type: "text/plain" });
					const url = URL.createObjectURL(blob);
					const a = document.createElement("a");
					a.href = url;
					a.download = "logs.txt";
					a.click();
					URL.revokeObjectURL(url);
				}}
			>
				Download Logs
			</Button>
		</Box>
	);
};

export default Logs;
