package api

import (
	"context"
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
	// We use a separate context with timeout if needed, but for now, use the request context.
	messages, err := h.MessageService.GetConversationHistory(context.Background(), senderID, recipientID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve message history"})
		return
	}

	// 5. Success Response
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}
