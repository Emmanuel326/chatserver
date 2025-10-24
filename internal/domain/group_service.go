package domain

import (
	"context"
	"errors"
	"fmt"  
	"time" 
)

// groupService is the concrete implementation of the GroupService interface.
type groupService struct {
	groupRepo GroupRepository
	userRepo  UserRepository 
}

// NewGroupService creates a new instance of the GroupService.
func NewGroupService(groupRepo GroupRepository, userRepo UserRepository) GroupService {
	return &groupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// CreateGroup creates a new group and relies on the repository to handle adding the owner as the first member.
func (s *groupService) CreateGroup(ctx context.Context, name string, ownerID int64) (*Group, error) { 
	// 1. Create the Group structure. The CreatedAt field is omitted here because your 
	//    SQLite implementation expects the DB to handle setting it implicitly upon creation.
	group := &Group{
		Name:    name,
		OwnerID: ownerID,
	}

	// 2. Call the repository's Create method. 
	//    Based on your SQLite repo, this method should return a Group with ID and CreatedAt populated, 
	//    and should internally add the owner as a member.
	newGroup, err := s.groupRepo.Create(ctx, group) // <-- CORRECT CALL (s.groupRepo.Create is the method)
	if err != nil {
		return nil, fmt.Errorf("failed to save group: %w", err)
	}

	return newGroup, nil
}

// AddMember adds a user to a specified group.
func (s *groupService) AddMember(ctx context.Context, groupID, userID int64, inviterID int64) error {
	// 1. Validation: Ensure the user being added exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return errors.New("user to be added does not exist")
	}

	// 2. Validation: Ensure the group exists 
	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil { 
		return errors.New("group does not exist")
	}

	// 3. Add member via repository (must create the struct required by the interface)
	member := &GroupMember{
		GroupID: groupID,
		UserID:  userID,
		IsAdmin: false, 
		JoinedAt: time.Now(), 
	} 
	
	if err := s.groupRepo.AddMember(ctx, member); err != nil { 
		return fmt.Errorf("failed to add user %d to group %d: %w", userID, groupID, err)
	}
	
	return nil
}

// GetMembers retrieves a list of User IDs belonging to a group.
func (s *groupService) GetMembers(ctx context.Context, groupID int64) ([]int64, error) {
	if groupID == 0 {
		return nil, errors.New("group ID cannot be zero")
	}
	return s.groupRepo.FindMembersByGroupID(ctx, groupID) } 
