package repo

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
	"time"
)

type SessionRepo struct {
	db *gorm.DB
}

type Session struct {
	gorm.Model
	ID           uuid.UUID
	Email        string
	RefreshToken string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db}
}

func (r *SessionRepo) CreateSession(ctx context.Context, c entity.CreateSessionRequest) (entity.Session, error) {

	createSession := Session{
		ID:           c.ID,
		Email:        c.Email,
		RefreshToken: c.RefreshToken,
		ExpiresAt:    c.ExpiresAt,
	}
	if result := r.db.Create(&createSession); result.Error != nil {
		//	 internal server error
		return entity.Session{}, fmt.Errorf("SessionRepo - r.db.Create")
	}
	session := entity.Session{
		ID:           createSession.ID,
		Email:        createSession.Email,
		RefreshToken: createSession.RefreshToken,
		ExpiresAt:    createSession.ExpiresAt,
	}
	return session, nil
}

func (r *SessionRepo) GetSession(ctx context.Context, refreshTokkenId uuid.UUID) (entity.Session, error) {
	var sessionDb Session

	result := r.db.Where(&Session{ID: refreshTokkenId}).First(&sessionDb)

	if err := result.Error; err != nil {
		return entity.Session{}, err
	}

	session := entity.Session{
		ID:           sessionDb.ID,
		Email:        sessionDb.Email,
		RefreshToken: sessionDb.RefreshToken,
		ExpiresAt:    sessionDb.ExpiresAt,
		CreatedAt:    sessionDb.CreatedAt,
	}

	return session, nil
}
