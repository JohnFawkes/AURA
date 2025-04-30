import React, { useState, useEffect } from "react";
import {
	Box,
	Card,
	CardContent,
	Typography,
	TextField,
	Grid,
	Tooltip,
	IconButton,
	Button,
} from "@mui/material";
import HelpOutlineIcon from "@mui/icons-material/HelpOutline";
import { fetchConfig } from "../../../frontend/src/services/api.settings";
import { AppConfig } from "../../../frontend/src/types/config";
import { useNavigate } from "react-router-dom";

const SettingsPage: React.FC = () => {
	const navigate = useNavigate();
	const [config, setConfig] = useState<AppConfig | null>(null);

	// Fetch configuration data (mocked for now)
	useEffect(() => {
		// Simulate fetching config from a file or API
		const fetchConfigFromAPI = async () => {
			try {
				const resp = await fetchConfig();
				if (!resp) {
					throw new Error("No response from API");
				}
				if (resp.status !== "success") {
					throw new Error(resp.message);
				}
				const appConfig = resp.data;
				if (!appConfig) {
					throw new Error("No config found in response");
				}
				setConfig(appConfig);
				console.log("Config fetched successfully:", appConfig);
			} catch (error) {
				console.error("Error fetching config:", error);
				setConfig(null);
			}
		};
		fetchConfigFromAPI();
	}, []);

	const handleViewLogs = () => {
		navigate("/logs");
	};

	const parseCronToHumanReadable = (cronExpression: string): string => {
		const parts = cronExpression.split(" ");
		if (parts.length !== 5) return "Current cron expression is invalid";

		const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

		const minuteText =
			minute === "0" ? "at the start of" : `at minute ${minute}`;
		const hourText = hour === "*" ? "every hour" : `hour ${hour}`;
		const dayOfMonthText =
			dayOfMonth === "*" ? "every day" : `on day ${dayOfMonth}`;
		const monthText = month === "*" ? "every month" : `in month ${month}`;
		const dayOfWeekText =
			dayOfWeek === "*"
				? "every day of the week"
				: `on day ${dayOfWeek} of the week`;

		return `Currently runs ${minuteText} ${hourText}, ${dayOfMonthText}, ${monthText}, and ${dayOfWeekText}.`;
	};

	// Help tooltips for each field
	const helpTexts = {
		MediaServerType:
			"The type of media server (e.g., Plex, Emby, Jellyfin).",
		MediaServerURL: "The base URL of the media server.",
		MediaServerToken:
			"The authentication token for accessing the media server.",
		MediaServerLibraries:
			"The list of libraries managed by the media server.",
		TMDBApiKey:
			"The API key for accessing TMDB services. This is not used in this version.",
		MediuxToken: "The authentication token for accessing Mediux services.",
		CacheImages: "Whether to cache images locally.",
		SaveImageNextToContent:
			"Whether to save images next to the associated content.",
		AutoDownloadEnabled: "Whether auto-download is enabled.",
		AutoDownloadCron: "The cron expression for scheduling auto-downloads.",
		LoggingLevel: "The logging level (e.g., DEBUG, INFO, WARN, ERROR).",
		LoggingFile: "The file path where logs are stored.",
	};

	return (
		<Box
			sx={{
				width: "70%",
				margin: "0 auto",
				padding: 2,
				display: "flex",
				flexDirection: "column",
				alignItems: "center",
				minHeight: "100vh",
			}}
		>
			<Typography variant="h4" sx={{ marginBottom: 4 }}>
				Settings
			</Typography>

			{/* Media Server Info */}
			<Card sx={{ width: "100%", marginBottom: 4 }}>
				<CardContent>
					<Typography variant="h6" sx={{ marginBottom: 2 }}>
						Media Server Information
					</Typography>
					<Grid container spacing={2}>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Server Type"
								value={config?.MediaServer.Type || ""}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={
													helpTexts.MediaServerType
												}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Server URL"
								value={config?.MediaServer.URL || ""}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={helpTexts.MediaServerURL}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Authentication Token"
								value={
									config?.MediaServer.Token
										? "***" +
										  config.MediaServer.Token.slice(-4)
										: ""
								}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={
													helpTexts.MediaServerToken
												}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<Typography
								variant="subtitle1"
								sx={{ marginBottom: 1 }}
							>
								Libraries
							</Typography>
							{config?.MediaServer.Libraries.map(
								(library, index) => (
									<Typography key={index} variant="body2">
										- {library.Name}
									</Typography>
								)
							)}
						</Grid>
					</Grid>
				</CardContent>
			</Card>

			{/* Other API Info */}
			<Card sx={{ width: "100%", marginBottom: 4 }}>
				<CardContent>
					<Typography variant="h6" sx={{ marginBottom: 2 }}>
						Other API Information
					</Typography>
					<Grid container spacing={2}>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Mediux Token"
								value={config?.Mediux.Token || ""}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={helpTexts.MediuxToken}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
					</Grid>
				</CardContent>
			</Card>

			{/* Other Settings Info */}
			<Card sx={{ width: "100%", marginBottom: 4 }}>
				<CardContent>
					<Typography variant="h6" sx={{ marginBottom: 2 }}>
						Other Settings
					</Typography>
					<Grid container spacing={2}>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Cache Images"
								value={
									config?.CacheImages ? "Enabled" : "Disabled"
								}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={helpTexts.CacheImages}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Save Images Next to Content"
								value={
									config?.SaveImageNextToContent
										? "Enabled"
										: "Disabled"
								}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={
													helpTexts.SaveImageNextToContent
												}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Auto Download"
								value={
									config?.AutoDownload.Enabled
										? "Enabled"
										: "Disabled"
								}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={
													helpTexts.AutoDownloadEnabled
												}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Auto Download Cron"
								value={config?.AutoDownload?.Cron || ""}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={
													helpTexts.AutoDownloadCron +
													` ${parseCronToHumanReadable(
														config?.AutoDownload
															?.Cron || ""
													)}`
												}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
					</Grid>
				</CardContent>
			</Card>

			{/* Logging Info */}
			<Card sx={{ width: "100%", marginBottom: 4 }}>
				<CardContent>
					<Typography variant="h6" sx={{ marginBottom: 2 }}>
						Logging
					</Typography>
					<Grid container spacing={2}>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Logging Level"
								value={config?.Logging.Level || ""}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={helpTexts.LoggingLevel}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<TextField
								label="Log File Path"
								value={config?.Logging.File || ""}
								fullWidth
								disabled
								slotProps={{
									input: {
										endAdornment: (
											<Tooltip
												title={helpTexts.LoggingFile}
											>
												<IconButton>
													<HelpOutlineIcon />
												</IconButton>
											</Tooltip>
										),
									},
								}}
							/>
						</Grid>
						<Grid size={{ xs: 12 }}>
							<Button
								variant="contained"
								color="primary"
								fullWidth
								onClick={handleViewLogs}
								sx={{ marginTop: 2 }}
							>
								View Logs
							</Button>
						</Grid>
					</Grid>
				</CardContent>
			</Card>

			<Box
				sx={{
					marginTop: "auto",
					padding: 2,
					textAlign: "center",
					width: "100%",
					borderTop: "1px solid #ccc",
				}}
			>
				<Typography variant="body2" color="textSecondary">
					App Version: {import.meta.env.VITE_APP_VERSION || "dev"}
				</Typography>
			</Box>
		</Box>
	);
};

export default SettingsPage;
