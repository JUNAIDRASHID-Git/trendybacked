// order_websocket.go
package orderControllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/junaidrashid-git/ecommerce-api/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsClients = make(map[*websocket.Conn]bool)

func OrderWebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	wsClients[conn] = true

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(wsClients, conn)
			break
		}
	}
}

func broadcastNewOrder(order models.Order) {
	data, err := json.Marshal(order)
	if err != nil {
		return
	}
	for client := range wsClients {
		client.WriteMessage(websocket.TextMessage, data)
	}
}
