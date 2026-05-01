package ws

import (
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func NewWsFiber(app *fiber.App) WsHub {
	hub := NewWsHub()

	// app.Use("/ws", func(c fiber.Ctx) error {
	// 	if websocket.IsWebSocketUpgrade(c) {
	// 		c.Locals("allowed", true)
	// 		return c.Next()
	// 	}
	// 	return fiber.ErrUpgradeRequired
	// })

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		conn := hub.Register(c)

		hub.handleMessage("connect", WsMessage{
			Event: "connect",
			Data:  []interface{}{},
		}, conn, true)

		conn.Emit("connect", "Welcome to the WebSocket server!")

		defer func() {
			// 3. ON DISCONNECT
			hub.Unregister(conn)
			c.Close()

			hub.handleMessage("disconnect", WsMessage{
				Metadata: WsMetadata{
					ID: conn.ID(),
				},
				Event: "disconnect",
				Data:  []interface{}{},
			}, conn, true)
		}()

		var (
			mt  int
			msg []byte
			err error
		)
		for {

			if mt, msg, err = hub.readMessage(c); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					logger.Sugar.Errorf("unexpected read error: %v", err)
				}
				break
			}

			if mt == websocket.TextMessage {
				parsedMsg, err := hub.parseJson(msg)

				if err != nil {
					logger.Sugar.Errorf("error parsing message: %v", err)
					continue
				}

				hub.handleMessage(parsedMsg.Event, parsedMsg, conn, false)
			}

		}
	}))
	return hub
}
