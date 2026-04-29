package ws

import "github.com/gofiber/websocket/v2"

type Hub struct {
	clients map[*websocket.Conn]bool
	Room    map[string]map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
		Room:    make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *Hub) Add(c *websocket.Conn, room *string) {

	if room != nil {
		if _, ok := h.Room[*room]; !ok {
			h.Room[*room] = make(map[*websocket.Conn]bool)
		}
	}
	h.Room[*room][c] = true
}

func (h *Hub) Remove(c *websocket.Conn) {
	delete(h.clients, c)
	for room := range h.Room {
		delete(h.Room[room], c)
		if len(h.Room[room]) == 0 {
			delete(h.Room, room)
		}
	}
}

func (h *Hub) Broadcast(msg []byte, room *string) {

	for client := range h.clients {
		if room != nil {
			if _, ok := h.Room[*room][client]; ok {
				client.WriteMessage(websocket.TextMessage, msg)
			}
		} else {
			client.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
