// order_websocket.go
package orderControllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/junaidrashid-git/ecommerce-api/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 70 * time.Second    // allow slightly > nginx timeout
	pingPeriod     = (pongWait * 9) / 10 // send pings before pongWait expires
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// tighten this for production if you can (check origin/host)
		return true
	},
}

// ---------- Hub ----------
type hub struct {
	clients    map[*client]bool
	broadcast  chan []byte
	register   chan *client
	unregister chan *client
}

func newHub() *hub {
	return &hub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}

var globalHub = newHub()

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				c.conn.Close()
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				// non-blocking send: if client send buffer full, drop client
				select {
				case c.send <- msg:
				default:
					// client too slow, remove it
					close(c.send)
					delete(h.clients, c)
					c.conn.Close()
				}
			}
		}
	}
}

// ---------- Client ----------
type client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *client) readPump(h *hub) {
	defer func() {
		h.unregister <- c
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// We don't expect to process incoming messages for now,
		// but we must read to detect closed connections / pongs.
		if _, _, err := c.conn.NextReader(); err != nil {
			break
		}
	}
}

func (c *client) writePump(h *hub) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write the message as single TextMessage
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}
			if err := w.Close(); err != nil {
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

// ---------- Handler & Broadcast API ----------
func OrderWebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &client{
		conn: conn,
		send: make(chan []byte, 16), // buffered to avoid blocking hub
	}

	globalHub.register <- client

	// Start pumps
	go client.writePump(globalHub)
	client.readPump(globalHub) // readPump runs in current goroutine and blocks until closed
}

// BroadcastNewOrder marshals the order and sends it to all connected clients.
// Call this from your order creation code (exported function).
func BroadcastNewOrder(order models.Order) {
	data, err := json.Marshal(order)
	if err != nil {
		log.Println("broadcast marshal error:", err)
		return
	}
	// log so we can see when a broadcast is triggered
	log.Printf("[ws] BroadcastNewOrder called for order id=%v\n", order.ID)
	globalHub.broadcast <- data
}

func init() {
	// start hub loop
	go globalHub.run()
}
