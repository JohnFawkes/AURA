package api

import (
	"aura/internal/logging"
	"aura/internal/masking"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func Mediux_StartWebSocketClient() {

	for {
		err := connectAndSubscribe("show_sets")
		if err != nil {
			logging.LOGGER.Error().Timestamp().Err(err).Msg("Mediux WebSocket connection error")
		}
		logging.LOGGER.Warn().Timestamp().Msg("Reconnecting to WebSocket in 5 seconds...")
		// Wait before reconnecting
		time.Sleep(5 * time.Second)
	}

}

func connectAndSubscribe(collectionType string) error {

	// Build WebSocket URL with token
	u, err := url.Parse(MediuxBaseURL)
	if err != nil {
		return err
	}
	u.Scheme = "wss"
	u.Path = "/websocket"
	q := u.Query()
	q.Set("access_token", Global_Config.Mediux.Token)
	u.RawQuery = q.Encode()
	URL := u.String()
	maskedToken := masking.Masking_Token(Global_Config.Mediux.Token)
	maskedURL := u.String()
	maskedURL = strings.ReplaceAll(maskedURL, Global_Config.Mediux.Token, maskedToken)

	logging.LOGGER.Info().Timestamp().Str("url", maskedURL).Msg("Connecting to Mediux WebSocket")

	// Connect to WebSocket
	c, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	subscribeMsg := map[string]any{
		"type":       "subscribe",
		"collection": collectionType,
		// "query": map[string]any{
		// 	"filter": map[string]any{
		// 		"id": map[string]any{
		// 			"_eq": 711,
		// 		},
		// 	},
		// },
	}

	if err := c.WriteJSON(subscribeMsg); err != nil {
		return err
	}
	logging.LOGGER.Debug().Timestamp().Str("collection", collectionType).Msg("Subscribed to Mediux WebSocket collection")

	// Listen for messages until error/close
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		// Decode message
		var msg MediuxWebSocketResponseMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to unmarshal WebSocket message")
			continue
		}

		// Handle Ping messages
		if msg.Type == "ping" {
			pongMsg := map[string]string{
				"type": "pong",
			}
			if err := c.WriteJSON(pongMsg); err != nil {
				logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to send pong message")
			}
			continue
		}

		// Handle message based on type/event as needed
		if msg.Event == "init" {
			continue
		} else if msg.Event == "update" {
			for set := range msg.UpdatedSets {
				ctx, ld := logging.CreateLoggingContext(context.Background(), "Mediux Update - Processing Set")
				logAction := ld.AddAction("Fetching All Items from DB", logging.LevelDebug)
				ctx = logging.WithCurrentAction(ctx, logAction)

				setID := msg.UpdatedSets[set].ID
				showID := msg.UpdatedSets[set].ShowID
				showTitle := msg.UpdatedSets[set].Title

				logAction.AppendResult("set_id", setID)
				logAction.AppendResult("set_show_id", showID)
				logAction.AppendResult("set_title", showTitle)

				logging.LOGGER.Info().Timestamp().Int("set_id", msg.UpdatedSets[set].ID).Str("set_title", msg.UpdatedSets[set].Title).Msg("Show set updated")

				// Check if items exist in the database for this set
				dbItems, _, _, pageErr := DB_GetAllItemsWithFilter(
					ctx,
					showID,
					"",                  // searchLibrary
					0,                   // searchYear
					"",                  // searchTitle
					[]string{},          // librarySections
					[]string{},          // filteredTypes
					"all",               // filterAutoDownload
					false,               // multisetOnly
					[]string{},          // filteredUsernames
					500,                 // itemsPerPage
					1,                   // pageNumber
					"dateDownloaded",    // sortOption
					"desc",              // sortOrder
					strconv.Itoa(setID), // posterSetID as string
				)
				if pageErr.Message != "" {
					logging.LOGGER.Error().Timestamp().Str("error", pageErr.Message).Msg("Failed to fetch items for updated set")
					continue
				}
				if len(dbItems) == 0 {
					// No items found, nothing to process
					continue
				}
				logging.LOGGER.Debug().Timestamp().Int("set_id", setID).Str("set_title", showTitle).Int("item_count", len(dbItems)).Msg("DB items found for updated set")

				// Process each item in the set
				for _, dbItem := range dbItems {
					logging.LOGGER.Debug().Timestamp().Msgf("Processing '%s' (%s) for auto-download check", dbItem.MediaItem.Title, dbItem.LibraryTitle)
					result := AutoDownload_CheckItem(ctx, dbItem)
					switch result.OverAllResult {
					case "Success":
						logging.LOGGER.Info().Timestamp().Msgf("Successfully processed auto-download for item '%s' (%s)", dbItem.MediaItem.Title, dbItem.LibraryTitle)
						logAction.AppendResult("auto_download_result", result)
					case "Warn":
						logging.LOGGER.Warn().Timestamp().Msgf("Auto-download warning for item '%s' (%s)", dbItem.MediaItem.Title, dbItem.LibraryTitle)
						logAction.AppendResult("auto_download_result", result)
					case "Error":
						logging.LOGGER.Error().Timestamp().Msgf("Auto-download error for item '%s' (%s)", dbItem.MediaItem.Title, dbItem.LibraryTitle)
						logAction.AppendResult("auto_download_result", result)
					case "Skipped":
						logging.LOGGER.Debug().Timestamp().Msgf("Auto-download skipped for item '%s' (%s)", dbItem.MediaItem.Title, dbItem.LibraryTitle)
						logAction.AppendResult("auto_download_result", result)
					}
				}

				ld.Log()
			}
		} else {
			logging.LOGGER.Warn().Timestamp().Str("type", msg.Type).Str("event", msg.Event).Msg("Unhandled WebSocket message type/event")
		}

	}
}

type MediuxWebSocketResponseMessage struct {
	Type        string                      `json:"type"`  // Always "subscription" for responses
	Event       string                      `json:"event"` // Example: "init", "ping", "update"
	UpdatedSets []MediuxWebSocketUpdateData `json:"data"`
}

type MediuxWebSocketUpdateData struct {
	ID          int    `json:"id"`
	Title       string `json:"set_title"`
	ShowID      string `json:"show_id"`
	DateUpdated string `json:"date_updated"`
}
