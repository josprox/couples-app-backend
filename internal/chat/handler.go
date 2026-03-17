package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/jlhal/parejas/internal/location"
	"github.com/jlhal/parejas/internal/models"
	wsPkg "github.com/jlhal/parejas/pkg/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow any origin in development
	},
}

type Handler struct {
	service    Service
	locService location.Service
}

func NewHandler(service Service, locService location.Service) *Handler {
	return &Handler{service, locService}
}

func (h *Handler) GetMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	limitStr := c.DefaultQuery("limit", "50")
	beforeTime := c.Query("cursor")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	messages, err := h.service.GetMessages(userID, limit, beforeTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// ServeWS handles WebSocket requests from the peer.
func (h *Handler) ServeWS(c *gin.Context) {
	userID := c.GetString("user_id")

	// Verify relationship
	// To do this simply, we will get it from DB during connection
	relID := c.Query("relationship_id")
	if relID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relationship_id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade Error:", err)
		return
	}

	client := &wsPkg.Client{
		Hub:            wsPkg.GlobalHub,
		Conn:           conn,
		UserID:         userID,
		RelationshipID: relID,
		Send:           make(chan wsPkg.WSEvent, 256), // Use WSEvent now
	}

	client.Hub.Register <- client

	// We pass the closure to handle generic actions
	go client.WritePump()
	go client.ReadPump(func(action string, data json.RawMessage) error {
		switch action {
		case "message":
			var msg models.Message
			if err := json.Unmarshal(data, &msg); err != nil {
				return err
			}
			msg.SenderID = &client.UserID
			msg.RelationshipID = client.RelationshipID

			savedMsg, err := h.service.SaveMessage(msg)
			if err != nil {
				return err
			}

			client.Hub.Broadcast <- wsPkg.WSEvent{
				Action:         "message",
				Data:           savedMsg,
				RelationshipID: client.RelationshipID,
			}

		case "status_update":
			var update struct {
				MessageID string `json:"message_id"`
				Status    string `json:"status"`
			}
			if err := json.Unmarshal(data, &update); err != nil {
				return err
			}

			if err := h.service.UpdateMessageStatus(update.MessageID, update.Status); err != nil {
				return err
			}

			client.Hub.Broadcast <- wsPkg.WSEvent{
				Action:         "status_update",
				Data:           update,
				RelationshipID: client.RelationshipID,
			}

		case "location":
			var locData struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}
			if err := json.Unmarshal(data, &locData); err != nil {
				return err
			}

			// SAVE to DB (Non-simulation)
			if err := h.locService.SaveLocation(client.UserID, client.RelationshipID, locData.Lat, locData.Lng); err != nil {
				log.Printf("Error saving location: %v", err)
			}

			client.Hub.Broadcast <- wsPkg.WSEvent{
				Action: "location",
				Data: struct {
					UserID string  `json:"user_id"`
					Lat    float64 `json:"lat"`
					Lng    float64 `json:"lng"`
				}{
					UserID: client.UserID,
					Lat:    locData.Lat,
					Lng:    locData.Lng,
				},
				RelationshipID: client.RelationshipID,
			}
		}
		return nil
	})
}
