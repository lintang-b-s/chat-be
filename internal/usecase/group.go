package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
)

type GroupUseCase struct {
	gRepo GroupRepo
	uRepo UserRepo
}

func NewGroupUseCase(gRepo GroupRepo, uRepo UserRepo) *GroupUseCase {
	return &GroupUseCase{
		gRepo: gRepo,
		uRepo: uRepo,
	}
}

func (uc *GroupUseCase) CreateGroup(ctx context.Context, e entity.CreateGroupReqUc) (entity.Group, error) {
	userLogin, err := uc.uRepo.GetUserByUsername(e.UserName)
	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - CreateGroup - uc.uRepo.GetUserByUsername : %w", err)
	}

	var membersId []uuid.UUID

	for _, memberName := range e.Members {
		isFriendErr := uc.uRepo.GetUserFriend(ctx, userLogin.Username, memberName)
		if isFriendErr != nil {
			return entity.Group{}, fmt.Errorf("GroupUseCase - CreateGroup -  uc.uRepo.GetUserFriend : %w", isFriendErr)
		}

		member, err := uc.uRepo.GetUserByUsername(memberName)
		if err != nil {
			return entity.Group{}, fmt.Errorf("GroupUseCase - CreateGroup - uc.uRepo.GetUserByUsername : %w", err)
		}
		membersId = append(membersId, member.Id)
	}
	createReq := entity.CreateGroupRequest{
		Name:    e.Name,
		UserId:  userLogin.Id,
		Members: membersId,
	}
	group, err := uc.gRepo.CreateGroup(ctx, createReq)

	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - CreateGroup - uc.gRepo.CreateGroup : %w", err)
	}

	return group, nil
}

func (uc *GroupUseCase) AddNewGroupMember(ctx context.Context, e entity.AddNewGroupMemberReqUc) (entity.Group, error) {
	userLogin, err := uc.uRepo.GetUserByUsername(e.UserName)
	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember - uc.uRepo.GetUserByUsername : %w", err)
	}

	var membersId []uuid.UUID

	for _, memberName := range e.Members {
		isFriendErr := uc.uRepo.GetUserFriend(ctx, userLogin.Username, memberName)
		if isFriendErr != nil {
			return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember -  uc.uRepo.GetUserFriend : %w", isFriendErr)
		}
		member, err := uc.uRepo.GetUserByUsername(memberName)
		if err != nil {
			return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember - uc.uRepo.GetUserByUsername : %w", err)
		}
		membersId = append(membersId, member.Id)
	}

	addReq := entity.AddNewGroupMemberReq{
		Name:    e.Name,
		UserId:  userLogin.Id,
		Members: membersId,
	}

	group, err := uc.gRepo.AddNewGroupMember(ctx, addReq)
	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember - uc.gRepo.AddNewGroupMember: %w", err)
	}

	return group, nil
}

func (uc *GroupUseCase) RemoveGroupMember(ctx context.Context, e entity.RemoveGroupMemberReqUc) (entity.Group, error) {
	userLogin, err := uc.uRepo.GetUserByUsername(e.UserName)
	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember - uc.uRepo.GetUserByUsername : %w", err)
	}
	userToRemove, err := uc.uRepo.GetUserByUsername(e.UsertoRemove)
	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember - uc.uRepo.GetUserByUsername : %w", err)
	}
	removeReq := entity.RemoveGroupMemberReq{
		Name:   e.Name,
		UserId: userLogin.Id,
		Member: userToRemove.Id,
	}

	group, err := uc.gRepo.RemoveMember(ctx, removeReq)
	if err != nil {
		return entity.Group{}, fmt.Errorf("GroupUseCase - AddNewGroupMember - uc.gRepo.RemoveMember: %w", err)
	}

	return group, nil
}
