package handlers

import (
	"fmt"
	"time"

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

func sendPing(conn *websocket.Conn, interval int) {
	pingPeriod := time.Duration(interval) * time.Second
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()
	time.Sleep(pingPeriod) // wait a while before start ping,
	for {
		select {
		case <-pingTicker.C:
			pingMsg := fmt.Sprint("ping ", time.Now().Format("2006-01-02_15-04-05.999"), " UTC")
			log.Debug(pingMsg)
			err := conn.WriteControl(websocket.PingMessage, []byte(pingMsg), time.Now().Add(time.Second*5))
			if err != nil {
				log.Error(err)
			}
			err = conn.WriteMessage(websocket.TextMessage, []byte(pingMsg))
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func RunHub() {
	for {
		select {
		case connection := <-Register:
			// go sendPing(connection, 2)

			Clients[connection] = Client{}
			log.Debug("connection registered")
			// Broadcast <- "New Client"

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
