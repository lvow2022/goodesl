package net

import (
	"bufio"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"sync"
)

type Connection struct {
	tpConn *textproto.Conn
	r      *bufio.Reader
	w      *bufio.Writer
	mu     sync.Mutex // 保证并发读写安全
}

// 初始化连接
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		tpConn: textproto.NewConn(conn),
	}
}

// 并发安全：发送命令并获取响应
func (c *Connection) SendRecv(format string, args ...any) (*Message, error) {
	id, err := c.tpConn.Cmd(format, args...)
	if err != nil {
		return nil, err
	}

	c.tpConn.StartResponse(id)
	defer c.tpConn.EndResponse(id)

	return c.readMessage()
}

// 可用于事件订阅（持续读）等用途
func (c *Connection) RecvEvent() (*Message, error) {
	return c.readMessage()
}

// 读取 FreeSWITCH 响应内容
func (c *Connection) readMessage() (*Message, error) {
	headers, err := c.tpConn.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Headers:       headers,
		ContentType:   headers.Get("Content-Type"),
		ContentLength: headers.Get("Content-Length"),
		ReplyText:     headers.Get("Reply-Text"),
	}

	if msg.ContentLength != "" {
		contentLength, err := strconv.Atoi(msg.ContentLength)
		if err != nil {
			return nil, err
		}
		msg.Body = make([]byte, contentLength)

		_, err = io.ReadFull(c.tpConn.R, msg.Body)
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (c *Connection) Close() error {
	return c.tpConn.Close()
}
