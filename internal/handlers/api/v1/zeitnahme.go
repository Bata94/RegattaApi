package api_v1

import (
	"strconv"
	"time"

	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
)

func WsTestHandler(c *websocket.Conn) {
	// When the function returns, unregister the client and close the connection
	defer func() {
		handlers.Unregister <- c
		c.Close()
	}()

	// Register the client
	handlers.Register <- c

	for {
		// Set a ping interval
		err := c.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(time.Second*10))
		if err != nil {
			log.Debug("ping error:", err)
			break
		}

		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("read error:", err)
			}

			return // Calls the deferred function, i.e. closes the connection on error
		}

		if messageType == websocket.TextMessage {
			// Broadcast the received message
			handlers.Broadcast <- strconv.Itoa(messageType) + ": " + string(message[:])
		} else {
			log.Error("websocket message received of type", messageType)
		}
	}
}
