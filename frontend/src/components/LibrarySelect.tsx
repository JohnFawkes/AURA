import React from "react";
import {
	Select,
	MenuItem,
	InputLabel,
	FormControl,
	Box,
	Chip,
	OutlinedInput,
} from "@mui/material";
import { LibrarySection } from "../types/mediaItem";
import { Theme, useTheme } from "@mui/material/styles";
import { SelectChangeEvent } from "@mui/material"; // Ensure this import is present

const ITEM_HEIGHT = 48;
const ITEM_PADDING_TOP = 8;
const MenuProps = {
	PaperProps: {
		style: {
			maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
			width: 250,
		},
	},
};

interface LibrarySelectProps {
	filteredListOptions: LibrarySection[];
	selectedLibrary: string[];
	onLibraryChange: (value: string[]) => void;
}

function getStyles(
	name: string,
	selectedLibraries: readonly string[],
	theme: Theme
) {
	return {
		fontWeight: selectedLibraries.includes(name)
			? theme.typography.fontWeightMedium
			: theme.typography.fontWeightRegular,
	};
}

const LibrarySelect: React.FC<LibrarySelectProps> = ({
	filteredListOptions,
	selectedLibrary,
	onLibraryChange,
}) => {
	const theme = useTheme();

	const handleChange = (event: SelectChangeEvent<string[]>) => {
		const {
			target: { value },
		} = event;
		onLibraryChange(
			typeof value === "string" ? value.split(",") : (value as string[])
		);
	};

	return (
		<FormControl sx={{ mt: 2, mx: 0 }} fullWidth>
			<InputLabel id="library-select-label">
				Filter by Library Name
			</InputLabel>
			<Select
				labelId="library-select-label"
				id="library-select"
				multiple
				value={selectedLibrary}
				onChange={handleChange}
				input={
					<OutlinedInput
						id="select-multiple-chip"
						label="Filter by Library Name"
					/>
				}
				renderValue={(selected) => (
					<Box sx={{ display: "flex", flexWrap: "wrap", gap: 0.5 }}>
						{selected.map((value) => (
							<Chip key={value} label={value} />
						))}
					</Box>
				)}
				MenuProps={MenuProps}
			>
				{filteredListOptions.map((section) => (
					<MenuItem
						key={section.ID}
						value={section.Title}
						style={getStyles(section.Title, selectedLibrary, theme)}
					>
						{section.Title}
					</MenuItem>
				))}
			</Select>
		</FormControl>
	);
};

export default LibrarySelect;
