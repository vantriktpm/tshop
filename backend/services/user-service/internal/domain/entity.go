package domain

import "time"

type User struct {
	ID                string     // uuid NOT NULL
	UserName          *string    // character(255)
	FullName          *string    // character(255)
	Phone             *string    // character(255)
	PasswordHash      *string    // text
	Salt              *string    // character(255)
	Status            *int16     // smallint
	IsVerified        *bool      // boolean
	UserID            *string    // character(255) - external user id
	Provider          *string    // character(255)
	ProviderUserID    *string    // character(255)
	AccessToken       *string    // text
	PasswordChangedAt *time.Time // timestamp with time zone
	RefreshToken      *string    // text
	TokenVersion      *int       // integer, used to revoke all tokens
	CreatedBy         *string    // character(50)
	UpdatedBy         *string    // character(50)
	CreatedAt         *time.Time // timestamp with time zone
	UpdatedAt         *time.Time // timestamp with time zone
	ExpiresAt         *time.Time // timestamp with time zone
}
