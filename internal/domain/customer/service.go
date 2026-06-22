package customer

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/customer"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/customeraddress"
	"github.com/Snowitty-Re/e-fiber-admin/internal/pkg/auth"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient    *ent.Client
	redisClient  *redis.Client
	tokenManager *auth.TokenManager
}

func NewService(entClient *ent.Client, redisClient *redis.Client, tm *auth.TokenManager) *Service {
	return &Service{
		entClient:    entClient,
		redisClient:  redisClient,
		tokenManager: tm,
	}
}

func refreshKey(jti string) string {
	return "customer:refresh:" + jti
}

type CustomerIdentity struct {
	ID              int64
	Email           string
	FirstName       string
	LastName        string
	DefaultCurrency string
	DefaultLocale   string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RegisterInput struct {
	Email           string
	Password        string
	FirstName       string
	LastName        string
	DefaultCurrency string
	DefaultLocale   string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*CustomerIdentity, *TokenPair, error) {
	exists, err := s.entClient.Customer.Query().Where(customer.EmailEQ(in.Email)).Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("query customer: %w", err)
	}
	if exists > 0 {
		return nil, nil, pkgerr.New("CUSTOMER_EMAIL_EXISTS", 409, "email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	if in.DefaultCurrency == "" {
		in.DefaultCurrency = "USD"
	}
	if in.DefaultLocale == "" {
		in.DefaultLocale = "en"
	}

	c, err := s.entClient.Customer.Create().
		SetEmail(in.Email).
		SetPasswordHash(string(hash)).
		SetFirstName(in.FirstName).
		SetLastName(in.LastName).
		SetDefaultCurrency(in.DefaultCurrency).
		SetDefaultLocale(in.DefaultLocale).
		SetStatus(customer.StatusActive).
		Save(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("create customer: %w", err)
	}

	identity := CustomerIdentity{
		ID: int64(c.ID), Email: c.Email, FirstName: c.FirstName,
		LastName: c.LastName, DefaultCurrency: c.DefaultCurrency,
		DefaultLocale: c.DefaultLocale,
	}
	pair, err := s.issueTokens(identity)
	if err != nil {
		return nil, nil, err
	}
	return &identity, pair, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*CustomerIdentity, *TokenPair, error) {
	c, err := s.entClient.Customer.Query().Where(customer.EmailEQ(email)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, pkgerr.New("AUTH_INVALID_CREDENTIALS", 401, "invalid email or password")
		}
		return nil, nil, fmt.Errorf("query customer: %w", err)
	}
	if c.Status != customer.StatusActive {
		return nil, nil, pkgerr.New("AUTH_ACCOUNT_DISABLED", 403, "account is disabled")
	}
	if c.PasswordHash == "" {
		return nil, nil, pkgerr.New("AUTH_INVALID_CREDENTIALS", 401, "no password set, please reset")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.PasswordHash), []byte(password)); err != nil {
		return nil, nil, pkgerr.New("AUTH_INVALID_CREDENTIALS", 401, "invalid email or password")
	}

	identity := CustomerIdentity{
		ID: int64(c.ID), Email: c.Email, FirstName: c.FirstName,
		LastName: c.LastName, DefaultCurrency: c.DefaultCurrency,
		DefaultLocale: c.DefaultLocale,
	}
	pair, err := s.issueTokens(identity)
	if err != nil {
		return nil, nil, err
	}
	return &identity, pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil, pkgerr.ErrTokenInvalid.WithCause(err)
	}
	if claims.TokenType != auth.CustomerRefresh {
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
	c, err := s.entClient.Customer.Get(ctx, int(claims.CustomerID))
	if err != nil {
		return nil, pkgerr.ErrTokenInvalid.WithCause(err)
	}
	identity := CustomerIdentity{
		ID: int64(c.ID), Email: c.Email, FirstName: c.FirstName,
		LastName: c.LastName, DefaultCurrency: c.DefaultCurrency,
		DefaultLocale: c.DefaultLocale,
	}
	return s.issueTokens(identity)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil
	}
	if claims.TokenType == auth.CustomerRefresh {
		s.redisClient.Del(ctx, refreshKey(claims.ID))
	}
	return nil
}

func (s *Service) Me(ctx context.Context, id int64) (*CustomerIdentity, error) {
	c, err := s.entClient.Customer.Get(ctx, int(id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query customer: %w", err)
	}
	return &CustomerIdentity{
		ID: int64(c.ID), Email: c.Email, FirstName: c.FirstName,
		LastName: c.LastName, DefaultCurrency: c.DefaultCurrency,
		DefaultLocale: c.DefaultLocale,
	}, nil
}

func (s *Service) ParseAccess(tokenStr string) (*auth.Claims, error) {
	claims, err := s.tokenManager.ParseAccess(tokenStr)
	if err != nil {
		if err == auth.ErrExpired {
			return nil, pkgerr.ErrTokenExpired
		}
		return nil, pkgerr.ErrTokenInvalid
	}
	if claims.TokenType != auth.CustomerAccess {
		return nil, pkgerr.ErrTokenInvalid
	}
	return claims, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int, search string) ([]*ent.Customer, int, error) {
	q := s.entClient.Customer.Query()
	if search != "" {
		q = q.Where(customer.EmailContains(search))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, err := q.Order(ent.Desc(customer.FieldID)).Limit(pageSize).Offset((page - 1) * pageSize).All(ctx)
	return items, total, err
}

func (s *Service) Get(ctx context.Context, id int) (*ent.Customer, error) {
	c, err := s.entClient.Customer.Query().
		Where(customer.IDEQ(id)).
		WithAddresses().
		WithGroups().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query customer: %w", err)
	}
	return c, nil
}

func (s *Service) Disable(ctx context.Context, id int) error {
	return s.entClient.Customer.UpdateOneID(id).SetStatus(customer.StatusDisabled).Exec(ctx)
}

func (s *Service) ListAddresses(ctx context.Context, customerID int) ([]*ent.CustomerAddress, error) {
	return s.entClient.CustomerAddress.Query().
		Where(customeraddress.CustomerIDEQ(customerID)).
		All(ctx)
}

func (s *Service) AddAddress(ctx context.Context, customerID int, in ent.CustomerAddress) (*ent.CustomerAddress, error) {
	return s.entClient.CustomerAddress.Create().
		SetCustomerID(customerID).
		SetFirstName(in.FirstName).
		SetLastName(in.LastName).
		SetCompany(in.Company).
		SetAddress1(in.Address1).
		SetAddress2(in.Address2).
		SetCity(in.City).
		SetProvince(in.Province).
		SetPostalCode(in.PostalCode).
		SetCountryCode(in.CountryCode).
		SetPhone(in.Phone).
		SetIsDefaultShipping(in.IsDefaultShipping).
		SetIsDefaultBilling(in.IsDefaultBilling).
		Save(ctx)
}

func (s *Service) ListGroups(ctx context.Context) ([]*ent.CustomerGroup, error) {
	return s.entClient.CustomerGroup.Query().All(ctx)
}

func (s *Service) CreateGroup(ctx context.Context, slug, name string) (*ent.CustomerGroup, error) {
	return s.entClient.CustomerGroup.Create().SetSlug(slug).SetName(name).Save(ctx)
}

func (s *Service) AddToGroup(ctx context.Context, customerID, groupID int) error {
	_, err := s.entClient.Customer.UpdateOneID(customerID).AddGroupIDs(groupID).Save(ctx)
	return err
}

func (s *Service) issueTokens(id CustomerIdentity) (*TokenPair, error) {
	access, _, err := s.tokenManager.IssueAccess(auth.CustomerAccess, 0, id.ID, nil, nil)
	if err != nil {
		return nil, err
	}
	refresh, refreshJTI, err := s.tokenManager.IssueRefresh(auth.CustomerRefresh, 0, id.ID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	s.redisClient.Set(ctx, refreshKey(refreshJTI), id.ID, s.tokenManager.RefreshTTL())
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.tokenManager.AccessTTL().Seconds()),
	}, nil
}