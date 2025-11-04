package api

import (
	
	"net/http"
	"strconv"

	"github.com/Emmanuel326/chatserver/internal/api/middleware"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/gin-gonic/gin"
)

const defaultHistoryLimit = 50

// MessageHandler holds dependencies for message-related API endpoints.
type MessageHandler struct {
	MessageService domain.MessageService
	GroupService   domain.GroupService
}

// NewMessageHandler creates a new handler instance.
func NewMessageHandler(messageService domain.MessageService, groupService domain.GroupService) *MessageHandler {
	return &MessageHandler{
		MessageService: messageService,
		GroupService:   groupService,
	}
}

// GetConversationHistory retrieves the history of messages between the authenticated user
// and another specified user (recipientID).
func (h *MessageHandler) GetConversationHistory(c *gin.Context) {
	// 1. Get Authenticated User ID (the "self" user)
	senderID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed or user ID missing"})
		return
	}

	// 2. Get Recipient ID from URL path (the "other" user)
	recipientIDStr := c.Param("recipientID")
	recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
	if err != nil || recipientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
		return
	}

	// Prevent user from querying history with themselves (optional validation)
	if senderID == recipientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot retrieve conversation history with self"})
		return
	}

	// 3. Get Optional Limit Query Parameter
	limit := defaultHistoryLimit
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultHistoryLimit))
	
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	beforeIDStr := c.DefaultQuery("before_id", "0")
	beforeID, _ := strconv.ParseInt(beforeIDStr, 10, 64)
	
	// 4. Call Domain Service to retrieve history
	// FIX: Use request context instead of context.Background()
	messages, err := h.MessageService.GetConversationHistory(c.Request.Context(), senderID, recipientID, limit, beforeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve message history"})
		return
	}

	// 5. Success Response
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count": len(messages),
	})
}

// GetRecentConversations retrieves the latest message from each of the user's conversations.
// GET /v1/chats
func (h *MessageHandler) GetRecentConversations(c *gin.Context) {
	// 1. Get Authenticated User ID
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed or user ID missing"})
		return
	}

	// 2. Call Domain Service to retrieve recent conversations
	messages, err := h.MessageService.GetRecentConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recent conversations"})
		return
	}

	// Ensure we return an empty array instead of null if no conversations are found
	if messages == nil {
		messages = []*domain.Message{}
	}

	// 3. Success Response
	c.JSON(http.StatusOK, messages)
}

// SendGroupMessage handles sending a new message to a specific group.
// POST /v1/groups/:groupID/messages
func (h *MessageHandler) SendGroupMessage(c *gin.Context) { 
	// 1. Get authenticated UserID (senderID)
	senderID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sender not authenticated"})
		return
	}

	// 2. Get Group ID from URL parameter
	groupIDStr := c.Param("groupID")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	// 3. Parse request body using the new struct
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
    
    // VALIDATION: Ensure at least Content or MediaURL is present
    if req.Content == "" && req.MediaURL == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Message must contain either text content or a media URL"})
        return
    }

	// 4. Call MessageService to send the message (now includes MediaURL and Type)
	_, err = h.MessageService.SendGroupMessage(c.Request.Context(), senderID, groupID, req.Content, req.MediaURL, req.Type)
	
	if err != nil {
		// Differentiate between domain errors (e.g., membership) and server errors
		if _, ok := err.(*domain.NotFoundError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Failed to send group message", "details": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send group message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message sent successfully to group"})
}

// SendP2PMessage handles sending a new message to a specific user.
// POST /v1/users/:recipientID/messages
func (h *MessageHandler) SendP2PMessage(c *gin.Context) {
	// 1. Get authenticated UserID (senderID)
	senderID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sender not authenticated"})
		return
	}

	// 2. Get Recipient ID from URL parameter
	recipientIDStr := c.Param("recipientID")
	recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
	if err != nil || recipientID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID format"})
		return
	}

	// Prevent user from messaging themselves
	if senderID == recipientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send a P2P message to self"})
		return
	}

	// 3. Parse request body
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// VALIDATION: Ensure at least Content or MediaURL is present
	if req.Content == "" && req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message must contain either text content or a media URL"})
		return
	}

	// 4. Call MessageService to send the message (includes MediaURL and Type)
	_, err = h.MessageService.SendP2PMessage(c.Request.Context(), senderID, recipientID, req.Content, req.MediaURL, req.Type)
	
	if err != nil {
		// Differentiate between domain errors (e.g., user not found) and server errors
		if err.Error() == "recipient user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipient user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send P2P message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "P2P message sent successfully"})
}

// GetGroupConversationHistory retrieves the message history for a specific group.
// GET /v1/groups/:groupID/messages
func (h *MessageHandler) GetGroupConversationHistory(c *gin.Context) {
	// 1. Get Authenticated User ID
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed or user ID missing"})
		return
	}

	// 2. Get Group ID from URL path
	groupIDStr := c.Param("groupID")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// 3. Authorization Check: Ensure user is a member of the group
	members, err := h.GroupService.GetMembers(c.Request.Context(), groupID)
	if err != nil {
		// This could be because the group doesn't exist, or a DB error.
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found or error checking membership"})
		return
	}

	isMember := false
	for _, memberID := range members {
		if memberID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this group"})
		return
	}

	// 4. Get Optional Limit and BeforeID Query Parameters
	limit := defaultHistoryLimit
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultHistoryLimit))
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	beforeIDStr := c.DefaultQuery("before_id", "0")
	beforeID, _ := strconv.ParseInt(beforeIDStr, 10, 64)

	// 5. Call Domain Service to retrieve history
	messages, err := h.MessageService.GetGroupConversationHistory(c.Request.Context(), groupID, limit, beforeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve group message history"})
		return
	}

	// 6. Success Response
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}
