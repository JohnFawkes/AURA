import React, { useState } from "react";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import {
	Chip,
	Link,
	Box,
	Grid,
	Divider,
	MenuItem,
	IconButton,
	Menu,
	Dialog,
	DialogActions,
	DialogContent,
	DialogContentText,
	DialogTitle,
	Button,
} from "@mui/material";
import { Theme } from "@mui/material/styles";
import MediaItemPosterImage from "./MediaItemPosterImage";
import { ClientMessage } from "../types/clientMessage";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import CancelIcon from "@mui/icons-material/Cancel";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import { deleteItemFromDB, patchSelectedTypesInDB } from "../services/api.db";

const formatDate = (dateString: string) => {
	try {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat("en-US", {
			year: "numeric",
			month: "long",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		}).format(date);
	} catch {
		return "Invalid Date";
	}
};

const SavedSetsCard: React.FC<{
	id: string;
	savedSet: ClientMessage;
}> = ({ id, savedSet }) => {
	const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
	const [isEditModalOpen, setIsEditModalOpen] = useState(false);
	const [editSelectedTypes, setEditSelectedTypes] = useState<string[]>(
		savedSet.SelectedTypes
	);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [updateError, setUpdateError] = useState("");

	const chipStyles = {
		margin: "2px",
		backgroundColor: (theme: Theme) => theme.palette.background.paper,
		color: (theme: Theme) => theme.palette.primary.main,
	};

	const renderChips = () =>
		savedSet.SelectedTypes.map((type) => (
			<Chip
				key={type}
				label={
					type === "poster"
						? "Poster"
						: type === "backdrop"
						? "Backdrop"
						: type === "seasonPoster"
						? "Season Posters"
						: type === "titlecard"
						? "Title Card"
						: type
				}
				variant="outlined"
				sx={chipStyles}
			/>
		));

	const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
		setMenuAnchorEl(event.currentTarget);
	};

	const handleMenuClose = () => {
		setMenuAnchorEl(null);
	};

	const handleEdit = () => {
		handleMenuClose();
		setIsEditModalOpen(true);
	};

	const confirmEdit = async () => {
		const resp = await patchSelectedTypesInDB(id, editSelectedTypes);
		if (resp.status !== "success") {
			setUpdateError(resp.message);
		} else {
			setUpdateError("");
			window.location.reload();
			setIsEditModalOpen(false);
		}
	};

	const cancelEdit = () => {
		setIsEditModalOpen(false);
		setEditSelectedTypes(savedSet.SelectedTypes);
		setUpdateError("");
	};

	const handleDelete = () => {
		handleMenuClose();
		setIsDeleteModalOpen(true);
	};

	const confirmDelete = async () => {
		const resp = await deleteItemFromDB(id);
		if (resp.status !== "success") {
			setUpdateError(resp.message);
		} else {
			setIsDeleteModalOpen(false);
			setUpdateError("");
			window.location.reload();
		}
	};

	const cancelDelete = () => {
		setIsDeleteModalOpen(false);
		setUpdateError("");
	};

	return (
		<>
			<Grid container spacing={1} sx={{ marginBottom: 2 }}>
				<Grid size={{ xs: 12, sm: 12, md: 12 }}>
					<Card
						sx={{
							display: "flex",
							flexDirection: "column",
							padding: 2,
							boxShadow: 3,
							position: "relative",
							transition: "transform 0.2s",
							"&:hover": {
								transform: "scale(1.02)",
							},
							marginBottom: 2, // Add margin between cards
							width: {
								xs: 300,
								sm: 450,
							}, // Responsive width
						}}
					>
						{/* Auto Download Icon */}
						<Box sx={{ position: "absolute", top: 8, left: 8 }}>
							{savedSet.AutoDownload ? (
								<CheckCircleIcon color="success" />
							) : (
								<CancelIcon color="error" />
							)}
						</Box>

						{/* "..." Menu in the top right */}
						<Box sx={{ position: "absolute", top: 8, right: 8 }}>
							<IconButton onClick={handleMenuOpen}>
								<MoreVertIcon />
							</IconButton>
							<Menu
								id="edit-delete-menu"
								anchorEl={menuAnchorEl}
								open={Boolean(menuAnchorEl)}
								onClose={handleMenuClose}
							>
								<MenuItem onClick={handleEdit}>Edit</MenuItem>
								<MenuItem onClick={handleDelete}>
									Delete
								</MenuItem>
							</Menu>
						</Box>

						{/* Rest of the card content */}
						<Grid container spacing={2}>
							{/* Poster Image */}
							<Grid
								size={{ xs: 12, sm: 12, md: 12 }}
								sx={{
									display: "flex",
									justifyContent: "center",
									alignItems: "center",
								}}
							>
								<MediaItemPosterImage
									ratingKey={id}
									alt={savedSet.MediaItem.Title}
								/>
							</Grid>

							{/* Card Content */}
							<Grid size={{ xs: 12, sm: 12, md: 12 }}>
								<CardContent sx={{ padding: 0 }}>
									{/* Title */}
									<Typography variant="h6" gutterBottom>
										{savedSet.MediaItem.Title.length > 45
											? `${savedSet.MediaItem.Title.slice(
													0,
													45
											  )}...`
											: savedSet.MediaItem.Title}
									</Typography>

									{/* Year */}
									<Typography
										variant="body2"
										color="text.secondary"
									>
										{savedSet.MediaItem.Year}
									</Typography>

									{/* Last Updated */}
									<Typography
										variant="body2"
										color="text.secondary"
									>
										Last Updated:{" "}
										{formatDate(savedSet.LastUpdate || "")}
									</Typography>

									{/* Divider */}
									<Divider sx={{ marginY: 2 }} />

									{/* Selected Types */}
									<Box
										sx={{
											display: "flex",
											flexWrap: "wrap",
										}}
									>
										{renderChips()}
									</Box>

									{/* Set Link */}
									<Typography sx={{ marginTop: 2 }}>
										<Link
											href={`https://mediux.pro/sets/${savedSet.Set.ID}`}
											target="_blank"
											rel="noopener noreferrer"
											sx={{
												color: (theme: Theme) =>
													theme.palette.primary.main,
												fontWeight: "bold",
											}}
										>
											View Set: {savedSet.Set.ID}
										</Link>
									</Typography>
								</CardContent>
							</Grid>
						</Grid>
					</Card>
				</Grid>
			</Grid>

			{/* Edit Modal */}
			<Dialog
				open={isEditModalOpen}
				onClose={cancelEdit}
				aria-labelledby="edit-dialog-title"
				aria-describedby="edit-dialog-description"
				disableEnforceFocus
			>
				<DialogTitle id="edit-dialog-title">Edit Saved Set</DialogTitle>
				<DialogContent>
					<DialogContentText id="edit-dialog-description">
						Select the types you want to include in the saved set.
					</DialogContentText>
					<Box
						sx={{ display: "flex", flexWrap: "wrap", marginTop: 2 }}
					>
						{[
							"poster",
							"backdrop",
							"seasonPoster",
							"titlecard",
						].map((type) => (
							<Chip
								key={type}
								label={
									type === "poster"
										? "Poster"
										: type === "backdrop"
										? "Backdrop"
										: type === "seasonPoster"
										? "Season Posters"
										: type === "titlecard"
										? "Title Card"
										: type
								}
								variant={
									editSelectedTypes.includes(type)
										? "filled"
										: "outlined"
								}
								onClick={() => {
									setEditSelectedTypes((prev) =>
										prev.includes(type)
											? prev.filter((t) => t !== type)
											: [...prev, type]
									);
								}}
								sx={{
									margin: 1,
									cursor: "pointer",
								}}
							/>
						))}
					</Box>
					{updateError && (
						<Typography color="error">{updateError}</Typography>
					)}
				</DialogContent>
				<DialogActions>
					<Button onClick={cancelEdit} color="primary">
						Close
					</Button>
					<Button onClick={confirmEdit} color="primary" autoFocus>
						Update
					</Button>
				</DialogActions>
			</Dialog>

			{/* Delete Confirmation Modal */}
			<Dialog
				open={isDeleteModalOpen}
				onClose={cancelDelete}
				aria-labelledby="delete-confirmation-title"
				aria-describedby="delete-confirmation-description"
				disableEnforceFocus
			>
				<DialogTitle id="delete-confirmation-title">
					Confirm Delete
				</DialogTitle>
				<DialogContent>
					<DialogContentText id="delete-confirmation-description">
						Are you sure you want to delete this saved set? This
						action cannot be undone.
					</DialogContentText>
					{updateError && (
						<Typography color="error">{updateError}</Typography>
					)}
				</DialogContent>
				<DialogActions>
					<Button onClick={cancelDelete} color="primary">
						Cancel
					</Button>
					<Button onClick={confirmDelete} color="error" autoFocus>
						Delete
					</Button>
				</DialogActions>
			</Dialog>
		</>
	);
};

export default SavedSetsCard;
