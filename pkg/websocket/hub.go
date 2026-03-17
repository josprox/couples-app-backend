package websocket

import (
	"encoding/json"
	"log"

	redisPkg "github.com/jlhal/parejas/pkg/redis"
)

// WSEvent is a generic wrapper for all websocket communications
type WSEvent struct {
	Action         string      `json:"action"`
	Data           interface{} `json:"data"`
	RelationshipID string      `json:"-"`
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients inside a Relationship Hub
	// map[relationship_id]map[*Client]bool
	rooms map[string]map[*Client]bool

	// Broadcast channel locally
	Broadcast chan WSEvent

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

var GlobalHub *Hub

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		Broadcast:  make(chan WSEvent),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.rooms[client.RelationshipID] == nil {
				h.rooms[client.RelationshipID] = make(map[*Client]bool)
				// Subscribe to Redis PubSub for this room if it's the first connection
				go h.subscribeToRoom(client.RelationshipID)
			}
			h.rooms[client.RelationshipID][client] = true
			log.Printf("Client %s connected to room %s", client.UserID, client.RelationshipID)

		case client := <-h.Unregister:
			if _, ok := h.rooms[client.RelationshipID][client]; ok {
				delete(h.rooms[client.RelationshipID], client)
				close(client.Send)
				
				if len(h.rooms[client.RelationshipID]) == 0 {
					delete(h.rooms, client.RelationshipID)
					// Note: Unsubscribing from Redis could be done here
				}
				log.Printf("Client %s disconnected from room %s", client.UserID, client.RelationshipID)
			}

		case event := <-h.Broadcast:
			// Publish to REDIS so other instances get it!
			msgBytes, _ := json.Marshal(event)
			err := redisPkg.Client.Publish(redisPkg.Ctx, "room:"+event.RelationshipID, msgBytes).Err()
			if err != nil {
				log.Printf("Redis Publish error: %v", err)
			}
		}
	}
}

// subscribeToRoom listens to Redis channel and pushes to connected local clients
func (h *Hub) subscribeToRoom(relationshipID string) {
	pubsub := redisPkg.Client.Subscribe(redisPkg.Ctx, "room:"+relationshipID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var event WSEvent
		err := json.Unmarshal([]byte(msg.Payload), &event)
		if err != nil {
			log.Printf("Error unmarshaling redis pubsub message: %v", err)
			continue
		}

		// Distribute to local clients in this room
		event.RelationshipID = relationshipID // Ensure consistency
		h.distributeLocal(event)
	}
}

// distributeLocal sends a message strictly to Local websocket connections of the Hub
func (h *Hub) distributeLocal(event WSEvent) {
	if clients, ok := h.rooms[event.RelationshipID]; ok {
		for client := range clients {
			select {
			case client.Send <- event:
			default:
				close(client.Send)
				delete(h.rooms[event.RelationshipID], client)
			}
		}
	}
}
