package api

import (
	"log" 
	"net/http"
	"strconv"

	"github.com/Emmanuel326/chatserver/internal/api/middleware"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/gin-gonic/gin"
)

// GroupHandler handles HTTP requests related to chat groups.
type GroupHandler struct {
	GroupService domain.GroupService
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(groupService domain.GroupService) *GroupHandler {
	return &GroupHandler{GroupService: groupService}
}

// CreateGroup handles the creation of a new chat group.
// POST /v1/groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	// 1. Get authenticated UserID from context (set by AuthMiddleware)
	userID, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	ownerID := userID.(int64)

	// 2. Parse request body
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: group name is required"})
		return
	}

	// 3. Call GroupService to create the group
	group, err := h.GroupService.CreateGroup(c.Request.Context(), req.Name, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Group created successfully",
		"group_id": group.ID,
		"name": group.Name,
		"owner_id": group.OwnerID,
	})
}

// AddMember handles adding a user to an existing group.
// POST /v1/groups/:groupID/members
func (h *GroupHandler) AddMember(c *gin.Context) {
	// 1. Get authenticated user ID (the inviter)
	inviterIDValue, exists := c.Get(middleware.ContextUserIDKey) 
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inviter not authenticated"})
		return
	}
	inviterID, ok := inviterIDValue.(int64)
	if !ok {
		log.Printf("Error: Inviter ID in context is not int64: %v", inviterIDValue)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "internal auth error"})
		return
	}
	
	// 2. Get Group ID from URL parameter
	groupIDStr := c.Param("groupID")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	// 3. Parse request body (User ID to be added)
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: user_id to add is required"})
		return
	}

	// 4. Call GroupService to add the member
	err = h.GroupService.AddMember(c.Request.Context(), groupID, req.UserID, inviterID)
	if err != nil {
		// Note: GroupService will handle checks like user existence or group existence
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

// GetMembers handles the retrieval of members for a specific group.
// GET /v1/groups/:groupID/members
func (h *GroupHandler) GetMembers(c *gin.Context) {
	// 1. Get Group ID from URL parameter
	groupIDStr := c.Param("groupID")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	// 2. Call GroupService to get members
	members, err := h.GroupService.GetMembers(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve group members", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_id": groupID, "members": members})
}
