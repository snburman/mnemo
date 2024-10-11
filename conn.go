package mnemo

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type (
	// Conn is a websocket connection with a unique key, a pointer to a pool, and a channel for messages.
	Conn struct {
		websocket *websocket.Conn
		Pool      *Pool
		Key       interface{}
		Messages  chan interface{}
	}
)

// NewConn upgrades an http connection to a websocket connection and returns a Conn
// or an error if the upgrade fails.
func NewConn(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	upgrader := websocket.Upgrader{}
	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, NewError[Conn](err.Error()).WithStatus(http.StatusInternalServerError)
	}

	c := &Conn{
		websocket: websocket,
		Key:       uuid.New(),
		Messages:  make(chan interface{}, 16),
	}
	return c, nil
}

// Close closes the websocket connection and removes the Conn from the pool.
// It returns an error if the Conn is nil.
func (c *Conn) Close() error {
	if c == nil {
		return NewError[Conn]("connection is nil")
	}
	c.Pool.removeConnection(c)
	c.websocket.Close()
	return nil
}

// Listen listens for messages on the Conn's Messages channel and writes them to the websocket connection.
func (c *Conn) Listen() {
	go func(c *Conn) {
		for {
			if _, _, err := c.websocket.ReadMessage(); err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure,
				) {
					NewError[Conn](err.Error()).Log()
				}
				close(c.Messages)
				break
			}
		}
	}(c)

	for {
		msg, ok := <-c.Messages
		if !ok {
			c.Close()
			break
		}

		if err := c.websocket.WriteJSON(msg); err != nil {
			NewError[Conn](err.Error()).Log()
			c.Close()
		}
	}
}

// Publish publishes a message to the Conn's Messages channel.
func (c *Conn) Publish(msg interface{}) {
	// if msg is not json encodable, return
	_, err := json.Marshal(msg)
	if err != nil {
		NewError[Conn](err.Error()).Log()
		return
	}
	c.Messages <- msg
}
