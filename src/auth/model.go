package auth

import "github.com/dgrijalva/jwt-go"

// Claims represents the JWT claims for authentication
type Claims struct {
	jwt.StandardClaims
}

// Token contains the JWT token for authentication
type Token struct {
	Token string `json:"token"`
}

// User represents a registered user
type User struct {
	ID       string `bson:"_id"`
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

// RegisterRequest represents the request body for the /register endpoint
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for the /login endpoint
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
