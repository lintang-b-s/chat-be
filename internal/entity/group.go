package entity

import (
	"github.com/google/uuid"
	"time"
)

// CreateGroupRequest membuat group chat baru
type CreateGroupRequest struct {
	Name    string      `json:"name"`
	UserId  uuid.UUID   `json:"user_id"`
	Members []uuid.UUID `json:"members"`
}

// Group
type Group struct {
	Id        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Members   []uuid.UUID `json:"members"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

// AddNewGroupMemberReq menamahkan member group chat baru
type AddNewGroupMemberReq struct {
	Name    string      `json:"name"`
	UserId  uuid.UUID   `json:"user_id"`
	Members []uuid.UUID `json:"members"`
}

// RemoveGroupMemberReq menghapus member dari group
type RemoveGroupMemberReq struct {
	Name   string    `json:"name"`
	UserId uuid.UUID `json:"user_id"`
	Member uuid.UUID `json:"member"`
}

// creategroup request in usecase
type CreateGroupReqUc struct {
	Name     string   `json:"name"`
	UserName string   `json:"user_id"`
	Members  []string `json:"members"`
}

// AddNewMemberReqUc request in usecasee
type AddNewGroupMemberReqUc struct {
	Name     string   `json:"name"`
	UserName string   `json:"user_id"`
	Members  []string `json:"members"`
}

type RemoveGroupMemberReqUc struct {
	Name         string `json:"name"`
	UserName     string `json:"user_id"`
	UsertoRemove string `json:"userto_remove"`
}
