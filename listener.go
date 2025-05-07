package goodesl

import (
	"encoding/json"
	"errors"
	"fmt"
	inet "github.com/lvow2022/goodesl/net"
	"io"
	"net"
	"strings"
	"time"
)

type HandleFunc func(msg map[string]string)

type ListenerOptions func(l *Listener)
type Listener struct {
	conn          *inet.Connection
	addr          string
	password      string
	timeout       time.Duration
	poller        *inet.Poller
	stopCh        chan struct{}
	router        map[string]HandleFunc
	connected     bool
	maxRetryCount int
	events        []string
	eventFormat   string
	interval      time.Duration
}

func NewListener(addr string, password string, opts ...ListenerOptions) (*Listener, error) {
	el := &Listener{
		addr:          addr,
		password:      password,
		timeout:       DEFAULT_TIMEOUT,
		stopCh:        make(chan struct{}),
		router:        make(map[string]HandleFunc),
		interval:      10 * time.Second,
		maxRetryCount: 3,
	}

	for _, opt := range opts {
		opt(el)
	}

	err := el.connect()
	if err != nil {
		return nil, err
	}

	return el, nil
}

func WithMaxRetryCount(count int) ListenerOptions {
	return func(l *Listener) {
		l.maxRetryCount = count
	}
}

func WithListenerTimeout(timeout time.Duration) ListenerOptions {
	return func(l *Listener) {
		l.timeout = timeout
	}
}

func WithEvents(format string, events []string) ListenerOptions {
	return func(l *Listener) {
		l.eventFormat = format
		l.events = events
	}
}

// connect 是受保护的连接过程
func (l *Listener) connect() error {
	c, err := net.DialTimeout("tcp", l.addr, 5*time.Second)
	if err != nil {
		Debugf("Dial failed to %s: %v", l.addr, err)
		return err
	}

	l.conn = inet.NewConnection(c)
	l.poller = inet.NewPoller(l.conn)

	// 设置事件处理器
	l.poller.SetApiResponseCallback(l.ApiResponseCallback)
	l.poller.SetCommandReplyCallback(l.CommandReplyCallback)
	l.poller.SetTextEventPlainCallback(l.TextEventPlainCallback)
	l.poller.SetTextEventJsonCallback(l.TextEventJsonCallback)
	l.poller.SetAuthRequestCallback(l.OnAuthRequest)

	if err = l.poller.Poll(); err != nil {
		l.connected = false
		Debugf("Authentication failed during poll to %s: %v", l.addr, err)
		return fmt.Errorf("authentication error: %w", err)
	}

	Debugf("Connected to %s", l.addr)
	l.connected = true

	if err := l.event(l.eventFormat, l.events); err != nil {
		Debugf("Event subscription failed to %s: %v", l.addr, err)
		return err
	}

	return nil
}

func (l *Listener) Listen() {
	go l.EventLoop()
}

func (l *Listener) EventLoop() {
	Debug("Event loop started")
	retryCount := 0

	for {
		select {
		case <-l.stopCh:
			Debug("Event loop stopped")
			if l.conn != nil {
				_ = l.conn.Close()
			}
			return

		default:
			if !l.connected {
				if l.conn != nil {
					_ = l.conn.Close()
					l.conn = nil
				}

				err := l.connect()
				if err != nil {
					retryCount++
					Debugf("Reconnect attempt %d to %s failed: %v", retryCount, l.addr, err)
					if retryCount >= l.maxRetryCount {
						Debugf("Maximum reconnect attempts (%d) reached for %s", l.maxRetryCount, l.addr)
						return
					}
					time.Sleep(l.interval)
					continue
				}

				retryCount = 0
				Debugf("Reconnected to %s", l.addr)
			}

			if err := l.poller.Poll(); err != nil {
				Debugf("Polling failed: %v", err)
				if errors.Is(err, io.EOF) {
					Debug("Disconnected (EOF)")
					l.connected = false
					continue
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (l *Listener) Stop() {
	select {
	case <-l.stopCh:

	default:
		close(l.stopCh)
	}
}

func (l *Listener) On(event string, handler HandleFunc) {
	l.router[event] = handler
}

func (l *Listener) ApiResponseCallback(msg *inet.Message) {

}

func (l *Listener) CommandReplyCallback(msg *inet.Message) {

}

func (l *Listener) TextEventPlainCallback(msg *inet.Message) {

}

func (l *Listener) TextEventJsonCallback(msg *inet.Message) {
	if msg.Body == nil {
		Debug("Received JSON event with empty body")
		return
	}

	var event map[string]string
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		Debugf("Failed to parse JSON event: %v", err)
		return
	}

	eventName := event["Event-Subclass"]
	if eventName == "" {
		eventName = event["Event-Name"]
	}

	if handler, ok := l.router[eventName]; ok {
		Debugf("Handling event: %s", eventName)
		handler(event)
	} else {
		Debugf("Unhandled event: %s", eventName)
	}
}

func (l *Listener) OnAuthRequest(msg *inet.Message) {
	msg, err := l.conn.SendRecv("auth %s\r\n", l.password)
	if err != nil {
		Debugf("Authenticatio failed: %v", err)
	}
	Debugf("Authenticated with %s", l.addr)
}

func (l *Listener) event(format string, events []string) error {
	if format != "plain" && format != "json" {
		return fmt.Errorf("unsupported event format: %s", format)
	}

	msg, err := l.conn.SendRecv("event %s %s\r\n", format, strings.Join(events, " "))
	if err != nil {
		return err
	}

	result := &Result{Message: msg}
	if isErrorBody(msg.Body) {
		errStr, _ := strings.CutPrefix(string(msg.Body), "-ERR")
		result.Err = errors.New(strings.TrimSpace(errStr))
	}
	return nil
}
