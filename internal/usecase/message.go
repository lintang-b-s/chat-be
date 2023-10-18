package usecase

import (
	"context"
	"fmt"
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
		return entity.PrivateChatUsers{}, fmt.Errorf("MessageuseCase - GetMessageByUserLogin -uc.userPgRepo.GetUserByUsername: %w", err)
	}

	pcReq := entity.GetPrivateChatQueryByUserRequest{
		UserId: user.Id,
	}
	pc, err := uc.pcRepo.GetPrivateChatByUser(pcReq)
	if err != nil {
		return entity.PrivateChatUsers{}, fmt.Errorf("MessageuseCase - GetMessageByUserLogin - uc.pcRepo.GetPrivateChatByUser: %w", err)
	}

	return pc, nil
}

func (uc *MessageuseCase) GetMessagesByRecipient(ctx context.Context, e entity.GetPCBySdrAndRcvrRequest) (entity.PrivateChats, error) {
	sender, err := uc.userPgRepo.GetUserByUsername(e.SenderUsername)
	if err != nil {
		return entity.PrivateChats{}, fmt.Errorf("MessageuseCase - GetMessagesByRecipient - uc.userPgRepo.GetUserByUsername: %w", err)
	}
	receiver, err := uc.userPgRepo.GetUserByUsername(e.ReceiverUsername)
	if err != nil {
		return entity.PrivateChats{}, fmt.Errorf("MessageuseCase - GetMessagesByRecipient - uc.userPgRepo.GetUserByUsername: %w", err)
	}

	pcReq := entity.GetPCQueryBySdrAndRcvrRequest{
		SenderId:   sender.Id,
		ReceiverId: receiver.Id,
	}

	pcs, err := uc.pcRepo.GetPrivateChatBySenderAndReceiver(pcReq)
	if err != nil {
		return entity.PrivateChats{}, err
	}

	return pcs, nil
}
