package handlers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type Client struct{} // Add more data to this type if needed

var Clients = make(map[*websocket.Conn]Client) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
var Register = make(chan *websocket.Conn)
var Broadcast = make(chan string)
var Unregister = make(chan *websocket.Conn)

func WsUpgrade(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func RunHub() {
	for {
		select {
		case connection := <-Register:
			Clients[connection] = Client{}
			log.Debug("connection registered")

		case message := <-Broadcast:
			log.Debug("message received:", message)

			// Send the message to all clients
			for connection := range Clients {
				if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Debug("write error:", err)

					Unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-Unregister:
			// Remove the client from the hub
			delete(Clients, connection)

			log.Debug("connection unregistered")
		}
	}
}
