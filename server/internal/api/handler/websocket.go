package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mobileproxy/server/internal/domain"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type WSHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *WSHub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	log.Printf("WebSocket client connected (total: %d)", len(h.clients))

	// Keep connection alive, read messages (but we only broadcast)
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
			log.Printf("WebSocket client disconnected (total: %d)", len(h.clients))
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func (h *WSHub) Broadcast(msg domain.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WebSocket marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
