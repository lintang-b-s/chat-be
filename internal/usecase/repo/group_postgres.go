package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
	"time"
)

var (
	UserNotMemberErr      = errors.New("users tidak termasuk dalam group")
	GroupAlreadyExistsErr = errors.New("group with same username already exists")
	UserAlreadyMembersErr = errors.New("users already in group ")
)

type GroupRepo struct {
	db *gorm.DB
}

type Group struct {
	gorm.Model
	Id        uuid.UUID
	Name      string
	Members   []UsersGroup
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UsersGroup struct {
	gorm.Model
	Id      uuid.UUID
	UserId  uuid.UUID
	GroupId uuid.UUID
}

// Reename table
type Tabler interface {
	TableName() string
}

// TableName overrides the table name used by UsersGroup to `users_group`
func (UsersGroup) TableName() string {
	return "users_group"
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db}
}

// CreateGroup membuat group baru
func (r *GroupRepo) CreateGroup(ctx context.Context, e entity.CreateGroupRequest) (entity.Group, error) {
	groupId := uuid.New()

	var ug []UsersGroup

	var grp []Group

	// group dg username ada di db
	r.db.Where(&Group{Name: e.Name}).Find(&grp)
	if len(grp) != 0 {
		return entity.Group{}, fmt.Errorf("GroupRepo - CreateGroup -  r.db.Create: %w", GroupAlreadyExistsErr)
	}

	//Create Group
	g := &Group{Id: groupId, Name: e.Name}
	if result := r.db.Create(&g); result.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - CreateGroup -  r.db.Create: %w", result.Error)
	}

	// Add group Members

	userG := UsersGroup{Id: uuid.New(), UserId: e.UserId, GroupId: g.Id}

	if result := r.db.Create(&userG); result.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - CreateGroup -  r.db.Create: %w", result.Error)
	}

	ug = append(ug, userG)

	for _, memberId := range e.Members {
		userG = UsersGroup{Id: uuid.New(), UserId: memberId, GroupId: g.Id}
		//create new group and insert user to group
		if result := r.db.Create(&userG); result.Error != nil {
			return entity.Group{}, fmt.Errorf("GroupRepo - CreateGroup -  r.db.Create: %w", result.Error)
		}
		ug = append(ug, userG)
	}
	g.Members = ug
	if res := r.db.Save(&g); res.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - CreateGroup -r.db.Save(&g)", res.Error)

	}

	res := entity.Group{
		Id:        groupId,
		Name:      e.Name,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}

	return res, nil
}

func (r *GroupRepo) AddNewGroupMember(ctx context.Context, e entity.AddNewGroupMemberReq) (entity.Group, error) {
	var group Group

	if res := r.db.Where(&Group{Name: e.Name}).Joins("LEFT JOIN users_group on users_group.group_id=groups.id").Preload("Members").First(&group); res.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", res.Error)
	}

	uMap := make(map[string]bool)

	isMember := false
	// cek apakah user yg login termasuk member dari groupnya
	for _, member := range group.Members {
		if member.UserId == e.UserId {
			isMember = true
		}
		uMap[member.UserId.String()] = true
	}

	if isMember == false {
		return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", UserNotMemberErr)
	}

	for _, memToRegisterId := range e.Members {
		memToRegisteridStr := memToRegisterId.String()
		if _, prs := uMap[memToRegisteridStr]; prs == true {
			return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", UserAlreadyMembersErr)
		}
	}

	for _, memberId := range e.Members {
		group.Members = append(group.Members, UsersGroup{Id: uuid.New(), UserId: memberId})
	}

	if res := r.db.Save(&group); res.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember - r.db.Save(&group): %w", res.Error)
	}

	res := entity.Group{
		Id:        group.Id,
		Name:      e.Name,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}

	return res, nil
}

func (r *GroupRepo) RemoveMember(ctx context.Context, e entity.RemoveGroupMemberReq) (entity.Group, error) {
	var group Group
	if res := r.db.Where(&Group{Name: e.Name}).Joins("LEFT JOIN users_group on users_group.group_id=groups.id").Preload("Members").First(&group); res.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", res.Error)
	}

	isMember := false
	for _, member := range group.Members {
		if member.UserId == e.UserId {
			isMember = true
		}
	}

	if isMember == false {
		return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", UserNotMemberErr)
	}

	if result := r.db.Exec("DELETE FROM users_group WHERE user_id=" + "'" + e.Member.String() + "'" + " AND group_id='" + group.Id.String() + "';"); result.RowsAffected == 0 {
		return entity.Group{}, fmt.Errorf("GroupRepo - RemoveMember - r.db.Save(&group): rows already deleted")
	}

	res := entity.Group{
		Id:        group.Id,
		Name:      e.Name,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
	return res, nil
}

func (r *GroupRepo) GetGroupMembers(groupId uuid.UUID, userId uuid.UUID) (entity.Group, error) {
	var group Group
	if res := r.db.Where(&Group{Id: groupId}).Joins("LEFT JOIN users_group on users_group.group_id=groups.id").Preload("Members").First(&group); res.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - AddNewGroupMember -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", res.Error)
	}

	// cek apakah user yg login termasuk member dari groupnya
	uMap := make(map[string]bool)
	isMember := false
	for _, member := range group.Members {
		if member.UserId == userId {
			isMember = true
		}
		uMap[member.UserId.String()] = true
	}

	if isMember == false {
		return entity.Group{}, fmt.Errorf("GroupRepo - GetGroupMembers -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", UserNotMemberErr)
	}

	var members []uuid.UUID
	for _, member := range group.Members {
		members = append(members, member.UserId)
	}

	groupRes := entity.Group{
		Id:      group.Id,
		Name:    group.Name,
		Members: members,
	}

	return groupRes, nil

}

func (r *GroupRepo) GetGroupByName(groupName string, userId uuid.UUID) (entity.Group, error) {
	var group Group
	if res := r.db.Where(&Group{Name: groupName}).Joins("LEFT JOIN users_group on users_group.group_id=groups.id").Preload("Members").First(&group); res.Error != nil {
		return entity.Group{}, fmt.Errorf("GroupRepo - GetGroupByName -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", res.Error)
	}

	// cek apakah user yg login termasuk member dari groupnya
	uMap := make(map[string]bool)
	isMember := false
	for _, member := range group.Members {
		if member.UserId == userId {
			isMember = true
		}
		uMap[member.UserId.String()] = true
	}

	if isMember == false {
		return entity.Group{}, fmt.Errorf("GroupRepo - GetGroupByName -  r.db.Where(&Group{Name: e.Name}).First(&group): %w", UserNotMemberErr)
	}

	groupRes := entity.Group{
		Id:   group.Id,
		Name: group.Name,
	}
	return groupRes, nil
}
