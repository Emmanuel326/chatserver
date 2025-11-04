package sqlite

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/jmoiron/sqlx"
)

// GroupRepository implements the domain.GroupRepository interface using SQLite.
type GroupRepository struct {
	db *sqlx.DB
}

// NewGroupRepository creates a new GroupRepository instance.
func NewGroupRepository(db *sqlx.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create persists a new group to the database and adds the owner as the first member
func (r *GroupRepository) Create(ctx context.Context, group *domain.Group) (*domain.Group, error) {
	// Set the CreatedAt timestamp explicitly before saving
	if group.CreatedAt.IsZero() {
		group.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO groups (name, owner_id, created_at)
		VALUES (:name, :owner_id, :created_at)
	`

	res, err := r.db.NamedExecContext(ctx, query, group)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	group.ID = id
	
	// Automatically add the owner as the first member and an admin
	member := &domain.GroupMember{
		GroupID: id,
		UserID:  group.OwnerID,
		IsAdmin: true,
		JoinedAt: time.Now(), // Ensure JoinedAt is set for the internal AddMember call
	}
	// AddMember will use the JoinedAt set here.
	if err := r.AddMember(ctx, member); err != nil {
		log.Printf("Warning: Failed to add owner %d as member of group %d: %v", group.OwnerID, id, err)
	}

	return group, nil
}
// FindByID retrieves a group by its ID.
func (r *GroupRepository) FindByID(ctx context.Context, groupID int64) (*domain.Group, error) {
	group := &domain.Group{}
	query := `SELECT id, name, owner_id, created_at FROM groups WHERE id = ?`
	
	err := r.db.GetContext(ctx, group, query, groupID)
	if err == sql.ErrNoRows {
		return nil, nil // Group not found
	}
	return group, err
}

// AddMember adds a user to a group.
func (r *GroupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	// Ensure JoinedAt is set before insertion if the service layer missed it
	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}

	query := `
		INSERT INTO group_members (group_id, user_id, is_admin, joined_at)
		VALUES (:group_id, :user_id, :is_admin, :joined_at)
	`
	// The fix: explicitly including joined_at in the INSERT query
	
	_, err := r.db.NamedExecContext(ctx, query, member)
	return err
}

// FindMembersByGroupID retrieves the IDs of all users belonging to a group.
func (r *GroupRepository) FindMembersByGroupID(ctx context.Context, groupID int64) ([]int64, error) {
	var userIDs []int64
	query := `SELECT user_id FROM group_members WHERE group_id = ?`

	err := r.db.SelectContext(ctx, &userIDs, query, groupID)
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

// FindGroupsByUserID retrieves all groups a specific user is a member of.
func (r *GroupRepository) FindGroupsByUserID(ctx context.Context, userID int64) ([]*domain.Group, error) {
	var groups []*domain.Group
	query := `
		SELECT g.id, g.name, g.owner_id, g.created_at
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
	`
	err := r.db.SelectContext(ctx, &groups, query, userID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}
