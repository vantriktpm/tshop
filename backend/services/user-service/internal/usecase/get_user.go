package usecase

import (
	"context"
	"github.com/tshop/backend/services/user-service/internal/domain"
)

type GetUser struct {
	repo domain.UserRepository
}

func NewGetUser(repo domain.UserRepository) *GetUser {
	return &GetUser{repo: repo}
}

func (u *GetUser) Execute(ctx context.Context, userID string) (*domain.User, error) {
	return u.repo.GetByID(ctx, userID)
}
