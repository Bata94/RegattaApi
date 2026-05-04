package handlers

import (
	"time"
)

type Client struct{}

var Clients = make(map[interface{}]Client)
var Register = make(chan interface{})
var Broadcast = make(chan string)
var Unregister = make(chan interface{})

func WsUpgrade() interface{} {
	return nil
}

func RunHub() {
	for {
		select {
		case <-Register:
		case <-Unregister:
		case msg := <-Broadcast:
			for c := range Clients {
				_ = c
			}
			_ = msg
		}
		time.Sleep(time.Second)
	}
}
