import React, { useEffect, useState } from "react";
import { ClientMessage } from "../types/clientMessage";
import { Box } from "@mui/material";
import { fetchAllItemsFromDB } from "../services/api.db";
import Loader from "../components/Loader";
import ErrorMessage from "../components/ErrorMessage";
import SavedSetsCard from "../components/SavedSetsCard";
import Grid from "@mui/material/Grid";

const SavedSetsPage: React.FC = () => {
	const [savedSets, setSavedSets] = useState<ClientMessage[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<boolean>(false);
	const [errorMessage, setErrorMessage] = useState<string | "">("");

	const fetchSavedSets = async () => {
		try {
			const resp = await fetchAllItemsFromDB();
			if (resp.status !== "success") {
				throw new Error(resp.message);
			}
			const sets = resp.data;
			if (!sets) {
				throw new Error("No sets found");
			}
			setSavedSets(sets);
		} catch (error) {
			setError(true);
			setErrorMessage(
				error instanceof Error
					? error.message
					: "An unknown error occurred"
			);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		fetchSavedSets();
	}, []);

	if (loading) {
		return <Loader loadingText="Loading saved sets..." />;
	}
	if (error) {
		return (
			<Box
				sx={{
					position: "relative",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					padding: 3,
					gap: 3,
					overflow: "hidden",
				}}
			>
				<ErrorMessage message={errorMessage} />
			</Box>
		);
	}

	return (
		<>
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
				{savedSets.length > 0 ? (
					savedSets.map((savedSet) => (
						<Grid
							container
							key={savedSet.MediaItem.RatingKey}
							sx={{
								display: "flex",
							}}
						>
							<SavedSetsCard
								id={savedSet.MediaItem.RatingKey}
								savedSet={savedSet}
								onUpdate={fetchSavedSets}
							/>
						</Grid>
					))
				) : (
					<> </>
				)}
			</Box>
		</>
	);
};

export default SavedSetsPage;
