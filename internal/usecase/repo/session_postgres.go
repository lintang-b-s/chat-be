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
	Username     string
	RefreshToken string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db}
}

// CreateSession insert session /refrsh token baru ke database
func (r *SessionRepo) CreateSession(ctx context.Context, c entity.CreateSessionRequest) (entity.Session, error) {

	createSession := Session{
		ID:           c.ID,
		Username:     c.Username,
		RefreshToken: c.RefreshToken,
		ExpiresAt:    c.ExpiresAt,
	}
	if result := r.db.Create(&createSession); result.Error != nil {
		//	 internal server error
		return entity.Session{}, fmt.Errorf("SessionRepo - r.db.Create: %w", result.Error)
	}
	session := entity.Session{
		ID:           createSession.ID,
		Username:     createSession.Username,
		RefreshToken: createSession.RefreshToken,
		ExpiresAt:    createSession.ExpiresAt,
	}
	return session, nil
}

// GetSession mendapatkan sesssion/refresh token dari database
func (r *SessionRepo) GetSession(ctx context.Context, refreshTokkenId uuid.UUID) (entity.Session, error) {
	var sessionDb Session

	result := r.db.Where(&Session{ID: refreshTokkenId}).First(&sessionDb)

	if err := result.Error; err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - GetSession- r.db.Where(&Session{ID: refreshTokkenId}).First(&sessionDb): %w", err)
	}

	session := entity.Session{
		ID:           sessionDb.ID,
		Username:     sessionDb.Username,
		RefreshToken: sessionDb.RefreshToken,
		ExpiresAt:    sessionDb.ExpiresAt,
		CreatedAt:    sessionDb.CreatedAt,
	}

	return session, nil
}

func (r *SessionRepo) DeleteSession(ctx context.Context, uuid uuid.UUID) error {
	var sessionDb Session
	result := r.db.Where(&Session{ID: uuid}).Delete(&sessionDb)
	if err := result.Error; err != nil {
		return fmt.Errorf("Sessionrepo - r.db.Where(&Session{ID: uuid}).Delete(&sessionDb) - %w", err)
	}
	return nil
}
