"use client";

import { ShowFullSetsDisplay } from "@/components/shared/show-full-set";

import { usePosterSetsStore } from "@/lib/posterSetStore";

const SetPage = () => {
	const { setType, setTitle, setAuthor, setID, posterSets } = usePosterSetsStore();

	return (
		<ShowFullSetsDisplay
			setType={setType}
			setTitle={setTitle}
			setAuthor={setAuthor}
			setID={setID}
			posterSets={posterSets}
		/>
	);
};

export default SetPage;
