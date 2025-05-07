package goodesl

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClient_Api(t *testing.T) {
	c, err := NewClient("127.0.0.1:8021", "ClueCon")
	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := NewCmdBuilder(Status).Build()
	r, err := c.Api(ctx, cmd)
	if err != nil {
		fmt.Println("Error getting status:", err)
	}

	fmt.Println("Status resp:", string(r.Message.Body))

}

func TestClient_Concurrency(t *testing.T) {
	client, err := NewClient("127.0.0.1:8021", "ClueCon")
	if err != nil {
		t.Fatal("Failed to create client:", err)
	}
	//defer client.Close()

	var wg sync.WaitGroup
	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := fmt.Sprintf("expr %d*2", i)
			result, err := client.Api(context.Background(), cmd)
			if err != nil {
				t.Errorf("Error for %d: %v", i, err)
				return
			}
			expected := fmt.Sprintf("%d", i*2)
			if strings.TrimSpace(string(result.Message.Body)) != expected {
				t.Errorf("Unexpected result for %d: got %s, want %s", i, result.Message.Body, expected)
			}
		}(i)
	}
	wg.Wait()
}
