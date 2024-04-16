package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reservista.kz/internal/domain"
)

type Users interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	//TODO: delete user
}
type Sessions interface {
	SetSession(ctx context.Context, session domain.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) (domain.User, error)
	//TODO: delete session
}

type Models struct {
	Users    Users
	Sessions Sessions
}

func NewModels(db *mongo.Database) *Models {
	return &Models{
		Users:    NewUsersRepo(db),
		Sessions: NewSessionRepo(db),
	}
}
