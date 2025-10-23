package sqlite

import (
	"context"
	"database/sql"
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

// Create persists a new group to the database and adds the owner as the first member.
func (r *GroupRepository) Create(ctx context.Context, group *domain.Group) (*domain.Group, error) {
	query := `
		INSERT INTO groups (name, owner_id)
		VALUES (:name, :owner_id)
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
	group.CreatedAt = time.Now() // Approximate, DB handles precision

	// Automatically add the owner as the first member and an admin
	member := &domain.GroupMember{
		GroupID: id,
		UserID:  group.OwnerID,
		IsAdmin: true,
	}
	if err := r.AddMember(ctx, member); err != nil {
		// Log error, but return group since it was created
		return group, err 
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
	query := `
		INSERT INTO group_members (group_id, user_id, is_admin)
		VALUES (:group_id, :user_id, :is_admin)
	`
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
