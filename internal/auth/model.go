package auth

import "time"

type TokenWithUser struct {
	ID        int
	UserID    int
	Email     string
	Username  string
	TokenHash string
	ExpiresAt time.Time
}
