package data

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Users interface {
		Insert(*User) error
		GetByNickname(string) (*User, error)
		Update(*User) error
		Delete(*User) error
	}
	//Tokens interface {
	//	Insert(*Token) error
	//	DeleteAllForUser(string, pgtype.UUID) error
	//}
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Users: UserModel{DB: db},
		//Tokens: TokenModel{DB: db},
	}
}
