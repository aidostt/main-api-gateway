package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reservista.kz/internal/domain"
	"reservista.kz/internal/repository"
	"reservista.kz/pkg/hash"
	auth "reservista.kz/pkg/manager"
	"time"
)

type SessionService struct {
	repo            repository.Sessions
	hasher          hash.PasswordHasher
	tokenManager    auth.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	domain string
}

func NewSessionsService(repo repository.Sessions, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL time.Duration, domain string) *SessionService {
	return &SessionService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
	}
}

func (s *SessionService) RefreshTokens(ctx context.Context, refreshToken string) (res TokenPair, err error) {
	session, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return TokenPair{}, err
	}

	return s.CreateSession(ctx, session.UserID)
}

func (s *SessionService) CreateSession(ctx context.Context, userId primitive.ObjectID) (res TokenPair, err error) {
	res.AccessToken, err = s.tokenManager.NewAccessToken(userId.Hex(), s.accessTokenTTL)
	if err != nil {
		return TokenPair{}, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	session := domain.Session{
		UserID:       userId,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(s.refreshTokenTTL),
	}

	err = s.repo.SetSession(ctx, session)

	return
}

func (s *SessionService) GetToken(ctx context.Context, RT string) (string, error) {
	session, err := s.repo.GetByRefreshToken(ctx, RT)
	if err != nil {
		return "", err

	}
	return session.RefreshToken, nil
}
