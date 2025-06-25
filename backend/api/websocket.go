package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	HandshakeTimeout: 45 * time.Second,
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// WebSocket client manager
type Client struct {
	conn *websocket.Conn
	send chan []byte
	hub  *Hub
	id   string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

var wsHub = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

// Initialize WebSocket hub
func init() {
	go wsHub.run()
	go wsHub.broadcastDeviceUpdates()
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast device updates every 5 seconds
func (h *Hub) broadcastDeviceUpdates() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Create device update message
			deviceData := map[string]interface{}{
				"type":      "device_update",
				"timestamp": time.Now().Format(time.RFC3339),
				"devices":   devices,
				"count":     len(devices),
			}

			// Convert to JSON
			if message, err := json.Marshal(deviceData); err == nil {
				h.broadcast <- message
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Set read deadline and pong handler for keep-alive
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// Read message from client (if any)
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Authentication function that checks against the users from the JSON storage
func validateUser(email, password string) bool {
	// Debug: Print what we're looking for
	fmt.Printf("DEBUG: Validating email=%s, password=%s\n", email, password)

	// Get current users from storage
	users := GetUsersData()
	fmt.Printf("DEBUG: Total users in system: %d\n", len(users))

	// Check against all users in the system
	for i, user := range users {
		fmt.Printf("DEBUG: User %d - Email=%s, Password=%s\n", i, user.Email, user.Password)
		if user.Email == email && user.Password == password {
			fmt.Printf("DEBUG: Found matching user!\n")
			return true
		}
	}
	fmt.Printf("DEBUG: No matching user found\n")
	return false
}

// Struct for login requests
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// WebSocket connection handler
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}

	// Create client
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  wsHub,
		id:   fmt.Sprintf("client_%d", time.Now().UnixNano()),
	}

	// Register client
	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	log.Printf("WebSocket connection established for client: %s", client.id)
}

// SetupWebSocketRoute sets up the WebSocket route
func SetupWebSocketRoute(router *gin.Engine) {
	router.GET("/ws", WebSocketHandler)
	router.POST("/tokens", TokenHandler)
}

func TokenHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Validate credentials using the persistent user storage
	if !validateUser(req.Email, req.Password) {
		fmt.Printf("Authentication failed for user: %s\n", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": req.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
		return
	}

	fmt.Printf("Successfully authenticated user: %s\n", req.Email)
	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenString,
	})
}
