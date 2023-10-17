package usecase

import (
	"context"
	"github.com/lintangbs/chat-be/internal/entity"
)

type MessageuseCase struct {
	pcRepo     PrivateChatRepo
	userPgRepo UserRepo
}

func NewMessageuseCase(pcRepo PrivateChatRepo, upg UserRepo) *MessageuseCase {
	return &MessageuseCase{
		pcRepo:     pcRepo,
		userPgRepo: upg,
	}
}

func (uc *MessageuseCase) GetMessageByUserLogin(ctx context.Context, e entity.GetPrivateChatByUserRequest) (entity.PrivateChatUsers, error) {
	user, err := uc.userPgRepo.GetUserByUsername(e.Username)
	if err != nil {
		return entity.PrivateChatUsers{}, err
	}

	pcReq := entity.GetPrivateChatQueryByUserRequest{
		UserId: user.Id,
	}
	pc, err := uc.pcRepo.GetPrivateChatByUser(pcReq)
	if err != nil {
		return entity.PrivateChatUsers{}, err
	}

	return pc, nil
}
