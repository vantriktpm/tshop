package postgres

import (
	"context"
	"time"

	userdb "github.com/tshop/backend/services/user-service/database"
	"github.com/tshop/backend/services/user-service/internal/domain"
	"gorm.io/gorm"
)

const selectUserBase = `
SELECT
  id,
  user_name,
  full_name,
  phone,
  password_hash,
  salt,
  status,
  is_verified,
  user_id,
  provider,
  provider_user_id,
  access_token,
  password_changed_at,
  refresh_token,
  token_version,
  created_by,
  updated_by,
  created_at,
  updated_at,
  expires_at
FROM service.users
`

const (
	selectUserByIDQuery                        = selectUserBase + " WHERE id = ?"
	selectUserByUserNameQuery                  = selectUserBase + " WHERE user_name = ?"
	selectUserByProviderAndProviderUserIDQuery = selectUserBase + " WHERE provider = ? AND provider_user_id = ?"
)

type UserModel struct {
	ID                string     `gorm:"column:id;primaryKey;type:uuid"`
	UserName          *string    `gorm:"column:user_name;size:255"`
	FullName          *string    `gorm:"column:full_name;size:255"`
	Phone             *string    `gorm:"column:phone;size:255"`
	PasswordHash      *string    `gorm:"column:password_hash;type:text"`
	Salt              *string    `gorm:"column:salt;size:255"`
	Status            *int16     `gorm:"column:status"`
	IsVerified        *bool      `gorm:"column:is_verified"`
	UserID            *string    `gorm:"column:user_id;size:255"`
	Provider          *string    `gorm:"column:provider;size:255"`
	ProviderUserID    *string    `gorm:"column:provider_user_id;size:255"`
	AccessToken       *string    `gorm:"column:access_token;type:text"`
	PasswordChangedAt *time.Time `gorm:"column:password_changed_at"`
	RefreshToken      *string    `gorm:"column:refresh_token;type:text"`
	TokenVersion      *int       `gorm:"column:token_version"`
	CreatedBy         *string    `gorm:"column:created_by;size:50"`
	UpdatedBy         *string    `gorm:"column:updated_by;size:50"`
	CreatedAt         *time.Time `gorm:"column:created_at"`
	UpdatedAt         *time.Time `gorm:"column:updated_at"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
}

func (UserModel) TableName() string { return "users" }

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).Raw(selectUserByIDQuery, id).Scan(&m).Error; err != nil {
		return nil, err
	}
	return modelToUser(&m), nil
}

func (r *UserRepository) GetByUserName(ctx context.Context, userName string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Raw(selectUserByUserNameQuery, userName).Scan(&m).Error
	return modelToUser(&m), err
}

func (r *UserRepository) GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*domain.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).
		Raw(selectUserByProviderAndProviderUserIDQuery, provider, providerUserID).
		Scan(&m).Error; err != nil {
		return nil, err
	}
	return modelToUser(&m), nil
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	m := userToModel(u)
	return r.db.WithContext(ctx).Exec(
		userdb.InsertUserQuery,
		m.ID,
		m.UserName,
		m.FullName,
		m.Phone,
		m.PasswordHash,
		m.Salt,
		m.Status,
		m.IsVerified,
		m.UserID,
		m.Provider,
		m.ProviderUserID,
		m.AccessToken,
		m.PasswordChangedAt,
		m.RefreshToken,
		m.TokenVersion,
		m.CreatedBy,
		m.UpdatedBy,
		m.CreatedAt,
		m.UpdatedAt,
		m.ExpiresAt,
	).Error
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	m := userToModel(u)
	return r.db.WithContext(ctx).Exec(
		userdb.UpdateUserQuery,
		m.FullName,
		m.UserName,
		m.AccessToken,
		m.RefreshToken,
		m.UpdatedAt,
		m.ID,
	).Error
}

func userToModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:                u.ID,
		UserName:          u.UserName,
		FullName:          u.FullName,
		Phone:             u.Phone,
		PasswordHash:      u.PasswordHash,
		Salt:              u.Salt,
		Status:            u.Status,
		IsVerified:        u.IsVerified,
		UserID:            u.UserID,
		Provider:          u.Provider,
		ProviderUserID:    u.ProviderUserID,
		AccessToken:       u.AccessToken,
		PasswordChangedAt: u.PasswordChangedAt,
		RefreshToken:      u.RefreshToken,
		TokenVersion:      u.TokenVersion,
		CreatedBy:         u.CreatedBy,
		UpdatedBy:         u.UpdatedBy,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		ExpiresAt:         u.ExpiresAt,
	}
}

func modelToUser(m *UserModel) *domain.User {
	return &domain.User{
		ID:                m.ID,
		UserName:          m.UserName,
		FullName:          m.FullName,
		Phone:             m.Phone,
		PasswordHash:      m.PasswordHash,
		Salt:              m.Salt,
		Status:            m.Status,
		IsVerified:        m.IsVerified,
		UserID:            m.UserID,
		Provider:          m.Provider,
		ProviderUserID:    m.ProviderUserID,
		AccessToken:       m.AccessToken,
		PasswordChangedAt: m.PasswordChangedAt,
		RefreshToken:      m.RefreshToken,
		TokenVersion:      m.TokenVersion,
		CreatedBy:         m.CreatedBy,
		UpdatedBy:         m.UpdatedBy,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		ExpiresAt:         m.ExpiresAt,
	}
}
