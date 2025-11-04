package domain

import (
	"context"
	"time"
)

// Group defines the core structure for a chat group.
type Group struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	OwnerID   int64     `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GroupMember defines the relationship between a user and a group.
type GroupMember struct {
	GroupID   int64     `json:"group_id" db:"group_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
}

// ---------------------------------------------
// GROUP INTERFACES (Contracts for Group Management)
// ---------------------------------------------

// GroupService defines the business operations related to groups.
type GroupService interface {
	CreateGroup(ctx context.Context, name string, ownerID int64) (*Group, error)
	AddMember(ctx context.Context, groupID, userID int64, inviterID int64) error
	GetMembers(ctx context.Context, groupID int64) ([]int64, error)
	GetGroupsForUser(ctx context.Context, userID int64) ([]*Group, error)
}

// GroupRepository defines the data access operations for groups and membership.
type GroupRepository interface {
	Create(ctx context.Context, group *Group) (*Group, error)
	FindByID(ctx context.Context, groupID int64) (*Group, error)
	AddMember(ctx context.Context, member *GroupMember) error
	FindMembersByGroupID(ctx context.Context, groupID int64) ([]int64, error)
	FindGroupsByUserID(ctx context.Context, userID int64) ([]*Group, error)
}
