package mediux

import (
	"aura/config"
	"aura/logging"
	"encoding/json"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func StartWebSocketClient() {
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
	u, err := url.Parse(MediuxApiURL)
	if err != nil {
		return err
	}
	u.Scheme = "wss"
	u.Path = "/websocket"
	q := u.Query()
	q.Set("access_token", config.Current.Mediux.ApiToken)
	u.RawQuery = q.Encode()
	URL := u.String()
	maskedToken := config.MaskToken(config.Current.Mediux.ApiToken)
	maskedURL := u.String()
	maskedURL = strings.ReplaceAll(maskedURL, config.Current.Mediux.ApiToken, maskedToken)

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

		// Write payload to file
		err = writePayloadToFile(msg, path.Join(config.ConfigPath, "mediux-websocket", time.Now().Format("20060102_150405")+"_mediux_ws_payload.json"))
		if err != nil {
			logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to write WebSocket payload to file")
		}

		// Handle message based on type/event as needed
		if msg.Event == "init" {
			continue
		} else if msg.Event == "update" {
			for set := range msg.UpdatedSets {
				logging.LOGGER.Info().Timestamp().
					Int("set_id", msg.UpdatedSets[set].ID).
					Str("set_title", msg.UpdatedSets[set].Title).
					Str("show_id", msg.UpdatedSets[set].ShowID).
					Str("date_updated", msg.UpdatedSets[set].DateUpdated).
					Msg("Received update for Mediux Show Set")

			}
		} else {
			logging.LOGGER.Warn().Timestamp().
				Str("type", msg.Type).
				Str("event", msg.Event).
				Interface("full", msg).
				Msg("Unhandled WebSocket message type/event")
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

func writePayloadToFile(payload any, filename string) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
