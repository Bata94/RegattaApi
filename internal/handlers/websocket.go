package handlers

import (
	jsonv2 "encoding/json/v2"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

var (
	hubInstance *Hub
	hubOnce     sync.Once
)

func GetHub() *Hub {
	hubOnce.Do(func() {
		hubInstance = &Hub{
			Clients:    make(map[*Client]bool),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
			Broadcast:  make(chan []byte, 256),
		}
	})
	return hubInstance
}

func RunHub() {
	hub := GetHub()
	for {
		select {
		case client := <-hub.Register:
			hub.mu.Lock()
			hub.Clients[client] = true
			hub.mu.Unlock()
			slog.Debug("WS client registered", "total", len(hub.Clients))

		case client := <-hub.Unregister:
			hub.mu.Lock()
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
			hub.mu.Unlock()
			slog.Debug("WS client unregistered", "total", len(hub.Clients))

		case msg := <-hub.Broadcast:
			hub.mu.RLock()
			for client := range hub.Clients {
				select {
				case client.Send <- msg:
				default:
					slog.Warn("WS client send buffer full, dropping client")
					close(client.Send)
					delete(hub.Clients, client)
				}
			}
			hub.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastJSON(v any) {
	data, err := jsonv2.Marshal(v)
	if err != nil {
		slog.Error("BroadcastJSON marshal error", "err", err)
		return
	}
	h.Broadcast <- data
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}
