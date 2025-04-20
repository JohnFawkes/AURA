package mediux

// tmdbID is the same as the setID
func generateShowRequestBody(tmdbID string) map[string]any {
	return map[string]any{
		"query": `
query shows_by_id($tmdbID: ID!, $tmdbIDString: String!) {
    shows_by_id(id: $tmdbID) {
		id
		date_updated
		status
		title
		tagline
		first_air_date
		tvdb_id
		imdb_id
		trakt_id
		slug
		show_sets(
			filter: { show_id: { id: { _eq: $tmdbIDString } }, files: { file_type: { _neq: "album" } } }
			sort: "-user_created.username"
		) {
			id
			user_created {
				username
			}
			date_created
			date_updated
			files (
			sort: ["-season.season_number", "-episode.season_id.season_number", "-episode.episode_number"]
			){
				id
				file_type
				modified_on
				season {
					season_number
				}
				episode {
					episode_title
					episode_number
					season_id {
						season_number
					}
				}
			}
		}
	}
}
	`,
		"variables": map[string]string{
			"tmdbID":       tmdbID,
			"tmdbIDString": tmdbID,
		},
	}
}

func generateMovieRequestBody(tmdbID string) map[string]any {
	return map[string]any{
		"query": `
query movies_by_id($tmdbID: ID!, $tmdbIDString: String!) {
	movies_by_id(id: $tmdbID) {
		id
		date_updated
		status
		title
		release_date
		tagline
		tvdb_id
		imdb_id
		trakt_id
		slug
		collection_id {
			id
			collection_name
			collection_sets {
				id
				user_created {
					username
				}
				date_created
				date_updated
				files(filter: { movie: { id: { _eq: $tmdbIDString } }, file_type: { _neq: "album" } }) {
					id
					file_type
					modified_on
					movie {
						id
					}
				}
			}
		}
		movie_sets(sort: "-user_created.username") {
			id
			user_created {
				username
			}
			date_created
			date_updated
			files(filter: { movie: { id: { _eq: $tmdbIDString } }, file_type: { _neq: "album" } }) {
				id
				file_type
				modified_on
				movie {
					id
				}
			}
		}
	}
}
`,
		"variables": map[string]string{
			"tmdbID":       tmdbID,
			"tmdbIDString": tmdbID,
		},
	}
}

func generateCollectionSetByIDRequestBody(collectionID, tmdbID string) map[string]any {
	return map[string]any{
		"query": `
query collection_sets_by_id($collectionID: ID!, $tmdbID: String!) {
	collection_sets_by_id(id: $collectionID) {
		id
		user_created {
			username
		}
		date_created
		date_updated
		files(filter: { movie: { id: { _eq: $tmdbID } }, file_type: { _neq: "album" } }) {
			id
			file_type
			modified_on
			movie {
				id
			}
		}
	}
}

`,
		"variables": map[string]string{
			"collectionID": collectionID,
			"tmdbID":       tmdbID,
		},
	}
}

func generateMovieSetByIDRequestBody(movieSetID, tmdbID string) map[string]any {
	return map[string]any{
		"query": `
query movie_sets_by_id($movieSetID: ID!, $tmdbID: String!) {
	movie_sets_by_id(id: $movieSetID) {
		id
		user_created {
			username
		}
		date_created
		date_updated
		files(filter: { movie: { id: { _eq: $tmdbID } }, file_type: { _neq: "album" } }) {
			id
			file_type
			modified_on
			movie {
				id
			}
		}
	}
}
`,
		"variables": map[string]string{
			"movieSetID": movieSetID,
			"tmdbID":     tmdbID,
		},
	}
}

func generateShowSetByIDRequestBody(showSetID string) map[string]any {
	return map[string]any{
		"query": `
query show_sets_by_id($showSetID: ID!) {
	show_sets_by_id(id: $showSetID) {
		id
		user_created {
			username
		}
		date_created
		date_updated
		files  (
			sort: ["-season.season_number", "-episode.season_id.season_number", "-episode.episode_number"]
			){
			id
			file_type
			modified_on
			season {
				season_number
			}
			episode {
				episode_title
				episode_number
				season_id {
					season_number
				}
			}
		}
	}
	}
`,
		"variables": map[string]string{
			"showSetID": showSetID,
		},
	}
}
