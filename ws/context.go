package ws

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rickslab/ares/logger"
	"github.com/sirupsen/logrus"
)

type Context struct {
	request   *http.Request
	conn      *websocket.Conn
	writeChan chan *message
	logger    *logrus.Entry
	ctx       context.Context
	cancel    context.CancelFunc
	values    map[any]any
	mu        sync.RWMutex
}

type message struct {
	Type int
	Data []byte
}

func NewContext(r *http.Request, conn *websocket.Conn, service string, method string) *Context {
	ctx, cancel := context.WithCancel(r.Context())
	c := &Context{
		request:   r,
		conn:      conn,
		writeChan: make(chan *message),
		ctx:       ctx,
		cancel:    cancel,
		values:    make(map[any]any),
	}

	reqId := r.Header.Get("X-Request-Id")
	clientIp := r.Header.Get("X-Real-IP")
	userId := r.Header.Get("X-User-Id")
	scope := r.Header.Get("X-Auth-Scope")

	c.logger = logger.NewEntry(c, map[string]any{
		"request_id": reqId,
		"client_ip":  clientIp,
		"user_id":    userId,
		"scope":      scope,
		"service":    service,
		"method":     method,
	})
	c.values["MetaKv"] = []string{
		"request_id", reqId,
		"client_ip", clientIp,
		"user_id", userId,
		"scope", scope,
		"caller", service,
	}
	c.values["UserId"] = userId

	go c.writeLoop()
	return c
}

func (c *Context) ReadMessage() (int, []byte, error) {
	mt, data, err := c.conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			c.logger.Errorf("ReadMessage error: %v", err)
		}
		return 0, nil, err
	}

	c.conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
	return mt, data, nil
}

func (c *Context) writeLoop() {
	for {
		select {
		case msg := <-c.writeChan:
			err := c.conn.WriteMessage(msg.Type, msg.Data)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Errorf("WriteMessage error: %v", err)
				}
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Context) WriteMessage(mt int, data []byte) {
	select {
	case c.writeChan <- &message{
		Type: mt,
		Data: data,
	}:
	case <-c.ctx.Done():
	}
}

func (c *Context) WriteTextMessage(text string) {
	c.WriteMessage(websocket.TextMessage, []byte(text))
}

func (c *Context) WriteBinaryMessage(data []byte) {
	c.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Context) Cancel() {
	c.cancel()
}

func (c *Context) GetRequest() *http.Request {
	return c.request
}

func (c *Context) Logger() *logrus.Entry {
	return c.logger
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Context) Close() error {
	return c.conn.Close()
}

func (c *Context) Err() error {
	return c.ctx.Err()
}

func (c *Context) Value(key any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.values[key]
}

func (c *Context) Set(key, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[key] = value
}
