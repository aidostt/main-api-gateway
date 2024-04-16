package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reservista.kz/internal/repository"
	"reservista.kz/pkg/hash"
	auth "reservista.kz/pkg/manager"
	"time"
)

type UserSignUpInput struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

type UserSignInInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Session interface {
	RefreshTokens(context.Context, string) (TokenPair, error)
	CreateSession(context.Context, primitive.ObjectID) (TokenPair, error)
	GetToken(context.Context, string) (string, error)
}

type Users interface {
	SignUp(context.Context, UserSignUpInput) (TokenPair, error)
	SignIn(context.Context, UserSignInInput) (TokenPair, error)
	CreateSession(context.Context, primitive.ObjectID) (TokenPair, error)
}

type Services struct {
	Users   Users
	Session Session
}

type Dependencies struct {
	Repos           *repository.Models
	Hasher          hash.PasswordHasher
	TokenManager    auth.TokenManager
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Environment     string
	Domain          string
}

func NewServices(deps Dependencies) *Services {
	sessionService := NewSessionsService(deps.Repos.Sessions, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain)
	usersService := NewUsersService(deps.Repos.Users, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain, sessionService)
	return &Services{
		Users:   usersService,
		Session: sessionService,
	}
}
