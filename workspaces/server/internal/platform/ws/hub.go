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
	Leave(roomId string, conn WebSocketConnection)
	Emit(event string, data ...any)
	EmitTo(roomId string, event string, data ...any)
	readMessage(connection *websocket.Conn) (int, []byte, error)
	Register(conn *websocket.Conn) WebSocketConnection
	Unregister(conn WebSocketConnection)
	GetRoom(conn WebSocketConnection) *string
	parseJson(data []byte) (WsMessage, error)
	handleMessage(eventName string, message WsMessage, conn WebSocketConnection, isHubCall bool)
	isAllowToBroadcast(eventName string) bool
	IsConnectionExist(id string) bool
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
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.rooms[roomId]; !ok {
		w.rooms[roomId] = []WebSocketConnection{}
	}
	w.rooms[roomId] = append(w.rooms[roomId], conn)
}

func (w *wsHub) Leave(roomId string, conn WebSocketConnection) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.leaveInternal(roomId, conn)
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

func (w *wsHub) leaveInternal(roomId string, conn WebSocketConnection) {
	if conns, ok := w.rooms[roomId]; ok {
		for i, c := range conns {
			if c == conn {
				w.rooms[roomId] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
	}
}

func (w *wsHub) getRoomInternal(conn WebSocketConnection) *string {
	for roomId, conns := range w.rooms {
		for _, c := range conns {
			if c == conn {
				return &roomId
			}
		}
	}
	return nil
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
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.getRoomInternal(conn)
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

	if roomId := w.getRoomInternal(conn); roomId != nil {
		w.leaveInternal(*roomId, conn)
	}

	delete(w.clients, conn.ID())
}

func (w *wsHub) IsConnectionExist(id string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, exists := w.clients[id]
	return exists
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
