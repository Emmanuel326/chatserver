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
	
	// Inbound typing notifications from clients
	Typing chan *Message

	// Client registration/unregistration channels
	Register chan *Client
	Unregister chan *Client
	
	// Dependencies for business logic
	MessageService domain.MessageService
	GroupService domain.GroupService
	UserService domain.UserService

	// Mutex to protect the clients map
	mu sync.RWMutex

	// Channel to signal the Hub to stop
	quit chan struct{}
}

// NewHub creates and returns a new Hub, injected with MessageService, GroupService, and UserService.
func NewHub(messageService domain.MessageService, groupService domain.GroupService, userService domain.UserService) *Hub {
	return &Hub{
		Broadcast:      make(chan *Message),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Typing:         make(chan *Message),
		clients:        make(map[int64][]*Client),
		MessageService: messageService,
		GroupService:   groupService,
		UserService:    userService,
		quit:           make(chan struct{}), // Initialize the quit channel
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
		case notification := <-h.Typing:
			h.handleTypingNotification(notification)
		case <-h.quit:
			log.Println("🛑 WebSocket Hub stopping.")
			return // Exit the Run loop
		}
	}
}

// Stop sends a signal to the Hub's Run goroutine to terminate.
func (h *Hub) Stop() {
	close(h.quit)
}

// handleRegister adds a new client connection to the hub's map and triggers pending message delivery.
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	userID := client.UserID
	h.clients[userID] = append(h.clients[userID], client)
	log.Printf("Client registered. UserID: %d. Total connections for user: %d", userID, len(h.clients[userID]))
	h.mu.Unlock()

	client.Send <- NewSystemMessage("Welcome to the chat server.")

	// Launch a goroutine to fetch and deliver any messages that were sent while the user was offline.
	go h.deliverPendingMessages(client)
}

// deliverPendingMessages fetches and sends pending messages to a newly connected client.
func (h *Hub) deliverPendingMessages(client *Client) {
	// 1. Fetch pending messages from the service layer.
	pendingMessages, err := h.MessageService.GetPendingMessages(context.Background(), client.UserID)
	if err != nil {
		log.Printf("Error fetching pending messages for User %d: %v", client.UserID, err)
		return
	}

	if len(pendingMessages) == 0 {
		return
	}
	log.Printf("Delivering %d pending messages to User %d", len(pendingMessages), client.UserID)

	// 2. Send each pending message and collect the IDs of those successfully sent.
	var deliveredIDs []int64
	for _, msg := range pendingMessages {
		wsMsg := h.domainToWsMessage(msg)

		// Safely send the message by checking if the client is still connected under a lock.
		// This prevents a panic if the client disconnects and its channel is closed concurrently.
		sent := func() bool {
			h.mu.RLock()
			defer h.mu.RUnlock()

			// First, verify the client instance is still in the active clients map.
			if connections, ok := h.clients[client.UserID]; ok {
				isStillConnected := false
				for _, c := range connections {
					if c == client {
						isStillConnected = true
						break
					}
				}
				if !isStillConnected {
					return false // Client has been unregistered.
				}
			} else {
				return false // User has no connections.
			}

			// If client is still connected, attempt a non-blocking send.
			select {
			case client.Send <- wsMsg:
				return true // Success.
			default:
				return false // Buffer is full.
			}
		}()

		if sent {
			deliveredIDs = append(deliveredIDs, msg.ID)
		} else {
			// If the message could not be sent (client disconnected or buffer full),
			// abort delivery to ensure messages aren't marked as delivered incorrectly.
			log.Printf("Client for User %d disconnected or buffer full. Aborting pending message delivery.", client.UserID)
			goto updateStatus
		}
	}

updateStatus:
	// 3. Mark the successfully sent messages as 'DELIVERED'.
	if len(deliveredIDs) > 0 {
		if err := h.MessageService.MarkMessagesAsDelivered(context.Background(), deliveredIDs); err != nil {
			log.Printf("Error marking messages as delivered for User %d: %v", client.UserID, err)
		}
	}
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

// handleTypingNotification broadcasts a typing indicator to relevant users without persisting it.
func (h *Hub) handleTypingNotification(message *Message) {
	// A typing notification is transient and should not be persisted.
	// We determine if the recipient is a group or a single user and forward accordingly.

	members, err := h.GroupService.GetMembers(context.Background(), message.RecipientID)

	if err == nil && len(members) > 0 {
		// Case A: Group Typing Notification.
		// Broadcast to all group members except the sender.
		for _, memberID := range members {
			if memberID != message.SenderID {
				h.sendMessageToUser(memberID, message)
			}
		}
	} else {
		// Case B: P2P Typing Notification.
		// Send only to the single recipient.
		h.sendMessageToUser(message.RecipientID, message)
	}
}

// handleBroadcast determines if a message is P2P or Group and routes it to the appropriate handler.
func (h *Hub) handleBroadcast(message *Message) {
	// Check if the recipient ID corresponds to a valid group.
	members, err := h.GroupService.GetMembers(context.Background(), message.RecipientID)

	if err == nil && len(members) > 0 {
		// Case A: It's a group message.
		h.handleGroupBroadcast(message, members)
	} else {
		// Case B: It's a P2P message.
		h.handleP2PBroadcast(message)
	}
}

// handleGroupBroadcast persists a group message by calling the message service.
// The service is then responsible for calling back to the hub to dispatch the message.
func (h *Hub) handleGroupBroadcast(message *Message, members []int64) {
	log.Printf("Persisting GROUP message from User %d to Group %d...", message.SenderID, message.RecipientID)

	domainMsg := &domain.Message{
		SenderID:    message.SenderID,
		RecipientID: message.RecipientID,
		Type:        message.Type,
		Content:     message.Content,
		MediaURL:    message.MediaURL,
		Timestamp:   message.Timestamp,
		Status:      domain.MessageSent, // Group messages are not queued for offline users
	}

	// Persist the message. The MessageService will call hub.BroadcastGroupMessage for dispatch.
	if _, err := h.MessageService.Save(context.Background(), domainMsg); err != nil {
		log.Printf("Error persisting group message: %v", err)
	}
}

// handleP2PBroadcast persists a P2P message by calling the message service.
// The service is then responsible for calling back to the hub to dispatch the message.
func (h *Hub) handleP2PBroadcast(message *Message) {
	log.Printf("Persisting P2P message from User %d to User %d...", message.SenderID, message.RecipientID)

	// Determine message status based on recipient's online status.
	status := domain.MessageSent
	if !h.isUserOnline(message.RecipientID) {
		status = domain.MessagePending
	}

	domainMsg := &domain.Message{
		SenderID:    message.SenderID,
		RecipientID: message.RecipientID,
		Type:        message.Type,
		Content:     message.Content,
		MediaURL:    message.MediaURL,
		Timestamp:   message.Timestamp,
		Status:      status,
	}

	// Persist the message. The MessageService will call hub.BroadcastP2PMessage for dispatch.
	if _, err := h.MessageService.Save(context.Background(), domainMsg); err != nil {
		log.Printf("Error persisting P2P message: %v", err)
	}
}

// isUserOnline checks if a user has at least one active WebSocket connection.
func (h *Hub) isUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	connections, ok := h.clients[userID]
	return ok && len(connections) > 0
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
				// If the client's send buffer is full, it's likely stuck.
				// Unregister it in a new goroutine to avoid deadlocking the hub's Run loop.
				go func(c *Client) {
					h.Unregister <- c
				}(client)
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
// It is called by the MessageService after a group message has been saved.
func (h *Hub) BroadcastGroupMessage(groupID int64, message *domain.Message) {
	wsMsg := h.domainToWsMessage(message)

	// Get the list of group members to dispatch the message.
	members, err := h.GroupService.GetMembers(context.Background(), groupID)
	if err != nil {
		log.Printf("Error getting members for group %d to broadcast message: %v", groupID, err)
		return
	}

	log.Printf("Dispatching GROUP message (ID %d) to %d members of Group %d", wsMsg.ID, len(members), groupID)

	// Dispatch the message only to the ONLINE members of the group.
	for _, memberID := range members {
		if h.isUserOnline(memberID) {
			h.sendMessageToUser(memberID, wsMsg)
		}
	}
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
