package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AdminAccess        TokenType = "admin"
	AdminRefresh       TokenType = "admin_refresh"
	CustomerAccess     TokenType = "customer"
	CustomerRefresh    TokenType = "customer_refresh"
)

type Claims struct {
	jwt.RegisteredClaims
	TokenType TokenType `json:"typ"`
	AdminID   int64     `json:"admin_id,omitempty"`
	CustomerID int64    `json:"customer_id,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	Perms     []string  `json:"perms,omitempty"`
}

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m *TokenManager) IssueAccess(typ TokenType, adminID, customerID int64, roles, perms []string) (string, string, error) {
	now := time.Now()
	jti := uuid.NewString()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   subject(typ, adminID, customerID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
		TokenType:  typ,
		AdminID:    adminID,
		CustomerID: customerID,
		Roles:      roles,
		Perms:      perms,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, jti, nil
}

func (m *TokenManager) IssueRefresh(typ TokenType, adminID, customerID int64) (string, string, error) {
	now := time.Now()
	jti := uuid.NewString()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   subject(typ, adminID, customerID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
		TokenType:  typ,
		AdminID:    adminID,
		CustomerID: customerID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.refreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, jti, nil
}

func (m *TokenManager) ParseAccess(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.accessSecret)
}

func (m *TokenManager) ParseRefresh(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.refreshSecret)
}

func (m *TokenManager) parse(tokenStr string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, ErrInvalid
	}
	return claims, nil
}

func (m *TokenManager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }

func subject(typ TokenType, adminID, customerID int64) string {
	switch typ {
	case AdminAccess, AdminRefresh:
		return fmt.Sprintf("admin:%d", adminID)
	case CustomerAccess, CustomerRefresh:
		return fmt.Sprintf("customer:%d", customerID)
	}
	return ""
}

var (
	ErrExpired = errors.New("token expired")
	ErrInvalid = errors.New("token invalid")
)
