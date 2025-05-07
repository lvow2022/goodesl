package goodesl

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEventListener_EventJson(t *testing.T) {
	events := []string{
		"CUSTOM sofia::register",
	}

	el, err := NewListener("127.0.0.1:8021", "ClueCon",
		WithEvents("json", events))
	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}

	el.On("sofia::register", func(msg map[string]string) {
		assert.Equal(t, "CUSTOM", msg["Event-Name"])
		assert.Equal(t, "sofia::register", msg["Event-Subclass"]) // Example of verifying expected event details
		assert.NotEmpty(t, msg["Core-UUID"])                      // Ensure Core-UUID is not
	})
	el.Listen()

	time.Sleep(60 * time.Second)
}
