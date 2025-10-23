package domain

import (
	"context"
	"errors"
)

// GroupService is the concrete implementation of the GroupService interface.
type groupService struct {
	groupRepo GroupRepository
	userRepo  UserRepository // Dependency needed to verify users exist
}

// NewGroupService creates a new instance of the GroupService.
func NewGroupService(groupRepo GroupRepository, userRepo UserRepository) GroupService {
	return &groupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// CreateGroup creates a new group and automatically adds the owner as the first admin member.
func (s *groupService) CreateGroup(ctx context.Context, name string, ownerID int64) (*Group, error) {
	// 1. Basic validation
	if name == "" || ownerID == 0 {
		return nil, errors.New("group name and owner ID cannot be empty")
	}

	// 2. Create the Group (Repository automatically adds the owner as the first member)
	group := &Group{
		Name:    name,
		OwnerID: ownerID,
	}
	
	return s.groupRepo.Create(ctx, group)
}

// AddMember adds a user to a specified group.
func (s *groupService) AddMember(ctx context.Context, groupID, userID int64, inviterID int64) error {
	// 1. Validation: Ensure the user being added exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return errors.New("user to be added does not exist")
	}

	// 2. Validation: Ensure the inviter (who is making the API call) is an admin or the owner
	// NOTE: For simplicity now, we skip admin check. In production, you'd verify inviterID permissions.
	
	// 3. Add member via repository
	member := &GroupMember{
		GroupID: groupID,
		UserID:  userID,
		IsAdmin: false,
	}
	return s.groupRepo.AddMember(ctx, member)
}

// GetMembers retrieves a list of User IDs belonging to a group.
func (s *groupService) GetMembers(ctx context.Context, groupID int64) ([]int64, error) {
	if groupID == 0 {
		return nil, errors.New("group ID cannot be zero")
	}
	return s.groupRepo.FindMembersByGroupID(ctx, groupID)
}

// TODO:
