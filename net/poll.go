package net

import (
	"errors"
	"log"
	"net/textproto"
)

var (
	// 通用错误
	ErrConnectionFailed  = errors.New("connection failed")
	ErrInvalidContent    = errors.New("invalid content type")
	ErrAuthFailed        = errors.New("authentication failed")
	ErrUnexpectedReply   = errors.New("unexpected reply format")
	ErrPollTimeout       = errors.New("poll timeout")
	ErrResponseCorrupted = errors.New("response corrupted")
	ErrContentLength     = errors.New("invalid Content-Length")
)

type Message struct {
	StartLine     string
	Headers       textproto.MIMEHeader
	ContentType   string
	ContentLength string
	ReplyText     string
	Body          []byte
	Raw           []byte
}

type EventCallback func(msg *Message)
type Poller struct {
	conn                   *Connection
	CommandReplyCallback   EventCallback
	ApiResponseCallback    EventCallback
	AuthRequestCallback    EventCallback
	TextEventPlainCallback EventCallback
	TextEventJsonCallback  EventCallback
}

func NewPoller(conn *Connection) *Poller {
	return &Poller{conn: conn}
}

func (p *Poller) SetCommandReplyCallback(cb EventCallback) {
	p.CommandReplyCallback = cb
}

func (p *Poller) SetApiResponseCallback(cb EventCallback) {
	p.ApiResponseCallback = cb
}

func (p *Poller) SetAuthRequestCallback(cb EventCallback) {
	p.AuthRequestCallback = cb
}

func (p *Poller) SetTextEventPlainCallback(cb EventCallback) {
	p.TextEventPlainCallback = cb
}

func (p *Poller) SetTextEventJsonCallback(cb EventCallback) {
	p.TextEventJsonCallback = cb
}

func (p *Poller) Poll() error {
	msg, err := p.conn.RecvEvent()
	if err != nil {
		return err
	}

	switch msg.ContentType {
	case AUTH_REQUEST:
		if p.AuthRequestCallback != nil {
			p.AuthRequestCallback(msg)
		}
	case API_RESPONSE:
		if p.ApiResponseCallback != nil {
			p.ApiResponseCallback(msg)
		}
	case COMMAND_REPLY:
		if p.CommandReplyCallback != nil {
			p.CommandReplyCallback(msg)
		}
	case TEXT_EVENTPLAIN:
		if p.TextEventPlainCallback != nil {
			p.TextEventPlainCallback(msg)
		}
	case TEXT_EVENTJSON:
		if p.TextEventJsonCallback != nil {
			p.TextEventJsonCallback(msg)
		}
	default:
		log.Printf("[Poller] unknown content type: %q", msg.ContentType)
		return ErrInvalidContent
	}

	return nil
}
