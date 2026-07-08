//go:build js && wasm

package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"syscall/js"
	"time"
)

type OpenStart struct {
	ID              int32      `json:"id"`
	RennenNummer    *string    `json:"rennen_nummer,omitempty"`
	StartNummer     *string    `json:"start_nummer,omitempty"`
	TimeClient      *time.Time `json:"time_client,omitempty"`
	TimeServer      *time.Time `json:"time_server,omitempty"`
	MeasuredLatency *int       `json:"measured_latency,omitempty"`
	Verarbeitet     bool       `json:"verarbeitet"`
}

type PendingFinish struct {
	StartNummer     string    `json:"startNummer"`
	TimeClient      time.Time `json:"timeClient"`
	MeasuredLatency int       `json:"measuredLatency"`
	ClientID        string    `json:"clientId"`
	Seq             int       `json:"seq"`
}

type Store struct {
	mu       sync.RWMutex
	onChange func()
}

func NewStore() *Store {
	s := &Store{}
	slog.Info("Store initialized")
	return s
}

func (s *Store) OnChange(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onChange = fn
}

func (s *Store) notify() {
	if s.onChange != nil {
		s.onChange()
	}
}

func getLS(key string) string {
	v := js.Global().Get("localStorage").Call("getItem", key)
	if v.IsNull() || v.IsUndefined() {
		return ""
	}
	return v.String()
}

func setLS(key, value string) {
	js.Global().Get("localStorage").Call("setItem", key, value)
}

func (s *Store) GetOpenStarts() []OpenStart {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getOpenStarts()
}

func (s *Store) getOpenStarts() []OpenStart {
	raw := getLS("openStarts")
	if raw == "" {
		return []OpenStart{}
	}
	var starts []OpenStart
	if err := json.Unmarshal([]byte(raw), &starts); err != nil {
		slog.Error("unmarshal openStarts", "err", err)
		return []OpenStart{}
	}
	if starts == nil {
		return []OpenStart{}
	}
	return starts
}

func (s *Store) SetOpenStarts(starts []OpenStart) {
	s.mu.Lock()
	data, err := json.Marshal(starts)
	if err != nil {
		s.mu.Unlock()
		slog.Error("marshal openStarts", "err", err)
		return
	}
	setLS("openStarts", string(data))
	s.mu.Unlock()
	s.notify()
}

func (s *Store) AddOpenStart(start OpenStart) {
	s.mu.Lock()
	starts := s.getOpenStarts()
	starts = append(starts, start)
	data, _ := json.Marshal(starts)
	setLS("openStarts", string(data))
	s.mu.Unlock()
	s.notify()
}

func (s *Store) RemoveOpenStart(startNummer string) {
	s.mu.Lock()
	starts := s.getOpenStarts()
	filtered := make([]OpenStart, 0, len(starts))
	for _, st := range starts {
		if st.StartNummer != nil && *st.StartNummer == startNummer {
			continue
		}
		filtered = append(filtered, st)
	}
	data, _ := json.Marshal(filtered)
	setLS("openStarts", string(data))
	s.mu.Unlock()
	s.notify()
}

func (s *Store) AddPendingFinish(startNummer string, timeClient time.Time, latencyMs int) PendingFinish {
	s.mu.Lock()

	clientID := s.getOrCreateClientID()
	seq := s.getNextSeq()

	if latencyMs < 0 {
		latencyMs = 0
	}

	pf := PendingFinish{
		StartNummer:     startNummer,
		TimeClient:      timeClient,
		MeasuredLatency: latencyMs,
		ClientID:        clientID,
		Seq:             seq,
	}

	var pending []PendingFinish
	raw := getLS("pendingFinishes")
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &pending); err != nil {
			slog.Error("unmarshal pendingFinishes", "err", err)
		}
	}
	pending = append(pending, pf)
	data, _ := json.Marshal(pending)
	setLS("pendingFinishes", string(data))
	s.mu.Unlock()
	s.notify()
	return pf
}

func (s *Store) GetUnsynced() []PendingFinish {
	s.mu.RLock()
	defer s.mu.RUnlock()
	raw := getLS("pendingFinishes")
	if raw == "" {
		return []PendingFinish{}
	}
	var pending []PendingFinish
	if err := json.Unmarshal([]byte(raw), &pending); err != nil {
		slog.Error("unmarshal pendingFinishes", "err", err)
		return []PendingFinish{}
	}
	if pending == nil {
		return []PendingFinish{}
	}
	return pending
}

func (s *Store) RemovePending(clientID string, seq int) {
	s.mu.Lock()
	raw := getLS("pendingFinishes")
	if raw == "" {
		s.mu.Unlock()
		return
	}
	var pending []PendingFinish
	if err := json.Unmarshal([]byte(raw), &pending); err != nil {
		s.mu.Unlock()
		return
	}
	filtered := make([]PendingFinish, 0, len(pending))
	for _, pf := range pending {
		if pf.ClientID == clientID && pf.Seq == seq {
			continue
		}
		filtered = append(filtered, pf)
	}
	data, _ := json.Marshal(filtered)
	setLS("pendingFinishes", string(data))
	s.mu.Unlock()
}

func (s *Store) RemovePendingByStartNummer(startNummer string) {
	s.mu.Lock()
	raw := getLS("pendingFinishes")
	if raw == "" {
		s.mu.Unlock()
		return
	}
	var pending []PendingFinish
	if err := json.Unmarshal([]byte(raw), &pending); err != nil {
		s.mu.Unlock()
		return
	}
	filtered := make([]PendingFinish, 0, len(pending))
	for _, pf := range pending {
		if pf.StartNummer == startNummer {
			continue
		}
		filtered = append(filtered, pf)
	}
	data, _ := json.Marshal(filtered)
	setLS("pendingFinishes", string(data))
	s.mu.Unlock()
	s.notify()
}

func (s *Store) getOrCreateClientID() string {
	id := getLS("clientId")
	if id == "" {
		id = newUUID()
		setLS("clientId", id)
	}
	return id
}

func (s *Store) getNextSeq() int {
	raw := getLS("seq")
	seq := 0
	if raw != "" {
		json.Unmarshal([]byte(raw), &seq)
	}
	seq++
	data, _ := json.Marshal(seq)
	setLS("seq", string(data))
	return seq
}

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
