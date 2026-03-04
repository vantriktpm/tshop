package domain

import "context"

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUserName(ctx context.Context, userName string) (*User, error)
	GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}
