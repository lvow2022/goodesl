package goodesl

import (
	net2 "github.com/lvow2022/goodesl/net"
	"strings"
	"time"
)

const (
	DEFAULT_TIMEOUT = 5 * time.Second
	MaxRetries      = 5
)

type Result struct {
	Message *net2.Message
	Err     error
}

func isErrorBody(body []byte) bool {
	return body != nil && strings.HasPrefix(string(body), "-ERR")
}
