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

func generateAllUserSetsBody(username string) map[string]any {
	return map[string]any{
		"query": `
query user_sets($username: String!) {
	show_sets( filter: { user_created: { username: { _eq: $username } } }) {
		id
		user_created {
			username
		}
		set_title
		date_created
		date_updated
		show_id {
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
		}
		show_poster{
			id
			modified_on
			filesize
		}
		show_backdrop{
			id
			modified_on
			filesize
		}
		season_posters{
			id
			modified_on
			filesize
			season {
				season_number
			}
		}
		titlecards{
			id
			modified_on	
			filesize
			episode {
				episode_title
				episode_number
				season_id {
					season_number
				}
			}
		}
	}
	movie_sets( filter: { user_created: { username: { _eq: $username } } }) {
		id
		user_created {
			username
		}
		set_title
		date_created
		date_updated
		movie_id{
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
		}
		movie_poster{
			id
			modified_on
			filesize
		}
		movie_backdrop{
			id
			modified_on
			filesize
		}
	}
	collection_sets( filter: { user_created: { username: { _eq: $username } } }) {
		id
		user_created {
			username
		}
		set_title
		date_created
		date_updated
		movie_posters{
			id
			modified_on
			filesize
			movie{
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
			}
		}
		movie_backdrops{
			id
			modified_on
			filesize
			movie{
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
			}
		}
	}
	boxsets(
		 
		filter: { 
			user_created: { username: { _eq: $username } },
			_or: [
				{ movie_sets: { id: { _null: false } } },
				{ show_sets: { id: { _null: false } } },
				{ collection_sets: { id: { _null: false } } }
			]
		}
	) {
		id
		user_created {
			username
		}
		boxset_title
		date_created
		date_updated
		movie_sets{
			id
			set_title
			date_created
			date_updated
			movie_id{
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
			}
			movie_poster{
				id
				modified_on
				filesize
			}
			movie_backdrop{
				id
				modified_on
				filesize
			}
		}
		show_sets{
			id
			set_title
			date_created
			date_updated
			show_id {
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
			}
			show_poster{
				id
				modified_on
				filesize
			}
			show_backdrop{
				id
				modified_on
				filesize
			}
			season_posters{
				id
				modified_on
				filesize
				season {
					season_number
				}
			}
			titlecards{
				id
				modified_on	
				filesize
				episode {
					episode_title
					episode_number
					season_id {
						season_number
					}
				}
			}
		}
		collection_sets{
			id
			set_title
			date_created
			date_updated
			movie_posters{
				id
				modified_on
				filesize
				movie{
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
				}
			}
			movie_backdrops{
				id
				modified_on
				filesize
				movie{
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
				}
			}
		}
	}
}
`,
		"variables": map[string]string{
			"username": username,
		},
	}
}
