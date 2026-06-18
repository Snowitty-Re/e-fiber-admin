package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/Snowitty/e-fiber-admin/internal/ent"
	"github.com/Snowitty/e-fiber-admin/internal/ent/adminuser"
	"github.com/Snowitty/e-fiber-admin/internal/pkg/auth"
	pkgerr "github.com/Snowitty/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient    *ent.Client
	redisClient  *redis.Client
	tokenManager *auth.TokenManager
}

func NewService(entClient *ent.Client, redisClient *redis.Client, tm *auth.TokenManager) *Service {
	return &Service{entClient: entClient, redisClient: redisClient, tokenManager: tm}
}

type AdminIdentity struct {
	ID        int64
	Email     string
	FirstName string
	LastName  string
	Roles     []string
	Perms     []string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessJTI    string
	RefreshJTI   string
	ExpiresIn    int64
}

func (s *Service) Login(ctx context.Context, email, password string) (*AdminIdentity, *TokenPair, error) {
	admin, err := s.entClient.AdminUser.Query().
		Where(adminuser.EmailEQ(email)).
		WithRoles(func(q *ent.RoleQuery) { q.WithPermissions() }).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, pkgerr.ErrUnauthorized.WithCause(err)
		}
		return nil, nil, fmt.Errorf("query admin: %w", err)
	}
	if admin.Status != "active" {
		return nil, nil, pkgerr.New("AUTH_ACCOUNT_DISABLED", 403, "account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, nil, pkgerr.ErrUnauthorized.WithCause(err)
	}

	identity := s.toIdentity(admin)
	pair, err := s.issueTokens(identity)
	if err != nil {
		return nil, nil, err
	}

	_, _ = s.entClient.AdminUser.UpdateOne(admin).SetLastLoginAt(time.Now()).Save(ctx)
	return &identity, pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil, pkgerr.ErrTokenInvalid.WithCause(err)
	}
	if claims.TokenType != auth.AdminRefresh {
		return nil, pkgerr.ErrTokenInvalid
	}
	exists, err := s.redisClient.Exists(ctx, refreshKey(claims.ID)).Result()
	if err != nil {
		return nil, fmt.Errorf("check refresh jti: %w", err)
	}
	if exists == 0 {
		return nil, pkgerr.ErrTokenInvalid
	}
	s.redisClient.Del(ctx, refreshKey(claims.ID))

	admin, err := s.entClient.AdminUser.Query().
		Where(adminuser.IDEQ(int(claims.AdminID))).
		WithRoles(func(q *ent.RoleQuery) { q.WithPermissions() }).
		Only(ctx)
	if err != nil {
		return nil, pkgerr.ErrTokenInvalid.WithCause(err)
	}
	if admin.Status != "active" {
		return nil, pkgerr.New("AUTH_ACCOUNT_DISABLED", 403, "account is disabled")
	}
	identity := s.toIdentity(admin)
	return s.issueTokens(identity)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil
	}
	if claims.TokenType == auth.AdminRefresh {
		s.redisClient.Del(ctx, refreshKey(claims.ID))
	}
	return nil
}

func (s *Service) Me(ctx context.Context, adminID int64) (*AdminIdentity, error) {
	admin, err := s.entClient.AdminUser.Query().
		Where(adminuser.IDEQ(int(adminID))).
		WithRoles(func(q *ent.RoleQuery) { q.WithPermissions() }).
		Only(ctx)
	if err != nil {
		return nil, pkgerr.ErrNotFound.WithCause(err)
	}
	identity := s.toIdentity(admin)
	return &identity, nil
}

func (s *Service) ParseAccess(tokenStr string) (*auth.Claims, error) {
	claims, err := s.tokenManager.ParseAccess(tokenStr)
	if err != nil {
		if err == auth.ErrExpired {
			return nil, pkgerr.ErrTokenExpired
		}
		return nil, pkgerr.ErrTokenInvalid
	}
	if claims.TokenType != auth.AdminAccess {
		return nil, pkgerr.ErrTokenInvalid
	}
	return claims, nil
}

func (s *Service) issueTokens(id AdminIdentity) (*TokenPair, error) {
	access, accessJTI, err := s.tokenManager.IssueAccess(auth.AdminAccess, id.ID, 0, id.Roles, id.Perms)
	if err != nil {
		return nil, err
	}
	refresh, refreshJTI, err := s.tokenManager.IssueRefresh(auth.AdminRefresh, id.ID, 0)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	s.redisClient.Set(ctx, refreshKey(refreshJTI), id.ID, s.tokenManager.RefreshTTL())
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		AccessJTI:    accessJTI,
		RefreshJTI:   refreshJTI,
		ExpiresIn:    int64(s.tokenManager.AccessTTL().Seconds()),
	}, nil
}

func (s *Service) toIdentity(admin *ent.AdminUser) AdminIdentity {
	var roles, perms []string
	for _, r := range admin.Edges.Roles {
		roles = append(roles, r.Slug)
		for _, p := range r.Edges.Permissions {
			perms = append(perms, p.Resource+":"+p.Action)
		}
	}
	return AdminIdentity{
		ID:        int64(admin.ID),
		Email:     admin.Email,
		FirstName: admin.FirstName,
		LastName:  admin.LastName,
		Roles:     roles,
		Perms:     perms,
	}
}

func refreshKey(jti string) string {
	return "admin:refresh:" + jti
}
