package mediux

// tmdbID is the same as the setID
func generateShowRequestBody(tmdbID string) map[string]any {
	return map[string]any{
		"query": `
query shows_by_id($tmdbID: ID!) {
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
		posters {
			id
			modified_on
			filesize
			show_set {
				id
				set_title
				user_created {
					username
				}
				date_created
				date_updated
			}
		}
		backdrops {
			id
			modified_on
			filesize
			show_set {
				id
				set_title
				user_created {
					username
				}
				date_created
				date_updated
			}
		}
		seasons {
			season_number
			posters {
				id
				modified_on
				filesize
				show_set {
					id
					set_title
					user_created {
						username
					}
					date_created
					date_updated
				}
			}
			episodes {
				episode_title
				episode_number
				season_id {
					season_number
				}
				titlecards {
					id
					modified_on
					filesize
					show_set {
						id
						set_title
						user_created {
							username
						}
						date_created
						date_updated
					}
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

func generateMovieRequestBody(tmdbID string) map[string]any {
	return map[string]any{
		"query": `
query movies_by_id($tmdbID: ID!) {
	movies_by_id(id: $tmdbID) {
		id
		date_updated
		status
		title
		tagline
		release_date
		tvdb_id
		imdb_id
		trakt_id
		slug
		collection_id{
			id
			collection_name
			movies{
				id
				date_updated
				status
				title
				tagline
				release_date
				tvdb_id
				imdb_id
				trakt_id
				slug
				posters{
					id
					modified_on
					filesize
					collection_set{
						id
						set_title
						user_created {
							username
						}
						date_created
						date_updated
					}
				}
				backdrops{
					id
					modified_on
					filesize
					collection_set{
						id
						set_title
						user_created {
							username
						}
						date_created
						date_updated
					}
				}
			}
		}
		posters(filter: { movie_set: { id: { _neq: null } } }){
			id
			modified_on
			filesize
			movie_set{
				id
				set_title
				user_created {
					username
				}
				date_created
				date_updated
			}
		}
		backdrops(filter: { movie_set: { id: { _neq: null } } }){
			id
			modified_on
			filesize
			movie_set{
				id
				set_title
				user_created {
					username
				}
				date_created
				date_updated
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

func generateShowSetByIDRequestBody(setIDString string) map[string]any {
	return map[string]any{
		"query": `
query show_sets_by_id($showSetID: ID!, $showSetIDString: GraphQLStringOrFloat) {
	show_sets_by_id(id: $showSetID) {
		id
		user_created {
			username
		}
		date_created
		date_updated
		show_id {
			id
			title
			posters(filter: { show_set: { id: { _eq: $showSetIDString } } }) {
				id
				modified_on
				filesize
				show_set {
					id
				}
			}
			backdrops(filter: { show_set: { id: { _eq: $showSetIDString } } }) {
				id
				modified_on
				filesize
				show_set {
					id
				}
			}
			seasons {
				season_number
				posters(
					filter: { show_set: { id: { _eq: $showSetIDString } } }
				) {
					id
					modified_on
					filesize
					show_set {
						id
					}
				}
				episodes {
					episode_title
					episode_number
					season_id {
						season_number
					}
					titlecards(
						filter: { show_set: { id: { _eq: $showSetIDString } } }
					) {
						id
						modified_on
						filesize
						show_set {
							id
						}
					}
				}
			}
		}
	}
}
`,
		"variables": map[string]string{
			"showSetID":       setIDString,
			"showSetIDString": setIDString,
		},
	}
}

func generateUserFollowingAndHidingBody() map[string]any {
	return map[string]any{
		"query": `
query {
    user_follows (filter: {follower_id: {id: {_eq: "$CURRENT_USER" }}}) {
        followee_id {
            id
            username
        }
    }
	user_hides (filter: {hider_id: {id: {_eq: "$CURRENT_USER" }}}) {
		hiding_id {
			id
			username
		}
	}
}
`,
	}
}
