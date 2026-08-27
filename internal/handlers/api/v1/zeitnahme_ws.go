package api_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkWebSocketOrigin,
}

func checkWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	allowedOrigins := strings.Split(config.C.CORS.AllowedOrigins, ",")
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			prefix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, prefix) {
				return true
			}
			continue
		}
		if allowed == origin {
			return true
		}
	}

	if strings.HasSuffix(origin, ".localhost") || strings.HasSuffix(origin, "localhost") {
		return true
	}

	slog.Warn("WebSocket origin rejected", "origin", origin, "allowed", allowedOrigins)
	return false
}

func validateWebSocketToken(r *http.Request) error {
	tokenString := getWSToken(r)
	if tokenString == "" {
		return fmt.Errorf("missing token")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.C.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token: %w", err)
	}

	return nil
}

func getWSToken(r *http.Request) string {
	c, err := r.Cookie("auth_token")
	if err == nil && c != nil {
		return c.Value
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

type wsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type submitFinishPayload struct {
	StartNummer     string    `json:"startNummer"`
	TimeClient      time.Time `json:"timeClient"`
	MeasuredLatency int       `json:"measuredLatency"`
	ClientID        string    `json:"clientId"`
	Seq             string    `json:"seq"`
}

type submitStartPayload struct {
	RennenNummer    string    `json:"rennenNummer"`
	StartNummern    []string  `json:"startNummern"`
	TimeClient      time.Time `json:"timeClient"`
	MeasuredLatency int       `json:"measuredLatency"`
	ClientID        string    `json:"clientId"`
	Seq             string    `json:"seq"`
}

type assignFinishPayload struct {
	ClientID    string `json:"clientId"`
	Seq         string `json:"seq"`
	ZielID      int32  `json:"zielId"`
	StartNummer string `json:"startNummer"`
}

func HandleZeitnahmeWS(w http.ResponseWriter, r *http.Request) {
	if err := validateWebSocketToken(r); err != nil {
		slog.Warn("WS auth failed", "err", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WS upgrade failed", "err", err)
		return
	}

	hub := handlers.GetHub()
	client := &handlers.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	hub.Register <- client

	if err := sendSnapshot(conn); err != nil {
		slog.Error("sendSnapshot failed", "err", err)
	}

	go writePump(client)
	readPump(client)
}

func sendSnapshot(conn *websocket.Conn) error {
	ctx := context.TODO()
	starts, err := crud.GetOpenZeitnahmeStart(ctx)
	if err != nil {
		slog.Error("snapshot: GetOpenZeitnahmeStart failed", "err", err)
		return conn.WriteJSON(map[string]any{"type": "error", "message": "internal error"})
	}

	return conn.WriteJSON(map[string]any{
		"type": "snapshot",
		"data": map[string]any{
			"openStarts": starts,
		},
	})
}

func writePump(client *handlers.Client) {
	defer func() {
		client.Hub.Unregister <- client
		if err := client.Conn.Close(); err != nil {
			slog.Warn("WS Conn.Close error", "err", err)
		}
	}()

	for msg := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			slog.Error("WS write error", "err", err)
			return
		}
	}
}

func readPump(client *handlers.Client) {
	defer func() {
		client.Hub.Unregister <- client
		if err := client.Conn.Close(); err != nil {
			slog.Warn("WS Conn.Close error", "err", err)
		}
	}()

	for {
		_, raw, err := client.Conn.ReadMessage()
		if err != nil {
			slog.Debug("WS read error", "err", err)
			return
		}

		var msg wsMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Error("WS unmarshal error", "err", err)
			continue
		}

		switch msg.Type {
		case "record_finish":
			handleRecordFinish(client, msg.Data)
		case "assign_finish":
			handleAssignFinish(client, msg.Data)
		case "record_start":
			handleRecordStart(client, msg.Data)
		case "ping":
			handlePing(client, msg.Data)
		default:
			slog.Debug("WS unknown message type", "type", msg.Type)
		}
	}
}

func handlePing(client *handlers.Client, raw json.RawMessage) {
	var data struct {
		T int64 `json:"t"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		slog.Error("WS ping unmarshal error", "err", err)
		return
	}
	pong, err := json.Marshal(map[string]any{
		"type": "pong",
		"data": map[string]any{
			"t": data.T,
		},
	})
	if err != nil {
		slog.Error("WS pong marshal error", "err", err)
		return
	}
	select {
	case client.Send <- pong:
	default:
		slog.Warn("WS pong send buffer full")
	}
}

func handleRecordFinish(client *handlers.Client, raw json.RawMessage) {
	var p submitFinishPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		slog.Error("WS record_finish unmarshal error", "err", err)
		return
	}

	ctx := context.TODO()
	zn, err := crud.CreateZeitnahmeZiel(ctx, nil, nil, p.TimeClient, p.MeasuredLatency, p.ClientID, p.Seq)
	if err != nil {
		slog.Error("CreateZeitnahmeZiel failed", "err", err)
		return
	}

	msg, err := json.Marshal(map[string]any{
		"type": "finish_recorded",
		"data": map[string]any{
			"clientId":   p.ClientID,
			"seq":        p.Seq,
			"zielId":     zn.ID,
			"timeClient": zn.TimeClient,
		},
	})
	if err != nil {
		slog.Error("finish_recorded marshal error", "err", err)
		return
	}

	select {
	case client.Send <- msg:
	default:
		slog.Warn("WS finish_recorded send buffer full")
	}
}

func handleRecordStart(client *handlers.Client, raw json.RawMessage) {
	var p submitStartPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		slog.Error("WS record_start unmarshal error", "err", err)
		return
	}
	if len(p.StartNummern) == 0 {
		slog.Warn("WS record_start: no start numbers provided")
		return
	}

	var rennNr *string
	if p.RennenNummer != "" {
		rennNr = &p.RennenNummer
	}

	ctx := context.TODO()
	starts, err := crud.CreateZeitnahmeStart(ctx, rennNr, p.StartNummern, p.TimeClient, p.MeasuredLatency, p.ClientID, p.Seq)
	if err != nil {
		slog.Error("CreateZeitnahmeStart failed", "err", err)
		return
	}

	msg, err := json.Marshal(map[string]any{
		"type": "start_recorded",
		"data": map[string]any{
			"clientId": p.ClientID,
			"seq":      p.Seq,
			"starts":   starts,
		},
	})
	if err != nil {
		slog.Error("start_recorded marshal error", "err", err)
		return
	}

	select {
	case client.Send <- msg:
	default:
		slog.Warn("WS start_recorded send buffer full")
	}

	client.Hub.BroadcastJSON(map[string]any{
		"type": "new_start",
		"data": starts,
	})
}

func handleAssignFinish(client *handlers.Client, raw json.RawMessage) {
	var p assignFinishPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		slog.Error("WS assign_finish unmarshal error", "err", err)
		return
	}

	ctx := context.TODO()
	if _, err := crud.UpdateZeitnahmeZiel(ctx, p.ZielID, nil, &p.StartNummer); err != nil {
		slog.Error("UpdateZeitnahmeZiel failed", "err", err)
		return
	}

	hub := client.Hub
	// TODO: filter by rennen_nummer to avoid loading all open starts (issue #22)
	openStarts, err := crud.GetOpenZeitnahmeStart(ctx)
	if err != nil {
		slog.Error("GetOpenZeitnahmeStart for matching failed", "err", err)
		hub.BroadcastJSON(map[string]any{
			"type": "finish_unmatched",
			"data": map[string]any{
				"clientId":    p.ClientID,
				"seq":         p.Seq,
				"startNummer": p.StartNummer,
				"id":          p.ZielID,
			},
		})
		return
	}

	for _, s := range openStarts {
		if s.StartNummer != nil && *s.StartNummer == p.StartNummer {
			startNummerInt, atoiErr := strconv.Atoi(*s.StartNummer)
			if atoiErr != nil {
				slog.Error("StartNummer Atoi failed", "err", atoiErr)
				continue
			}

			meld, meldErr := crud.GetMeldungByStartNrUndTag(ctx, startNummerInt, crud.Tag(config.C.Zeitnahme.GetCurrentTag()))
			if meldErr != nil {
				slog.Debug("GetMeldungByStartNrUndTag failed", "err", meldErr)
				continue
			}
			if meld.Uuid == uuid.Nil {
				slog.Warn("Meldung not found for start number", "startNummer", p.StartNummer)
				continue
			}

			zn, getErr := crud.GetZeitnahmeZiel(ctx, int(p.ZielID))
			if getErr != nil {
				slog.Error("GetZeitnahmeZiel failed", "err", getErr)
				continue
			}

			err = crud.CreateZeitnahmeErgebnis(ctx, s, zn, meld)
			if err != nil {
				slog.Error("CreateZeitnahmeErgebnis failed", "err", err)
				continue
			}

			hub.BroadcastJSON(map[string]any{
				"type": "finish_confirmed",
				"data": map[string]any{
					"clientId":    p.ClientID,
					"seq":         p.Seq,
					"startNummer": p.StartNummer,
					"id":          p.ZielID,
					"endzeit":     zn.TimeClient.Sub(*s.TimeClient).Seconds(),
				},
			})
			return
		}
	}

	hub.BroadcastJSON(map[string]any{
		"type": "finish_unmatched",
		"data": map[string]any{
			"clientId":    p.ClientID,
			"seq":         p.Seq,
			"startNummer": p.StartNummer,
			"id":          p.ZielID,
		},
	})
}
