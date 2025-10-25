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
	var req SendGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
    
    // VALIDATION: Ensure at least Content or MediaURL is present
    if req.Content == "" && req.MediaURL == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Message must contain either text content or a media URL"})
        return
    }

	// 4. Call MessageService to send the message (now includes MediaURL)
    // FIX: Passing req.MediaURL to the service layer
	_, err = h.MessageService.SendGroupMessage(c.Request.Context(), senderID, groupID, req.Content, req.MediaURL)
	
	if err != nil {
		// Differentiate between domain errors (e.g., membership) and server errors
		if _, ok := err.(*domain.NotFoundError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Failed to send group message", "details": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send group message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully to group"})
}

// SendP2PMessage handles sending a new message to a specific user.
// POST /v1/messages/p2p/:recipientID
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
	// NOTE: We rely on the SendMessageRequest struct defined elsewhere (e.g., in models.go or at the package level)
	var req SendGroupMessageRequest // Re-use the existing request structure
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// VALIDATION: Ensure at least Content or MediaURL is present
	if req.Content == "" && req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message must contain either text content or a media URL"})
		return
	}

	// 4. Call MessageService to send the message (includes MediaURL)
	_, err = h.MessageService.SendP2PMessage(c.Request.Context(), senderID, recipientID, req.Content, req.MediaURL)
	
	if err != nil {
		// Differentiate between domain errors (e.g., user not found) and server errors
		if err.Error() == "recipient user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipient user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send P2P message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "P2P message sent successfully"})
}
