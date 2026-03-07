package mediux

import (
	"aura/logging"
	"aura/models"
	"aura/utils"
	"aura/utils/httpx"
	"context"
	_ "embed"
)

//go:embed gen_collection_images_by_movie_tmdbids.graphql
var queryCollectionImagesByMovieTMDBIDs string

type collectionImagesByMovieTMDBIDs_Response struct {
	Data   collectionImagesByMovieTMDBIDs_Data `json:"data"`
	Errors []ErrorResponse                     `json:"errors,omitempty"`
}

type collectionImagesByMovieTMDBIDs_Data struct {
	Base []collectionImagesByMovieTMDBIDs_Movie `json:"movies"`
}

type collectionImagesByMovieTMDBIDs_Movie struct {
	ID         string                              `json:"id"`
	Title      string                              `json:"title"`
	Collection collectionImagesByTMDBID_Collection `json:"collection_id"`
}

func GetCollectionImagesByMovieTMDBIDs(ctx context.Context, tmdbIDs []string) (collectionSets []models.SetRef, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Get Collection Images By Movie TMDB IDs", logging.LevelInfo)
	defer logAction.Complete()

	collectionSets = []models.SetRef{}
	Err = logging.LogErrorInfo{}

	// Send the GraphQL request
	resp, Err := makeGraphQLRequest(ctx, MediuxGraphQLQueryBody{
		Query:     queryCollectionImagesByMovieTMDBIDs,
		Variables: map[string]any{"tmdb_ids": tmdbIDs},
		QueryName: "getCollectionImagesByMovieTMDBIDs",
	})
	if Err.Message != "" {
		return collectionSets, Err
	}

	// Decode the response
	var collectionImagesResponse collectionImagesByMovieTMDBIDs_Response
	Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &collectionImagesResponse, "MediUX Collection Images By Movie TMDB IDs Response")
	if Err.Message != "" {
		return collectionSets, Err
	}

	// If any errors returned from MediUX, log and return
	if len(collectionImagesResponse.Errors) > 0 {
		logAction.SetError("MediUX GraphQL Errors", "Errors returned from MediUX GraphQL", map[string]any{
			"errors": collectionImagesResponse.Errors,
		})
		return collectionSets, Err
	}

	// Aggregate by collection_set.id
	type setAgg struct {
		info    models.BaseSetInfo
		images  []models.ImageFile
		itemIDs map[string]struct{}
		seen    map[string]struct{}
	}
	setsMap := make(map[string]*setAgg)

	// NEW: only use the first movie returned
	if len(collectionImagesResponse.Data.Base) == 0 {
		return collectionSets, Err
	}
	movie := collectionImagesResponse.Data.Base[0]

	coll := movie.Collection
	if coll.ID == "" {
		return collectionSets, Err
	}

	// Posters
	for _, poster := range coll.Posters {
		set := poster.Collection
		if set.ID == "" {
			continue
		}

		agg, exists := setsMap[set.ID]
		if !exists {
			agg = &setAgg{
				info: models.BaseSetInfo{
					ID:               set.ID,
					Title:            set.SetTitle,
					Type:             "collection",
					UserCreated:      set.UserCreated.Username,
					DateCreated:      set.DateCreated,
					DateUpdated:      set.DateUpdated,
					Popularity:       set.Popularity,
					PopularityGlobal: set.PopularityGlobal,
				},
				images:  []models.ImageFile{},
				itemIDs: map[string]struct{}{},
				seen:    map[string]struct{}{},
			}
			setsMap[set.ID] = agg
		}

		img := models.ImageFile{
			ID:          poster.ID,
			Type:        "collection_poster",
			Modified:    poster.ModifiedOn,
			FileSize:    utils.ParseFileSize(poster.Filesize),
			Src:         poster.Src,
			ItemTMDB_ID: coll.ID,
		}
		if poster.Blurhash != nil {
			img.Blurhash = *poster.Blurhash
		}

		key := img.Type + ":" + img.ID
		if _, ok := agg.seen[key]; !ok {
			agg.seen[key] = struct{}{}
			agg.images = append(agg.images, img)
		}

		agg.itemIDs[movie.ID] = struct{}{}
	}

	// Backdrops
	for _, backdrop := range coll.Backdrops {
		set := backdrop.Collection
		if set.ID == "" {
			continue
		}

		agg, exists := setsMap[set.ID]
		if !exists {
			agg = &setAgg{
				info: models.BaseSetInfo{
					ID:               set.ID,
					Title:            set.SetTitle,
					Type:             "collection",
					UserCreated:      set.UserCreated.Username,
					DateCreated:      set.DateCreated,
					DateUpdated:      set.DateUpdated,
					Popularity:       set.Popularity,
					PopularityGlobal: set.PopularityGlobal,
				},
				images:  []models.ImageFile{},
				itemIDs: map[string]struct{}{},
				seen:    map[string]struct{}{},
			}
			setsMap[set.ID] = agg
		}

		img := models.ImageFile{
			ID:          backdrop.ID,
			Type:        "collection_backdrop",
			Modified:    backdrop.ModifiedOn,
			FileSize:    utils.ParseFileSize(backdrop.Filesize),
			Src:         backdrop.Src,
			ItemTMDB_ID: coll.ID,
		}
		if backdrop.Blurhash != nil {
			img.Blurhash = *backdrop.Blurhash
		}

		key := img.Type + ":" + img.ID
		if _, ok := agg.seen[key]; !ok {
			agg.seen[key] = struct{}{}
			agg.images = append(agg.images, img)
		}

		agg.itemIDs[movie.ID] = struct{}{}
	}

	// Convert map to []SetRef
	for _, agg := range setsMap {
		itemIDs := make([]string, 0, len(agg.itemIDs))
		for id := range agg.itemIDs {
			itemIDs = append(itemIDs, id)
		}
		collectionSets = append(collectionSets, models.SetRef{
			PosterSet: models.PosterSet{
				BaseSetInfo: agg.info,
				Images:      agg.images,
			},
			ItemIDs: itemIDs,
		})
	}

	return collectionSets, Err
}
