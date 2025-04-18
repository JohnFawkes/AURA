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
			files {
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
query movie($tmdbID: String!) {
	movies(filter: { id: { _eq: $tmdbID } }) {
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
		movie_sets(sort: "-user_created.username") {
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
}
`,
		"variables": map[string]string{
			"tmdbID": tmdbID,
		},
	}
}

func generateCollectionSetByIDRequestBody(collectionID string) map[string]any {
	return map[string]any{
		"query": `
query collectionSet($collectionID: GraphQLStringOrFloat!) {
	movies(
		filter: {
			collection_id: { collection_sets: { id: { _eq: $collectionID } } }
		}
	) {
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
			collection_sets(filter: { id: { _eq: $collectionID } }) {
				id
				user_created {
					username
				}
				date_created
				date_updated
				files(filter: { file_type: { _neq: "album" } }) {
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
}
`,
		"variables": map[string]string{
			"collectionID": collectionID,
		},
	}
}

func generateMovieSetByIDRequestBody(movieSetID string) map[string]any {
	return map[string]any{
		"query": `
query movie($movieSetID: GraphQLStringOrFloat!) {
	movies(filter: { movie_sets: { id: { _eq: $movieSetID } } }) {
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
		movie_sets(
			filter: { id: { _eq: $movieSetID } }
			sort: "-user_created.username"
		) {
			id
			user_created {
				username
			}
			date_created
			date_updated
			files(filter: { file_type: { _neq: "album" } }) {
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
			"movieSetID": movieSetID,
		},
	}
}

func generateShowSetByIDRequestBody(showSetID string) map[string]any {
	return map[string]any{
		"query": `
query showSet($showSetID: GraphQLStringOrFloat!) {
	show_sets(filter: { id: { _eq: $showSetID } }) {
		id
		user_created {
			username
		}
		date_created
		date_updated
		files {
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
