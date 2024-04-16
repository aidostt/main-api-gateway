package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reservista.kz/internal/domain"
	"reservista.kz/internal/repository"
	"reservista.kz/pkg/hash"
	auth "reservista.kz/pkg/manager"
	"time"
)

type UsersService struct {
	repo            repository.Users
	hasher          hash.PasswordHasher
	tokenManager    auth.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	SessionService  *SessionService
	domain          string
}

func NewUsersService(repo repository.Users, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL time.Duration, domain string, sessionService *SessionService) *UsersService {
	return &UsersService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
		SessionService:  sessionService,
	}
}
func (s *UsersService) CreateSession(ctx context.Context, userId primitive.ObjectID) (TokenPair, error) {
	return s.SessionService.CreateSession(ctx, userId)
}
func (s *UsersService) SignUp(ctx context.Context, input UserSignUpInput) (TokenPair, error) {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return TokenPair{}, err
	}

	user := &domain.User{
		Email:    input.Email,
		Password: passwordHash,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return TokenPair{}, err
		}
		return TokenPair{}, err
	}
	return s.CreateSession(ctx, user.ID)
}

func (s *UsersService) SignIn(ctx context.Context, input UserSignInInput) (TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return TokenPair{}, err
		}

		return TokenPair{}, err
	}

	return s.CreateSession(ctx, user.ID)
}
