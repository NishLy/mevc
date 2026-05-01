package ws

import (
	"log"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func NewWsFiber(app *fiber.App) WsHub {
	hub := NewWsHub()

	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		conn := hub.Register(c)

		hub.handleMessage("connect", WsMessage{
			Event: "connect",
			Data:  []interface{}{},
		}, conn)

		defer func() {
			// 3. ON DISCONNECT
			hub.Unregister(conn)
			c.Close()

			hub.handleMessage("disconnect", WsMessage{
				Event: "disconnect",
				Data:  []interface{}{},
			}, conn)
		}()

		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = hub.readMessage(c); err != nil {
				log.Println("read:", err)
				break
			}

			if mt == websocket.TextMessage {
				parsedMsg, err := hub.parseJson(msg)

				if err != nil {
					log.Println("error parsing message:", err)
					continue
				}

				hub.handleMessage(parsedMsg.Event, parsedMsg, conn)
			}

			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))
	return hub
}
