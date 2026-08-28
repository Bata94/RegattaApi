//go:build js && wasm

package main

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"log/slog"
	"sync"
	"syscall/js"
	"time"
)

type WSMessage struct {
	Type string         `json:"type"`
	Data jsontext.Value `json:"data,omitempty"`
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
	pingDone         chan struct{}
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
	if c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.shouldReconnect = true
	c.reconnecting = true
	ws := c.ws
	if !ws.IsNull() && !ws.IsUndefined() {
		ws.Call("close")
	}
	c.mu.Unlock()
	c.notify()
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
	data, err := jsonv2.Marshal(map[string]any{
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
	if c.pingDone == nil {
		c.pingDone = make(chan struct{})
	}
	c.mu.Unlock()
	go func() {
		for {
			select {
			case <-c.pingDone:
				return
			case <-time.After(time.Duration(c.pingIntervalSec) * time.Second):
			}
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
		if !shouldReconnect && c.pingDone != nil {
			close(c.pingDone)
			c.pingDone = nil
		}
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
	if !c.reconnecting {
		c.mu.Unlock()
		return
	}
	backoff := c.reconnectBackoff
	if c.reconnectBackoff < 30*time.Second {
		c.reconnectBackoff = c.reconnectBackoff * 2
	}
	c.mu.Unlock()

	time.Sleep(backoff)

	c.mu.Lock()
	if !c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	slog.Info("WS reconnecting...", "backoff", backoff)
	c.Connect()
}

func (c *WSClient) handleMessage(raw string) {
	var msg WSMessage
	if err := jsonv2.Unmarshal([]byte(raw), &msg); err != nil {
		slog.Error("WS unmarshal message", "err", err)
		return
	}

	switch msg.Type {
	case "snapshot":
		var data struct {
			OpenStarts []OpenStart `json:"openStarts"`
		}
		if err := jsonv2.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal snapshot", "err", err)
			return
		}
		c.store.SetOpenStarts(data.OpenStarts)
		slog.Info("snapshot received", "count", len(data.OpenStarts))

	case "new_start":
		var starts []OpenStart
		if err := jsonv2.Unmarshal(msg.Data, &starts); err != nil {
			slog.Error("unmarshal new_start", "err", err)
			return
		}
		for _, s := range starts {
			c.store.AddOpenStart(s)
		}

	case "finish_recorded":
		var data struct {
			ClientID string `json:"clientId"`
			Seq      string `json:"seq"`
			ZielID   int32  `json:"zielId"`
		}
		if err := jsonv2.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal finish_recorded", "err", err)
			return
		}
		c.store.SetPendingZielID(data.ClientID, data.Seq, data.ZielID)
		slog.Info("finish recorded", "zielId", data.ZielID, "clientId", data.ClientID, "seq", data.Seq)
		if pf := c.store.GetPending(data.ClientID, data.Seq); pf != nil && pf.StartNummer != "" && !pf.Unmatched {
			c.SendAssignFinish(*pf)
		}

	case "start_recorded":
		var data struct {
			ClientID string      `json:"clientId"`
			Seq      string      `json:"seq"`
			Starts   []OpenStart `json:"starts"`
		}
		if err := jsonv2.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal start_recorded", "err", err)
			return
		}
		ids := make([]int32, 0, len(data.Starts))
		for _, st := range data.Starts {
			ids = append(ids, st.ID)
		}
		c.store.SetPendingStartSynced(data.ClientID, data.Seq, ids)
		slog.Info("start recorded", "count", len(data.Starts), "clientId", data.ClientID, "seq", data.Seq)

	case "finish_confirmed":
		var data struct {
			ClientID    string  `json:"clientId"`
			Seq         string  `json:"seq"`
			StartNummer string  `json:"startNummer"`
			Endzeit     float64 `json:"endzeit"`
		}
		if err := jsonv2.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal finish_confirmed", "err", err)
			return
		}
		c.store.RemoveOpenStart(data.StartNummer)
		c.store.RemovePending(data.ClientID, data.Seq)
		slog.Info("finish confirmed", "startNummer", data.StartNummer, "endzeit", data.Endzeit)

	case "pong":
		var data struct {
			T int64 `json:"t"`
		}
		if err := jsonv2.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal pong", "err", err)
			return
		}
		c.mu.Lock()
		c.latencyMs = time.Now().UnixMilli() - data.T
		c.mu.Unlock()
		c.notify()

	case "finish_unmatched":
		var data struct {
			ClientID    string `json:"clientId"`
			Seq         string `json:"seq"`
			StartNummer string `json:"startNummer"`
		}
		if err := jsonv2.Unmarshal(msg.Data, &data); err != nil {
			slog.Error("unmarshal finish_unmatched", "err", err)
			return
		}
		slog.Warn("finish unmatched", "startNummer", data.StartNummer)
		c.store.MarkUnmatched(data.ClientID, data.Seq)

	default:
		slog.Warn("unknown WS message type", "type", msg.Type)
	}
}

func (c *WSClient) sendMessage(data map[string]any) {
	c.mu.RLock()
	ws := c.ws
	c.mu.RUnlock()
	if ws.IsNull() || ws.IsUndefined() {
		slog.Error("WS not connected, message dropped", "type", data["type"])
		return
	}
	if ws.Get("readyState").Int() != 1 {
		slog.Error("WS not open, message dropped", "type", data["type"], "readyState", ws.Get("readyState").Int())
		return
	}
	raw, err := jsonv2.Marshal(data)
	if err != nil {
		slog.Error("marshal WS message", "err", err)
		return
	}
	ws.Call("send", string(raw))
}

func (c *WSClient) SendRecordFinish(pf PendingFinish) {
	if pf.ZielID != nil {
		return
	}
	c.sendMessage(map[string]any{
		"type": "record_finish",
		"data": pf,
	})
}

func (c *WSClient) SendAssignFinish(pf PendingFinish) {
	if pf.ZielID == nil || pf.StartNummer == "" {
		return
	}
	c.sendMessage(map[string]any{
		"type": "assign_finish",
		"data": pf,
	})
}

func (c *WSClient) SendRecordStart(ps PendingStart) {
	if ps.StartIDs != nil {
		return
	}
	c.sendMessage(map[string]any{
		"type": "record_start",
		"data": ps,
	})
}

func (c *WSClient) flushUnsynced() {
	pending := c.store.GetPendingFinishes()
	if len(pending) == 0 {
		return
	}
	slog.Info("flushing unsynced finishes", "count", len(pending))
	for _, pf := range pending {
		if pf.ZielID == nil {
			c.SendRecordFinish(pf)
		} else if pf.StartNummer != "" && !pf.Unmatched {
			c.SendAssignFinish(pf)
		}
	}

	starts := c.store.GetPendingStarts()
	if len(starts) == 0 {
		return
	}
	slog.Info("flushing unsynced starts", "count", len(starts))
	for _, ps := range starts {
		if ps.StartIDs == nil {
			c.SendRecordStart(ps)
		}
	}
}
