package ws

import (
	"context"
	"log"
	"sync"
	"time"
	
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
	
	// Dependencies for business logic (Now exported!)
	MessageService domain.MessageService // EXPORTED FIELD
	GroupService domain.GroupService     // EXPORTED FIELD

	// Mutex to protect the clients map
	mu sync.RWMutex
}

// NewHub creates and returns a new Hub, injected with MessageService and GroupService.
func NewHub(messageService domain.MessageService, groupService domain.GroupService) *Hub {
	return &Hub{
		Broadcast: make(chan *Message),
		Register: make(chan *Client),
		Unregister: make(chan *Client),
		clients: make(map[int64][]*Client),
		// FIX 1: Use exported field names in struct literal
		MessageService: messageService, 
		GroupService: groupService, 
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
	// 1. Construct the domain message struct
	domainMsg := &domain.Message{
		SenderID:    message.SenderID,
		RecipientID: message.RecipientID,
		Type:        message.Type,
		Content:     message.Content,
		MediaURL:    message.MediaURL,
		Timestamp:   time.Now(),
	}

	// 2. Persistence (Always Persist First)
	// FIX 2: Use exported field name h.MessageService
	persistedDomainMsg, err := h.MessageService.Save( 
		context.Background(),
		domainMsg,
	)

	if err != nil {
		log.Printf("Error persisting message from %d to %d: %v", domainMsg.SenderID, domainMsg.RecipientID, err)
		return
	}

	// 3. TYPE CONVERSION: Convert the *domain.Message to a *ws.Message for routing
	persistedWsMsg := h.domainToWsMessage(persistedDomainMsg)

	// 4. Determine Recipients (P2P vs Group)
	
	// Start with the sender (for echo)
	recipientIDs := []int64{persistedWsMsg.SenderID}

	// Try to get members list; if it succeeds, it's a group message.
	// FIX 3: Use exported field name h.GroupService
	members, err := h.GroupService.GetMembers(context.Background(), persistedWsMsg.RecipientID) 

	if err == nil && len(members) > 0 {
		// CASE A: Group Message (RecipientID is a valid GroupID)
		log.Printf("Broadcasting message (ID %d) from User %d to Group %d (%d members)...",
			persistedWsMsg.ID, persistedWsMsg.SenderID, persistedWsMsg.RecipientID, len(members))
		
		// Add all group members to the recipients list
		recipientIDs = append(recipientIDs, members...)

	} else {
		// CASE B: P2P Message (RecipientID is a UserID or an unknown ID)
		log.Printf("P2P message (ID %d) from User %d to User %d...",
			persistedWsMsg.ID, persistedWsMsg.SenderID, persistedWsMsg.RecipientID)

		// Add the single recipient user (only if it wasn't a valid group broadcast)
		if persistedWsMsg.RecipientID != 0 {
			recipientIDs = append(recipientIDs, persistedWsMsg.RecipientID)
		} else {
			log.Println("Warning: Global broadcast not yet implemented and RecipientID is 0.")
		}
	}
	
	// 5. Dispatch the message to all determined recipients
	sentTo := make(map[int64]bool)
	for _, userID := range recipientIDs {
		// Ensure we don't send the message multiple times (e.g., sender is also a group member)
		if !sentTo[userID] {
			h.sendMessageToUser(userID, persistedWsMsg)
			sentTo[userID] = true
		}
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

// domainToWsMessage converts a domain message to a WebSocket message format.
func (h *Hub) domainToWsMessage(dMsg *domain.Message) *Message {
    return &Message{
        SenderID:    dMsg.SenderID,
        RecipientID: dMsg.RecipientID,
        Type:        dMsg.Type,
        Content:     dMsg.Content,
	MediaURL:    dMsg.MediaURL,
        Timestamp:   dMsg.Timestamp,
        ID:          dMsg.ID, 
    }
}

// BroadcastGroupMessage implements the domain.Hub interface.
// It is called by the MessageService after a group message is saved.
func (h *Hub) BroadcastGroupMessage(groupID int64, message *domain.Message) {
    // Convert the domain message to the Hub's internal ws.Message type
    wsMsg := h.domainToWsMessage(message)
    
    // Send the message into the hub's main broadcast channel for processing/routing
    h.Broadcast <- wsMsg
}



// BroadcastP2PMessage implements the domain.Hub interface.
// It is called by the MessageService after a P2P message is saved.
func (h *Hub) BroadcastP2PMessage(senderID int64, recipientID int64, message *domain.Message) {
    // 1. Convert the domain message to the Hub's internal ws.Message type
    wsMsg := h.domainToWsMessage(message)

    // 2. Dispatch the message directly to the sender (for echo) and the recipient.
    
    // Send to sender (echo)
    h.sendMessageToUser(senderID, wsMsg)

    // Send to recipient (only if the recipient is not the sender)
    if senderID != recipientID {
        h.sendMessageToUser(recipientID, wsMsg)
    }

    log.Printf("P2P message (ID %d) from User %d dispatched to User %d.", wsMsg.ID, senderID, recipientID)
}
