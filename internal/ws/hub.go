package ws

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
// It runs in a single goroutine to safely manage the state (client map).
type Hub struct {
	// Registered clients: maps a UserID to a set of active Clients (since one user can have multiple devices).
	clients map[int64][]*Client

	// Inbound messages from the clients to be broadcast or processed.
	Broadcast chan *Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
	
	// Mutex to protect the clients map, although operations should primarily happen in the Run loop.
	// Used here mainly for safety checks if accessed outside Run.
	mu sync.RWMutex 
}

// NewHub creates and returns a new Hub.
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[int64][]*Client),
	}
}

// Run starts the Hub's main Goroutine. It reads from the channels and modifies the client map safely.
func (h *Hub) Run() {
	log.Println("⚡️ WebSocket Hub started successfully.")
	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)
		case client := <-h.Unregister:
			h.handleUnregister(client)
		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

// handleRegister adds a new client connection to the hub's map.
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	userID := client.UserID
	h.clients[userID] = append(h.clients[userID], client)
	log.Printf("Client registered. UserID: %d. Total connections for user: %d", userID, len(h.clients[userID]))

	// Optionally send a system message back to the user
	client.Send <- NewSystemMessage("Welcome to the chat server.")
}

// handleUnregister removes a client connection from the hub's map.
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	userID := client.UserID
	if connections, ok := h.clients[userID]; ok {
		// Find and remove the specific client instance
		for i, conn := range connections {
			if conn == client {
				// Efficiently remove client from the slice without preserving order
				h.clients[userID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
		
		// If the user has no more active connections, delete the entry entirely
		if len(h.clients[userID]) == 0 {
			delete(h.clients, userID)
		}
		
		// Close the client's send channel to stop its write pump
		close(client.Send)
		log.Printf("Client unregistered. UserID: %d. Remaining connections for user: %d", userID, len(h.clients[userID]))
	}
}

// handleBroadcast routes a message to the intended recipients.
func (h *Hub) handleBroadcast(message *Message) {
	// TODO: Integrate persistence logic here (i.e., call the MessageService)
	// For now, we only route the message.
	
	if message.RecipientID != 0 {
		// Handle private message (P2P)
		h.sendMessageToUser(message.SenderID, message)
		h.sendMessageToUser(message.RecipientID, message)
	} else {
		// Handle room/global message (TODO: Implement rooms later)
		log.Println("Warning: Global broadcast not yet implemented.")
	}
}

// sendMessageToUser sends a message to all active clients of a specific UserID.
func (h *Hub) sendMessageToUser(userID int64, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if connections, ok := h.clients[userID]; ok {
		for _, client := range connections {
			select {
			case client.Send <- message:
				// Message sent successfully
			default:
				// If the client's send buffer is full, it's likely stuck. Unregister it.
				h.handleUnregister(client)
			}
		}
	}
}
