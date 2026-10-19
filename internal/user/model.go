package user

import "github.com/golang-jwt/jwt/v4"

type Instance struct {
	ID       int
	Email    string
	Username string
	Password string
}

type Claims struct {
	jwt.RegisteredClaims

	ID       int
	Email    string
	Username string
	Exp      int64
}
