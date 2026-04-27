package socket

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type incomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	user     *models.User
	send     chan Message
	roomCode string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		parsed, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return isOriginAllowed(parsed)
	},
}

func ServeWs(w http.ResponseWriter, r *http.Request, user *models.User) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:  defaultHub,
		conn: conn,
		user: user,
		send: make(chan Message, 16),
	}

	go client.WritePump()
	client.ReadPump()
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Leave(c)
		close(c.send)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var message incomingMessage
		if err := c.conn.ReadJSON(&message); err != nil {
			return
		}

		switch message.Type {
		case MessageTypeJoinRoom:
			if err := c.handleJoinRoomMessage(message.Payload); err != nil {
				continue
			}
		case room.MessageTypeToggleReady:
			if err := c.hub.HandleRoomMessage(c, message.Type); err != nil {
				continue
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleJoinRoomMessage(payload json.RawMessage) error {
	var joinPayload JoinRoomPayload
	if err := json.Unmarshal(payload, &joinPayload); err != nil {
		return err
	}

	roomModel, err := loadRoomByCode(joinPayload.RoomCode)
	if err != nil {
		return err
	}

	initialMessages, err := c.hub.Join(roomModel, c)
	if err != nil {
		return err
	}

	for _, message := range initialMessages {
		c.send <- message
	}

	return nil
}

func loadRoomByCode(code string) (*models.Room, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	return room.FindByCode(database, code)
}

func allowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:4200"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func isOriginAllowed(origin *url.URL) bool {
	if origin == nil {
		return false
	}
	originStr := origin.Scheme + "://" + origin.Host
	for _, allowed := range allowedOrigins() {
		if strings.TrimRight(allowed, "/") == originStr {
			return true
		}
	}
	return false
}
