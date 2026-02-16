package mediux

import (
	"aura/logging"
	"aura/models"
	"aura/utils"
	"aura/utils/httpx"
	"context"
	_ "embed"
)

//go:embed gen_collection_images_by_tmdb_id.graphql
var queryCollectionImagesByTMDBID string

type collectionImagesByTMDBID_Response struct {
	Data   collectionImagesByTMDBID_Data `json:"data"`
	Errors []ErrorResponse               `json:"errors,omitempty"`
}

type collectionImagesByTMDBID_Data struct {
	Base collectionImagesByTMDBID_Collection `json:"collections_by_id"`
}

type collectionImagesByTMDBID_Collection struct {
	ID        string                                 `json:"id"`
	Name      string                                 `json:"collection_name"`
	Posters   []collectionImagesByTMDBID_ImageRecord `json:"posters"`
	Backdrops []collectionImagesByTMDBID_ImageRecord `json:"backdrops"`
}

type collectionImagesByTMDBID_ImageRecord struct {
	ImageAsset
	Collection BaseSetInfo `json:"collection_set"`
}

func GetCollectionImagesByTMDBID(ctx context.Context, tmdbID string) (collectionSets []models.SetRef, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Get Collection Images By TMDB ID", logging.LevelInfo)
	defer logAction.Complete()

	collectionSets = []models.SetRef{}
	Err = logging.LogErrorInfo{}

	// Send the GraphQL request
	resp, Err := makeGraphQLRequest(ctx, MediuxGraphQLQueryBody{
		Query:     queryCollectionImagesByTMDBID,
		Variables: map[string]any{"tmdb_id": tmdbID},
		QueryName: "getCollectionSetsByTMDBID",
	})
	if Err.Message != "" {
		return collectionSets, Err
	}

	// Decode the response
	var collectionImagesResponse collectionImagesByTMDBID_Response
	Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &collectionImagesResponse, "MediUX Collection Images By TMDB ID Response")
	if Err.Message != "" {
		return collectionSets, Err
	}

	// If any errors returned from MediUX, log and return
	if len(collectionImagesResponse.Errors) > 0 {
		logAction.SetError("Errors returned from MediUX API", "Review the errors for more details",
			map[string]any{
				"tmdb_id":   tmdbID,
				"item_type": "collection_set",
				"errors":    collectionImagesResponse.Errors,
			})
		return collectionSets, *logAction.Error
	}

	// If no collection sets, return
	if collectionImagesResponse.Data.Base.ID == "" {
		logAction.SetError("No Collection Set found", "No collection set found for the provided TMDB ID", map[string]any{
			"tmdb_id": tmdbID,
		})
		return collectionSets, *logAction.Error
	}

	// Aggregate images by collection set ID
	type setAgg struct {
		info    models.BaseSetInfo
		images  []models.ImageFile
		itemIDs []string // Not used here, but could be filled if needed
	}
	setsMap := make(map[string]*setAgg)

	// Process posters
	for _, poster := range collectionImagesResponse.Data.Base.Posters {
		agg, exists := setsMap[poster.Collection.ID]
		if !exists {
			agg = &setAgg{
				info: models.BaseSetInfo{
					ID:               poster.Collection.ID,
					Title:            poster.Collection.SetTitle,
					Type:             "collection",
					UserCreated:      poster.Collection.UserCreated.Username,
					DateCreated:      poster.Collection.DateCreated,
					DateUpdated:      poster.Collection.DateUpdated,
					Popularity:       poster.Collection.Popularity,
					PopularityGlobal: poster.Collection.PopularityGlobal,
				},
				images:  []models.ImageFile{},
				itemIDs: []string{},
			}
			setsMap[poster.Collection.ID] = agg
		}
		img := models.ImageFile{
			ID:       poster.ID,
			Type:     "collection-poster",
			Modified: poster.ModifiedOn,
			FileSize: utils.ParseFileSize(poster.Filesize),
			Src:      poster.Src,
		}
		if poster.Blurhash != nil {
			img.Blurhash = *poster.Blurhash
		}
		agg.images = append(agg.images, img)
	}

	// Process backdrops
	for _, backdrop := range collectionImagesResponse.Data.Base.Backdrops {
		agg, exists := setsMap[backdrop.Collection.ID]
		if !exists {
			agg = &setAgg{
				info: models.BaseSetInfo{
					ID:               backdrop.Collection.ID,
					Title:            backdrop.Collection.SetTitle,
					Type:             "collection",
					UserCreated:      backdrop.Collection.UserCreated.Username,
					DateCreated:      backdrop.Collection.DateCreated,
					DateUpdated:      backdrop.Collection.DateUpdated,
					Popularity:       backdrop.Collection.Popularity,
					PopularityGlobal: backdrop.Collection.PopularityGlobal,
				},
				images:  []models.ImageFile{},
				itemIDs: []string{},
			}
			setsMap[backdrop.Collection.ID] = agg
		}
		img := models.ImageFile{
			ID:       backdrop.ID,
			Type:     "collection-backdrop",
			Modified: backdrop.ModifiedOn,
			FileSize: utils.ParseFileSize(backdrop.Filesize),
			Src:      backdrop.Src,
		}
		if backdrop.Blurhash != nil {
			img.Blurhash = *backdrop.Blurhash
		}
		agg.images = append(agg.images, img)
	}

	// Convert map to []SetRef
	for _, agg := range setsMap {
		collectionSets = append(collectionSets, models.SetRef{
			PosterSet: models.PosterSet{
				BaseSetInfo: agg.info,
				Images:      agg.images,
			},
			ItemIDs: agg.itemIDs,
		})
	}

	return collectionSets, Err
}
