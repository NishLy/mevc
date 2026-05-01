package ws

import (
	"fmt"
	"sync"
)

type WsMetadata struct {
	ID     string  `json:"id"`
	RoomId *string `json:"roomId,omitempty"`
}

type WsMessage struct {
	Metadata WsMetadata    `json:"metadata"`
	Event    string        `json:"event"`
	Data     []interface{} `json:"data"`
}

type RawConn interface {
	WriteJSON(v interface{}) error
	Close() error
}

type WebSocketConnection interface {
	RawConn // Embed your local interface
	ID() string
	Emit(event string, data ...interface{}) error
	GetGroupId() *string
}

type webSocketConnection struct {
	conn    RawConn // Use the interface type here
	id      string
	GroupId string
	mu      sync.Mutex
}

func NewWebsocketConn(conn RawConn, id string) WebSocketConnection {
	return &webSocketConnection{
		conn:    conn,
		id:      id,
		GroupId: "",
		mu:      sync.Mutex{},
	}
}

func (w *webSocketConnection) GetGroupId() *string {
	if w.GroupId == "" {
		return nil
	}
	return &w.GroupId
}

func (w *webSocketConnection) Emit(event string, data ...interface{}) error {
	return w.conn.WriteJSON(map[string]interface{}{
		"metadata": map[string]interface{}{
			"id": w.id,
		},
		"event": event,
		"data":  data,
	})
}

func (w *webSocketConnection) ID() string {
	return w.id
}

func (w *webSocketConnection) Close() error {
	return w.conn.Close()
}

func (w *webSocketConnection) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

func generateWsId(seed int) string {
	return fmt.Sprintf("ws_%d", seed)
}
