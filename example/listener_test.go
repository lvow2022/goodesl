package example

import (
	"github.com/lvow2022/goodesl"
	"log"
	"testing"
	"time"
)

func TestListener(t *testing.T) {
	goodesl.SetDebug(true)
	events := []string{
		"CUSTOM sofia::register",
	}
	el, err := goodesl.NewListener("127.0.0.1:8021", "ClueCon",
		goodesl.WithListenerTimeout(5*time.Second),
		goodesl.WithMaxRetryCount(3),
		goodesl.WithEvents("json", events))

	if err != nil {
		panic(err)
	}

	el.On("sofia::register", func(msg map[string]string) {
		log.Printf("Event-Name: %s", msg["Event-Name"])
		log.Printf("Event-Subclass: %s", msg["Event-Subclass"]) // Example of verifying expected event details
		log.Printf("FreeSWITCH-Hostname: %s", msg["FreeSWITCH-Hostname"])
	})

	el.EventLoop()
}
