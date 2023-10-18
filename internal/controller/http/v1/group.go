package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	api "github.com/lintangbs/chat-be/internal/middleware"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"github.com/lintangbs/chat-be/pkg/logger"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type groupRoutes struct {
	g   usecase.Group
	l   logger.Interface
	jwt jwt.JwtTokenMaker
}

func newGroupRoutes(handler *gin.RouterGroup, g usecase.Group, l logger.Interface, jwt jwt.JwtTokenMaker) {
	r := &groupRoutes{g, l, jwt}

	h := handler.Group("/groups").Use(api.AuthMiddleware(r.jwt))
	{
		h.POST("", r.createGroup)
		h.PUT("/add", r.addNewGroupMember)
		h.PUT("/remove", r.removeGroupMember)
	}
}

type createGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type groupResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// @Summary     create new group
// @Description        create new group
// @ID          createNewgroup
// @Tags  	    group
// @Accept      json
// @Produce     json
// @Security OAuth2Application
// @Param       request body createGroupRequest true "set up new group"
// @Success     200 {object} groupResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/groups [post]
// Author: https://github.com/lintang-b-s
func (r *groupRoutes) createGroup(c *gin.Context) {
	var request createGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - createGroup")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)
	group, err := r.g.CreateGroup(
		c.Request.Context(),
		entity.CreateGroupReqUc{
			Name:     request.Name,
			UserName: authPayload.Username,
			Members:  request.Members,
		},
	)

	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		errRepo := errors.Unwrap(unwrapedErr)
		if errRepo == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}
		if errRepo == repo.ErrNotFoundContactErr {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}
		if errRepo == repo.GroupAlreadyExistsErr {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}

		r.l.Error("http - v1- createGroup")
		ErrorResponse(c, http.StatusInternalServerError, "createGroup service problems: "+err.Error())
		return
	}

	res := groupResponse{
		Id:        group.Id,
		Name:      group.Name,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}

	c.JSON(http.StatusCreated, res)
}

// AddNewMemberReqUc request in usecasee
type addNewGroupMemberRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// @Summary     add new group member
// @Description     add new group member
// @ID          addNewGroupMember
// @Tags  	    group
// @Accept      json
// @Produce     json
// @Security OAuth2Application
// @Param       request body addNewGroupMemberRequest true "set up new group"
// @Success     200 {object} groupResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/groups/add [put]
// Author: https://github.com/lintang-b-s
func (r *groupRoutes) addNewGroupMember(c *gin.Context) {
	var request addNewGroupMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - addNewGroupMember")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	group, err := r.g.AddNewGroupMember(
		c.Request.Context(),
		entity.AddNewGroupMemberReqUc{
			Name:     request.Name,
			UserName: authPayload.Username,
			Members:  request.Members,
		},
	)
	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		errRepo := errors.Unwrap(unwrapedErr)
		if errRepo == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}

		if errRepo == repo.ErrNotFoundContactErr {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}

		if errRepo == repo.UserAlreadyMembersErr {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}

		r.l.Error("http - v1- addNewGroupMember")
		ErrorResponse(c, http.StatusInternalServerError, "addNewGroupMember service problems: "+err.Error())
		return
	}

	res := groupResponse{
		Id:        group.Id,
		Name:      group.Name,
		CreatedAt: group.CreatedAt,
	}

	c.JSON(http.StatusCreated, res)
}

type removeGroupMember struct {
	Name         string `json:"name"`
	UsertoRemove string `json:"userto_remove"`
}

// @Summary    remove group member
// @Description    remove group member
// @ID          removeGroupMember
// @Tags  	    group
// @Accept      json
// @Produce     json
// @Security OAuth2Application
// @Param       request body removeGroupMember true "set up removeGroupMember"
// @Success     200 {object} groupResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/groups/remove [put]
// Author: https://github.com/lintang-b-s
func (r *groupRoutes) removeGroupMember(c *gin.Context) {
	var request removeGroupMember
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - removeGroupMember")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	group, err := r.g.RemoveGroupMember(
		c.Request.Context(),
		entity.RemoveGroupMemberReqUc{
			Name:         request.Name,
			UserName:     authPayload.Username,
			UsertoRemove: request.UsertoRemove,
		},
	)
	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		errRepo := errors.Unwrap(unwrapedErr)
		if errRepo == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}

		r.l.Error("http - v1- removeGroupMember")
		ErrorResponse(c, http.StatusInternalServerError, "removeGroupMember service problems: "+err.Error())
		return
	}

	res := groupResponse{
		Id:        group.Id,
		Name:      group.Name,
		CreatedAt: group.CreatedAt,
	}
	c.JSON(http.StatusCreated, res)
}
