package ws

import "github.com/gofiber/websocket/v2"

func Handler(hub *Hub) func(*websocket.Conn) {

	return func(c *websocket.Conn) {

		hub.Add(c, nil)
		defer hub.Remove(c)

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}

			hub.Broadcast(msg, nil)
		}
	}
}
