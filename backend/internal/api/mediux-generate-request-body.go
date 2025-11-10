package api

// tmdbID is the same as the setID
func Mediux_GenerateShowRequestBody(tmdbID string) map[string]any {
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
		poster_path
		backdrop_path
		posters {
			id
			modified_on
			filesize
			src
			blurhash
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
			src
			blurhash
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
				src
				blurhash
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
					src
					blurhash
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

func Mediux_GenerateMovieRequestBody(tmdbID string) map[string]any {
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
		poster_path
		backdrop_path
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
				poster_path
				backdrop_path
				posters{
					id
					modified_on
					filesize
					src
					blurhash
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
					src
					blurhash
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
			src
			blurhash
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
			src
			blurhash
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

func Mediux_GenerateShowSetByIDRequestBody(setIDString string) map[string]any {
	return map[string]any{
		"query": `
query show_sets_by_id($showSetID: ID!, $showSetIDString: GraphQLStringOrFloat) {
	show_sets_by_id(id: $showSetID) {
		id
		set_title
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
				src
				blurhash
				show_set {
					id
				}
			}
			backdrops(filter: { show_set: { id: { _eq: $showSetIDString } } }) {
				id
				modified_on
				filesize
				src
				blurhash
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
					src
					blurhash
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
						src
						blurhash
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

func Mediux_GenerateMovieSetByIDRequestBody(setIDString string) map[string]any {
	return map[string]any{
		"query": `
query movie_sets_by_id($movieSetID: ID!, $movieSetIDString: GraphQLStringOrFloat) {
	movie_sets_by_id(id: $movieSetID) {
		id
		set_title
		user_created {
			username
		}
		date_created
		date_updated
		movie_id {
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
			poster_path
			backdrop_path
			posters(filter: { movie_set: { id: { _eq: $movieSetIDString } } }) {
				id
				modified_on	
				filesize
				src
				blurhash
				movie_set {
					id
				}
			}
			backdrops(filter: { movie_set: { id: { _eq: $movieSetIDString } } }) {
				id
				modified_on
				filesize
				src
				blurhash
				movie_set {
					id
				}
			}
		}
	}
}
`,
		"variables": map[string]string{
			"movieSetID":       setIDString,
			"movieSetIDString": setIDString,
		},
	}
}

func Mediux_GenerateCollectionSetByIDRequestBody(setIDString string, movieIDString string) map[string]any {
	return map[string]any{
		"query": `
query collection_sets_by_id($collectionSetID: ID!, $collectionSetIDString: GraphQLStringOrFloat!, $movieIDString: String!) {
	collection_sets_by_id(id: $collectionSetID) {
		id
		set_title
		user_created {
			username
		}
		date_created
		date_updated
		collection_id {
			id
			collection_name
			movies (filter: { id: { _eq: $movieIDString } }) {
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
				poster_path
				backdrop_path
				posters (filter: { collection_set: { id: { _eq: $collectionSetIDString } } }) {
					id
					modified_on
					filesize
					src
					blurhash
					collection_set {
						id
					}
				}
				backdrops (filter: { collection_set: { id: { _eq: $collectionSetIDString } } }) {
					id
					modified_on
					filesize
					src
					blurhash
					collection_set {
						id
					}
				}
			}
		}
	}
}
`,
		"variables": map[string]string{
			"collectionSetID":       setIDString,
			"collectionSetIDString": setIDString,
			"movieIDString":         movieIDString,
		},
	}
}

func Mediux_GenerateUserFollowingAndHidingBody() map[string]any {
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

func Mediux_GenerateAllUserSetsBody(username string) map[string]any {
	return map[string]any{
		"query": `
query user_sets($username: String!) {
	show_sets(limit: 10000, filter: { user_created: { username: { _eq: $username } } }) {
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
			poster_path	
			backdrop_path
		}
		show_poster{
			id
			modified_on
			filesize
			src
			blurhash
		}
		show_backdrop{
			id
			modified_on
			filesize
			src
			blurhash
		}
		season_posters{
			id
			modified_on
			filesize
			src
			blurhash
			season {
				season_number
			}
		}
		titlecards{
			id
			modified_on	
			filesize
			src
			blurhash
			episode {
				episode_title
				episode_number
				season_id {
					season_number
				}
			}
		}
	}
	movie_sets(limit: 10000, filter: { user_created: { username: { _eq: $username } } }) {
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
				poster_path
				backdrop_path
		}
		movie_poster{
			id
			modified_on
			filesize
			src
			blurhash
		}
		movie_backdrop{
			id
			modified_on
			filesize
			src
			blurhash
		}
	}
	collection_sets(limit: 10000, filter: { user_created: { username: { _eq: $username } } }) {
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
			src
			blurhash
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
				poster_path
				backdrop_path
			}
		}
		movie_backdrops{
			id
			modified_on
			filesize
			src
			blurhash
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
				poster_path
				backdrop_path
			}
		}
	}
	boxsets(
		limit: 10000,
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
				poster_path
				backdrop_path
			}
			movie_poster{
				id
				modified_on
				filesize
				src
				blurhash
			}
			movie_backdrop{
				id
				modified_on
				filesize
				src
				blurhash
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
				poster_path	
				backdrop_path
			}
			show_poster{
				id
				modified_on
				filesize
				src
				blurhash
			}
			show_backdrop{
				id
				modified_on
				filesize
				src
				blurhash
			}
			season_posters{
				id
				modified_on
				filesize
				src
				blurhash
				season {
					season_number
				}
			}
			titlecards{
				id
				modified_on	
				filesize
				src
				blurhash
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
				src
				blurhash
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
					poster_path
					backdrop_path
				}
			}
			movie_backdrops{
				id
				modified_on
				filesize
				src
				blurhash
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
					poster_path
					backdrop_path
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

func Mediux_GenerateCollectionImagesByMovieIDsBody(tmdbIDs []string) map[string]any {
	return map[string]any{
		"query": `
query collections_images_by_movie_ids($ids: [String!]) {
  movies(filter: {id: {_in: $ids}}) {
    id
    title
    collection_id {
      id
      collection_name
      posters {
        id
        modified_on
        filesize
        src
        blurhash
        uploaded_by {
          username
        }
        collection_set{
          id
          set_title
        }
      }
      backdrops {
        id
        modified_on
        filesize
        src
        blurhash
        uploaded_by {
          username
        }
        collection_set{
          id
          set_title
        }
      }
    }
  }
}
`,
		"variables": map[string]any{
			"ids": tmdbIDs,
		},
	}
}
