//go:build js && wasm

package main

import (
	jsonv2 "encoding/json/v2"
	"log/slog"
	"syscall/js"
	"time"
)

func main() {
	slog.Info("WASM Zeitnahme module starting")

	store := NewStore()

	loc := js.Global().Get("window").Get("location")
	protocol := "ws"
	if loc.Get("protocol").String() == "https:" {
		protocol = "wss"
	}
	wsURL := protocol + "://" + loc.Get("host").String() + "/ws/zeitnahme"

	ws := NewWSClient(wsURL, store)
	ws.Connect()

	js.Global().Set("__wasm_recordFinish", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			slog.Warn("__wasm_recordFinish: need 1 arg")
			return nil
		}
		timeStr := args[0].String()

		timeClient, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			timeClient = time.Now()
		}

		pf := store.AddPendingFinish(timeClient, int(ws.GetLatencyMs()))
		ws.SendRecordFinish(pf)
		return nil
	}))

	js.Global().Set("__wasm_assignFinish", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 3 {
			slog.Warn("__wasm_assignFinish: need 3 args")
			return nil
		}
		clientID := args[0].String()
		seq := args[1].String()
		startNr := args[2].String()

		store.AssignStartNummer(clientID, seq, startNr)
		if pf := store.GetPending(clientID, seq); pf != nil {
			ws.SendAssignFinish(*pf)
		}
		return nil
	}))

	js.Global().Set("__wasm_removePending", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			slog.Warn("__wasm_removePending: need 2 args")
			return nil
		}
		clientID := args[0].String()
		seq := args[1].String()
		store.RemovePending(clientID, seq)
		return nil
	}))

	js.Global().Set("__wasm_recordStart", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 3 {
			slog.Warn("__wasm_recordStart: need 3 args")
			return false
		}
		rennNr := args[0].String()
		var startNummern []string
		if err := jsonv2.Unmarshal([]byte(args[1].String()), &startNummern); err != nil {
			slog.Error("__wasm_recordStart: unmarshal startNummern", "err", err)
			return false
		}
		timeStr := args[2].String()
		timeClient, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			timeClient = time.Now()
		}

		ps := store.AddPendingStart(rennNr, startNummern, timeClient, int(ws.GetLatencyMs()))
		ws.SendRecordStart(ps)
		return true
	}))

	js.Global().Set("__wasm_getPendingStarts", js.FuncOf(func(this js.Value, args []js.Value) any {
		pending := store.GetPendingStarts()
		data, err := jsonv2.Marshal(pending)
		if err != nil {
			slog.Error("marshal pendingStarts", "err", err)
			return "[]"
		}
		return string(data)
	}))

	js.Global().Set("__wasm_removePendingStart", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			slog.Warn("__wasm_removePendingStart: need 2 args")
			return nil
		}
		clientID := args[0].String()
		seq := args[1].String()
		store.RemovePendingStart(clientID, seq)
		return nil
	}))

	js.Global().Set("__wasm_getUnsyncedStartCount", js.FuncOf(func(this js.Value, args []js.Value) any {
		return len(store.GetUnsyncedStarts())
	}))

	js.Global().Set("__wasm_getPendingFinishes", js.FuncOf(func(this js.Value, args []js.Value) any {
		pending := store.GetPendingFinishes()
		data, err := jsonv2.Marshal(pending)
		if err != nil {
			slog.Error("marshal pendingFinishes", "err", err)
			return "[]"
		}
		return string(data)
	}))

	js.Global().Set("__wasm_getOpenStarts", js.FuncOf(func(this js.Value, args []js.Value) any {
		starts := store.GetOpenStarts()
		data, err := jsonv2.Marshal(starts)
		if err != nil {
			slog.Error("marshal openStarts", "err", err)
			return "[]"
		}
		return string(data)
	}))

	js.Global().Set("__wasm_onStateChange", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		cb := args[0]
		store.OnChange(func() {
			cb.Invoke()
		})
		return nil
	}))

	js.Global().Set("__wasm_getUnsyncedCount", js.FuncOf(func(this js.Value, args []js.Value) any {
		return len(store.GetUnsynced())
	}))

	js.Global().Set("__wasm_getWSConnected", js.FuncOf(func(this js.Value, args []js.Value) any {
		return ws.IsConnected()
	}))

	js.Global().Set("__wasm_onWSChange", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		cb := args[0]
		ws.OnChange(func() {
			cb.Invoke()
		})
		return nil
	}))

	js.Global().Set("__wasm_getLatencyMs", js.FuncOf(func(this js.Value, args []js.Value) any {
		return float64(ws.GetLatencyMs())
	}))

	js.Global().Set("__wasm_forceReconnect", js.FuncOf(func(this js.Value, args []js.Value) any {
		ws.ForceReconnect()
		return nil
	}))

	js.Global().Set("__wasm_isReconnecting", js.FuncOf(func(this js.Value, args []js.Value) any {
		return ws.IsReconnecting()
	}))

	slog.Info("WASM Zeitnahme module ready")

	select {}
}
