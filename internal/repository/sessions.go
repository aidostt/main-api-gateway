package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reservista.kz/internal/domain"
	"time"
)

type SessionRepo struct {
	db *mongo.Collection
}

func NewSessionRepo(db *mongo.Database) *SessionRepo {
	return &SessionRepo{
		db: db.Collection(sessionCollection),
	}
}

func (r *SessionRepo) SetSession(ctx context.Context, session domain.Session) error {
	filter := bson.M{"_id": session.UserID}
	update := bson.M{
		"$set": bson.M{
			"lastVisitAt":  time.Now(),
			"userID":       session.UserID,
			"refreshToken": session.RefreshToken,
			"expiresAt":    session.ExpiredAt,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedDoc bson.M
	err := r.db.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *SessionRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	session := &domain.Session{}
	if err := r.db.FindOne(ctx, bson.M{
		"refreshToken": refreshToken,
	}).Decode(&session); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &domain.Session{}, domain.ErrUserNotFound
		}

		return &domain.Session{}, err
	}

	return session, nil
}
func (r *SessionRepo) GetByUserID(ctx context.Context, userID primitive.ObjectID) (domain.User, error) {
	var user domain.User
	if err := r.db.FindOne(ctx, bson.M{
		"userID": userID,
	}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}
