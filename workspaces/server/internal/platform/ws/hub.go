package ws

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/v3/websocket"
)

type HandlerFunc func(conn WebSocketConnection, data ...any)

type WsHub interface {
	On(event string, callback HandlerFunc)
	Join(roomId string, conn WebSocketConnection)
	Emit(event string, data ...any)
	EmitTo(roomId string, event string, data ...any)
	readMessage(connection *websocket.Conn) (int, []byte, error)
	Register(conn *websocket.Conn) WebSocketConnection
	Unregister(conn WebSocketConnection)
	GetRoom(conn WebSocketConnection) *string
	parseJson(data []byte) (WsMessage, error)
	handleMessage(eventName string, message WsMessage, conn WebSocketConnection, isHubCall bool)
	isAllowToBroadcast(eventName string) bool
}

type wsHub struct {
	currentId int
	clients   map[string]WebSocketConnection
	rooms     map[string][]WebSocketConnection
	listeners map[string][]HandlerFunc
	mu        sync.RWMutex
}

func NewWsHub() WsHub {
	return &wsHub{
		clients:   make(map[string]WebSocketConnection),
		rooms:     make(map[string][]WebSocketConnection),
		listeners: make(map[string][]HandlerFunc),
		mu:        sync.RWMutex{},
	}
}

func (w *wsHub) On(event string, callback HandlerFunc) {
	if _, ok := w.listeners[event]; !ok {
		w.listeners[event] = []HandlerFunc{}
	}
	w.listeners[event] = append(w.listeners[event], callback)
}

func (w *wsHub) Join(roomId string, conn WebSocketConnection) {
	if _, ok := w.rooms[roomId]; !ok {
		w.rooms[roomId] = []WebSocketConnection{}
	}
	w.rooms[roomId] = append(w.rooms[roomId], conn)
}

func (w *wsHub) Emit(event string, data ...any) {
	if !w.isAllowedToEmit(event) {
		return
	}

	for _, conn := range w.clients {
		conn.Emit(event, data...)
	}
}

func (w *wsHub) EmitTo(roomId string, event string, data ...any) {
	if !w.isAllowedToEmit(event) {
		return
	}
	if conns, ok := w.rooms[roomId]; ok {
		for _, conn := range conns {
			conn.Emit(event, data...)
		}
	}
}

func (w *wsHub) readMessage(connection *websocket.Conn) (int, []byte, error) {
	return connection.ReadMessage()
}

func (w *wsHub) parseJson(data []byte) (WsMessage, error) {
	var parsedData WsMessage
	err := json.Unmarshal(data, &parsedData)
	return parsedData, err
}

func (w *wsHub) handleMessage(eventName string, message WsMessage, conn WebSocketConnection, isHubCall bool) {
	if !w.isAllowToBroadcast(eventName) && !isHubCall {
		return
	}

	if listeners, ok := w.listeners[eventName]; ok {
		for _, listener := range listeners {
			listener(conn, message.Data...)
		}
	}
}

func (w *wsHub) GetRoom(conn WebSocketConnection) *string {
	for roomId, conns := range w.rooms {
		for _, c := range conns {
			if c == conn {
				return &roomId
			}
		}
	}
	return nil
}

func (w *wsHub) Register(conn *websocket.Conn) WebSocketConnection {
	w.mu.Lock()
	defer w.mu.Unlock()
	id := generateWsId(w.currentId)
	newConn := NewWebsocketConn(conn, id)
	w.clients[id] = newConn
	w.currentId++
	return newConn
}

func (w *wsHub) Unregister(conn WebSocketConnection) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.clients, conn.ID())
}

func (w *wsHub) isAllowToBroadcast(eventName string) bool {
	var reversedNames = []string{"connect", "disconnect", "error"}
	for _, name := range reversedNames {
		if name == eventName {
			return false
		}
	}
	return true
}

func (w *wsHub) isAllowedToEmit(eventName string) bool {
	var reversedNames = []string{"connected", "disconnected", "error"}
	for _, name := range reversedNames {
		if name == eventName {
			return false
		}
	}
	return true
}
