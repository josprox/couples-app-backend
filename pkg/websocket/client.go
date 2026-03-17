package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096 // 4KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Solve CORS in development
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	UserID         string
	RelationshipID string
	Send           chan WSEvent
}

type WSAction struct {
	Action string          `json:"action"` // "message", "status_update", "location_update"
	Data   json.RawMessage `json:"data"`
}

func (c *Client) ReadPump(eventHandler func(action string, data json.RawMessage) error) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS Error: %v", err)
			}
			break
		}

		var action WSAction
		if err := json.Unmarshal(message, &action); err != nil {
			log.Printf("Invalid WS action payload: %v", err)
			continue
		}

		if err := eventHandler(action.Action, action.Data); err != nil {
			log.Printf("Error handling WS action %s: %v", action.Action, err)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteJSON(message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
