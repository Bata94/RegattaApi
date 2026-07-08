//go:build js && wasm

package main

import (
	"encoding/json"
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

	js.Global().Set("__wasm_submitFinish", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			slog.Warn("__wasm_submitFinish: need 2 args")
			return nil
		}
		startNr := args[0].String()
		timeStr := args[1].String()

		timeClient, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			timeClient = time.Now()
		}

		pf := store.AddPendingFinish(startNr, timeClient, int(ws.GetLatencyMs()))
		ws.SendFinish(pf)
		return nil
	}))

	js.Global().Set("__wasm_getOpenStarts", js.FuncOf(func(this js.Value, args []js.Value) any {
		starts := store.GetOpenStarts()
		data, err := json.Marshal(starts)
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
