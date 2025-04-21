import MovieRounded from "@mui/icons-material/MovieRounded";
import InputAdornment from "@mui/material/InputAdornment";
import TextField from "@mui/material/TextField";
import React from "react";
import HomeLibrarySelect from "./HomeLibrarySelect";
import { LibrarySection } from "../types/mediaItem";

const HomeTopSection: React.FC<{
	searchQuery: string;
	librarySections: LibrarySection[];
	filteredLibraries: string[];
	onSearchChange: (value: string) => void;
	onLibraryChange: (libraries: string[]) => void;
}> = ({
	searchQuery,
	librarySections,
	filteredLibraries,
	onSearchChange,
	onLibraryChange,
}) => {
	return (
		<>
			<TextField
				id="input-with-icon-textfield"
				label="Search Media Items"
				placeholder="Media Title"
				fullWidth
				value={searchQuery}
				onChange={(e) => onSearchChange(e.target.value)}
				slotProps={{
					input: {
						startAdornment: (
							<InputAdornment position="start">
								<MovieRounded />
							</InputAdornment>
						),
					},
				}}
				variant="outlined"
			/>
			<HomeLibrarySelect
				filteredListOptions={librarySections}
				selectedLibrary={filteredLibraries}
				onLibraryChange={onLibraryChange}
			/>
		</>
	);
};

export default HomeTopSection;
