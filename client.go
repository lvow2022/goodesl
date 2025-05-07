package goodesl

import (
	"context"
	"errors"
	"fmt"
	inet "github.com/lvow2022/goodesl/net"
	"net"
	"strings"
	"time"
)

// inbound mode
type Client struct {
	conn      *inet.Connection
	addr      string
	password  string
	timeout   time.Duration
	poller    *inet.Poller
	connected bool
}

type ClientOption func(c *Client)

func NewClient(addr string, password string, opts ...ClientOption) (*Client, error) {

	// 默认配置
	c := &Client{
		addr:     addr,
		password: password,
		timeout:  DEFAULT_TIMEOUT,
	}

	// 应用用户配置
	for _, opt := range opts {
		opt(c)
	}

	err := c.connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return err
	}

	c.conn = inet.NewConnection(conn)
	c.poller = inet.NewPoller(c.conn)

	// 设置事件处理器
	c.poller.SetApiResponseCallback(c.ApiResponseCallback)
	c.poller.SetCommandReplyCallback(c.CommandReplyCallback)
	c.poller.SetAuthRequestCallback(c.AuthRequestCallback)

	if err = c.poller.Poll(); err != nil {
		c.connected = false
		Debugf("Authentication failed during poll to %s: %v", c.addr, err)
		return fmt.Errorf("authentication error: %w", err)
	}

	return nil
}

func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = d
	}
}

func WithPassword(password string) ClientOption {
	return func(c *Client) {
		c.password = password
	}
}

func (c *Client) Api(ctx context.Context, command string) (*Result, error) {
	msg, err := c.conn.SendRecv("API %s\r\n", command)
	if err != nil {
		return nil, err
	}

	result := &Result{Message: msg}
	if isErrorBody(msg.Body) {
		errStr, _ := strings.CutPrefix(string(msg.Body), "-ERR")
		result.Err = errors.New(strings.TrimSpace(errStr))
	}
	return result, nil
}

func (c *Client) ApiResponseCallback(msg *inet.Message) {

}

func (c *Client) CommandReplyCallback(msg *inet.Message) {

}

func (c *Client) AuthRequestCallback(msg *inet.Message) {
	msg, err := c.conn.SendRecv("auth %s\r\n", c.password)
	if err != nil {
		Debugf("Authenticatio failed: %v", err)
	}
	Debugf("Authenticated with %s", c.addr)
}
