"use client";
import React, { useEffect, useState } from "react";
import { ClientMessage } from "@/types/clientMessage";
import { fetchAllItemsFromDB } from "@/services/api.db";
import Loader from "@/components/ui/loader";
import ErrorMessage from "@/components/ui/error-message";
import SavedSetsCard from "@/components/ui/saved-sets-cards";

const SavedSetsPage: React.FC = () => {
	const [isMounted, setIsMounted] = useState(false);
	const [savedSets, setSavedSets] = useState<ClientMessage[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(false);
	const [errorMessage, setErrorMessage] = useState<string>("");

	const fetchSavedSets = async () => {
		if (isMounted) return;
		setIsMounted(true);
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
			setIsMounted(false);
		}
	};

	useEffect(() => {
		fetchSavedSets();
	}, []);

	if (loading) {
		return <Loader message="Loading saved sets..." />;
	}

	if (error) {
		return (
			<div className="flex flex-col items-center p-6 gap-4">
				<ErrorMessage message={errorMessage} />
			</div>
		);
	}

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			{savedSets.length > 0 ? (
				savedSets.map((savedSet) => (
					<SavedSetsCard
						key={savedSet.MediaItem.RatingKey}
						id={savedSet.MediaItem.RatingKey}
						savedSet={savedSet}
						onUpdate={fetchSavedSets}
					/>
				))
			) : (
				<p className="text-muted-foreground">No saved sets found.</p>
			)}
		</div>
	);
};

export default SavedSetsPage;
