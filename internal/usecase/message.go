package usecase

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
)

type MessageuseCase struct {
	pcRepo     PrivateChatRepo
	userPgRepo UserRepo
	gcRepo     GroupChatRepo
	gpRepo     GroupRepo
}

func NewMessageuseCase(pcRepo PrivateChatRepo, upg UserRepo, gcRepo GroupChatRepo, gpRepo GroupRepo) *MessageuseCase {
	return &MessageuseCase{
		pcRepo:     pcRepo,
		userPgRepo: upg,
		gcRepo:     gcRepo,
		gpRepo:     gpRepo,
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

func (uc *MessageuseCase) GetMessagesByGroupChat(ctx context.Context, e entity.GroupChatMsgRequest) (entity.GroupChatMessages, error) {
	user, err := uc.userPgRepo.GetUserByUsername(e.UserName)
	if err != nil {
		return entity.GroupChatMessages{}, fmt.Errorf("MessageuseCase - GetMessagesByGroupChat - uc.userPgRepo.GetUserByUsername: %w", err)
	}
	group, err := uc.gpRepo.GetGroupByName(e.GroupName, user.Id)
	if err != nil {
		return entity.GroupChatMessages{}, fmt.Errorf("MessageuseCase - GetMessagesByGroupChat - uc.gpRepo.GetGroupByName: %w", err)
	}

	gcMessages, err := uc.gcRepo.GetMessagesByGroupId(group.Id)

	return gcMessages, nil
}
