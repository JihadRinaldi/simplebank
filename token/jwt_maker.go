package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const minSecretKeySize = 32

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         payload.ID,
		"username":   payload.Username,
		"issued_at":  payload.IssuedAt,
		"expired_at": payload.ExpiredAt,
	})
	tokenString, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (maker *JWTMaker) VerifyToken(tokenString string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, keyFunc)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	idStr, ok := claims["id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	issuedAtStr, ok := claims["issued_at"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	issuedAt, err := time.Parse(time.RFC3339, issuedAtStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	expiredAtStr, ok := claims["expired_at"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	expiredAt, err := time.Parse(time.RFC3339, expiredAtStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if time.Now().After(expiredAt) {
		return nil, ErrExpiredToken
	}

	payload := &Payload{
		ID:        id,
		Username:  username,
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}

	return payload, nil
}
