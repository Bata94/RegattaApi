//go:build js && wasm

package main

import (
	"encoding/json"
	"log/slog"
	"sync"
	"syscall/js"
	"time"
)

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type WSClient struct {
	url              string
	store            *Store
	ws               js.Value
	onSendFn         func(PendingFinish)
	shouldReconnect  bool
	reconnectBackoff time.Duration
	connected        bool
	latencyMs        int64
	reconnecting     bool
	pingStarted      bool
	pingIntervalSec  int
	onChange         func()
	mu               sync.RWMutex
}

func NewWSClient(url string, store *Store) *WSClient {
	return &WSClient{
		url:              url,
		store:            store,
		shouldReconnect:  true,
		reconnectBackoff: time.Second,
		latencyMs:        -1,
		pingIntervalSec:  1,
	}
}

func (c *WSClient) OnSend(fn func(PendingFinish)) {
	c.onSendFn = fn
}

func (c *WSClient) OnChange(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onChange = fn
}

func (c *WSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *WSClient) notify() {
	c.mu.RLock()
	fn := c.onChange
	c.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

func (c *WSClient) GetLatencyMs() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latencyMs
}

func (c *WSClient) IsReconnecting() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reconnecting
}

func (c *WSClient) ForceReconnect() {
	c.mu.Lock()
	c.shouldReconnect = true
	c.reconnecting = true
	ws := c.ws
	c.mu.Unlock()
	c.notify()
	if !ws.IsNull() && !ws.IsUndefined() {
		ws.Call("close")
	}
}

func (c *WSClient) sendPing() {
	now := time.Now().UnixMilli()
	c.mu.RLock()
	ws := c.ws
	c.mu.RUnlock()
	if ws.IsNull() || ws.IsUndefined() {
		return
	}
	if ws.Get("readyState").Int() != 1 {
		return
	}
	data, err := json.Marshal(map[string]any{
		"type": "ping",
		"data": map[string]any{
			"t": now,
		},
	})
	if err != nil {
		slog.Error("marshal ping", "err", err)
		return
	}
	ws.Call("send", string(data))
}

func (c *WSClient) startPingLoop() {
	c.mu.Lock()
	if c.pingStarted {
		c.mu.Unlock()
		return
	}
	c.pingStarted = true
	c.mu.Unlock()
	go func() {
		for {
			time.Sleep(time.Duration(c.pingIntervalSec) * time.Second)
			c.mu.RLock()
			connected := c.connected
			c.mu.RUnlock()
			if !connected {
				continue
			}
			c.sendPing()
		}
	}()
}

func (c *WSClient) Connect() {
	c.mu.Lock()
	ws := js.Global().Get("WebSocket").New(c.url)
	c.ws = ws
	c.mu.Unlock()

	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) any {
		slog.Info("WS connected")
		c.mu.Lock()
		c.reconnectBackoff = time.Second
		c.connected = true
		c.reconnecting = false
		c.mu.Unlock()
		c.notify()
		c.sendPing()
		c.flushUnsynced()
		return nil
	}))

	ws.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		raw := event.Get("data").String()
		c.handleMessage(raw)
		return nil
	}))

	ws.Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) any {
		slog.Warn("WS disconnected")
		c.mu.Lock()
		c.connected = false
		shouldReconnect := c.shouldReconnect
		if shouldReconnect {
			c.reconnecting = true
		}
		c.mu.Unlock()
		c.notify()
		if shouldReconnect {
			go c.reconnectLoop()
		}
		return nil
	}))

	ws.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		slog.Error("WS error")
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		c.notify()
		return nil
	}))

	c.startPingLoop()
}

func (c *WSClient) reconnectLoop() {
	c.mu.Lock()
	backoff := c.reconnectBackoff
	if c.reconnectBackoff < 30*time.Second {
		c.reconnectBackoff = c.reconnectBackoff * 2
	}
	c.mu.Unlock()

	time.Sleep(backoff)

	slog.Info("WS reconnecting...", "backoff", backoff)
	c.Connect()
}

func (c *WSClient) handleMessage(raw string) {
	var msg WSMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		slog.Error("WS unmarshal message", "err", err)
		return
	}

	switch msg.Type {
	case "snapshot":
		var data struct {
			OpenStarts []OpenStart `json:"openStarts"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal snapshot", "err", err)
			return
		}
		c.store.SetOpenStarts(data.OpenStarts)
		slog.Info("snapshot received", "count", len(data.OpenStarts))

	case "new_start":
		var starts []OpenStart
		if err := json.Unmarshal(msg.Data, &starts); err != nil {
			slog.Error("unmarshal new_start", "err", err)
			return
		}
		for _, s := range starts {
			c.store.AddOpenStart(s)
		}

	case "finish_confirmed":
		var data struct {
			StartNummer string  `json:"startNummer"`
			Endzeit     float64 `json:"endzeit"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal finish_confirmed", "err", err)
			return
		}
		c.store.RemoveOpenStart(data.StartNummer)
		c.store.RemovePendingByStartNummer(data.StartNummer)
		slog.Info("finish confirmed", "startNummer", data.StartNummer, "endzeit", data.Endzeit)

	case "pong":
		var data struct {
			T int64 `json:"t"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal pong", "err", err)
			return
		}
		c.mu.Lock()
		c.latencyMs = time.Now().UnixMilli() - data.T
		c.mu.Unlock()
		c.notify()

	case "finish_unmatched":
		var data struct {
			StartNummer string `json:"startNummer"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal finish_unmatched", "err", err)
			return
		}
		slog.Warn("finish unmatched", "startNummer", data.StartNummer)
		c.store.RemovePendingByStartNummer(data.StartNummer)

	default:
		slog.Warn("unknown WS message type", "type", msg.Type)
	}
}

func (c *WSClient) SendFinish(pf PendingFinish) {
	if c.ws.IsNull() || c.ws.IsUndefined() {
		slog.Warn("WS not connected, queued locally", "startNummer", pf.StartNummer)
		return
	}
	if c.ws.Get("readyState").Int() != 1 {
		slog.Warn("WS not open, queued locally", "startNummer", pf.StartNummer, "readyState", c.ws.Get("readyState").Int())
		return
	}
	data, _ := json.Marshal(map[string]any{
		"type": "submit_finish",
		"data": pf,
	})
	c.ws.Call("send", string(data))
}

func (c *WSClient) flushUnsynced() {
	pending := c.store.GetUnsynced()
	if len(pending) == 0 {
		return
	}
	slog.Info("flushing unsynced finishes", "count", len(pending))
	for _, pf := range pending {
		c.SendFinish(pf)
	}
}
