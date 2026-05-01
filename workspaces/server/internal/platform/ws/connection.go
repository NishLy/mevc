package ws

import "crypto"

type WsMessage struct {
	Event string        `json:"event"`
	Data  []interface{} `json:"data"`
}

type RawConn interface {
	WriteJSON(v interface{}) error
	Close() error
}

type WebSocketConnection interface {
	RawConn // Embed your local interface
	ID() string
	Emit(event string, data ...interface{}) error
}

type webSocketConnection struct {
	conn RawConn // Use the interface type here
	id   string
}

func NewWebsocketConn(conn RawConn, id string) WebSocketConnection {
	return &webSocketConnection{
		conn: conn,
		id:   id,
	}
}

func (w *webSocketConnection) Emit(event string, data ...interface{}) error {
	return w.conn.WriteJSON(map[string]interface{}{
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
	return w.conn.WriteJSON(v)
}

func generateWsId(seed int) string {
	hash := crypto.SHA256.New()
	hash.Write([]byte(string(seed)))
	return string(hash.Sum(nil))
}
