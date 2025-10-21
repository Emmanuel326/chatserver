package ws

import (
	"context"
	"log"
	"sync"
	
	"github.com/Emmanuel326/chatserver/internal/domain"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients: maps a UserID to a set of active Clients
	clients map[int64][]*Client

	// Inbound messages from the clients
	Broadcast chan *Message
	
	// Client registration/unregistration channels
	Register chan *Client
	Unregister chan *Client
	
	// Dependency for persistence
	messageService domain.MessageService

	// Mutex to protect the clients map
	mu sync.RWMutex
}

// NewHub creates and returns a new Hub, injected with MessageService.
func NewHub(messageService domain.MessageService) *Hub {
	return &Hub{
		Broadcast: make(chan *Message),
		Register: make(chan *Client),
		Unregister: make(chan *Client),
		clients: make(map[int64][]*Client),
		messageService: messageService,
	}
}

// Run starts the Hub's main Goroutine.
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

// handleBroadcast routes a message to the intended recipients and persists it.
func (h *Hub) handleBroadcast(message *Message) {
	// 1. Persist the message (returns a *domain.Message)
	persistedDomainMsg, err := h.messageService.Save(
		context.Background(),
		message.SenderID,
		message.RecipientID,
		message.Type,
		message.Content,
	)

	if err != nil {
		log.Printf("Error persisting message from %d to %d: %v", message.SenderID, message.RecipientID, err)
		return
	}

	// 2. TYPE CONVERSION: Convert the *domain.Message to a *ws.Message for routing
	persistedWsMsg := h.domainToWsMessage(persistedDomainMsg)

	// 3. Route the message
	if persistedWsMsg.RecipientID != 0 {
		// Handle private message (P2P)
		h.sendMessageToUser(persistedWsMsg.SenderID, persistedWsMsg) // Send echo to sender
		h.sendMessageToUser(persistedWsMsg.RecipientID, persistedWsMsg) // Send to recipient
	} else {
		// Handle room/global message
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

// Helper function to convert a *domain.Message to a *ws.Message for transport.
// This resolves the type mismatch error in handleBroadcast.
func (h *Hub) domainToWsMessage(dMsg *domain.Message) *Message {
    return &Message{
        SenderID:    dMsg.SenderID,
        RecipientID: dMsg.RecipientID,
        Type:        dMsg.Type,
        Content:     dMsg.Content,
        Timestamp:   dMsg.Timestamp, 
    }
}
