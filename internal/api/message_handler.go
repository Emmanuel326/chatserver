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
}

// NewMessageHandler creates a new handler instance.
func NewMessageHandler(messageService domain.MessageService) *MessageHandler {
	return &MessageHandler{
		MessageService: messageService,
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
	
	// 4. Call Domain Service to retrieve history
	// FIX: Use request context instead of context.Background()
	messages, err := h.MessageService.GetConversationHistory(c.Request.Context(), senderID, recipientID, limit)
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

// SendGroupMessage handles sending a new message to a specific group.
// POST /v1/messages/group/:groupID
func (h *MessageHandler) SendGroupMessage(c *gin.Context) { // <--- NEW METHOD IMPLEMENTATION
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

	// 3. Parse request body
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: message content is required"})
		return
	}

	// 4. Call MessageService to send the message (handles saving and broadcasting)
	_, err = h.MessageService.SendGroupMessage(c.Request.Context(), senderID, groupID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send group message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully to group"})
}
