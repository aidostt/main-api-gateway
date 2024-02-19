package data

import (
	"context"
	"errors"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"reservista/internal/validator"
	"strings"
	"time"
	"unicode"
)

var (
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateToken    = errors.New("duplicate token")
	ErrDuplicateNickname = errors.New("duplicate nickname")
	ErrNoRecord          = errors.New("no record found")
	ErrEditConflict      = errors.New("edit conflict")
)

type User struct {
	ID        pgtype.UUID `form:"id"`
	Name      string      `form:"name"`
	Nickname  string      `form:"nickname"`
	Email     string      `form:"email"`
	Password  password    `form:"-"`
	Activated bool        `form:"activated"`
	Version   int         `form:"-"`
	CreatedAt time.Time   `form:"created_at"`
}

type password struct {
	plainText *string
	hashed    []byte
}

type UserModel struct {
	DB *pgxpool.Pool
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Nickname != "", "nickname", "must be provided")
	ValidateEmail(v, user.Email)
	ValidatePlaintextPassword(v, user.Password.plainText)

	if user.Password.hashed == nil {
		panic("missing user's hashed password")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")

}

func ValidatePlaintextPassword(v *validator.Validator, plaintext *string) {
	v.Check(*plaintext != "", "password", "must be provided")
	v.Check(len(*plaintext) >= 8, "password", "must be more than 8 characters")
	v.Check(len(*plaintext) <= 72, "password", "must be less than 72 characters")
	//v.Check(validator.Matches(*plaintext, validator.PasswordRX), "password", "password must contain at least: "+
	//	"1 special sign, 1 uppercase letter, 1 lowercase letter, 1 number")
	var (
		up    bool
		low   bool
		digit bool
		spec  bool
	)

	for _, char := range *plaintext {
		switch {
		case unicode.IsUpper(char):
			up = true
		case unicode.IsLower(char):
			low = true
		case unicode.IsDigit(char):
			digit = true
		case strings.IndexRune("!#$%&'*+\\/=?^_`{|}~-", char) != -1:
			spec = true
		}
	}

	switch {
	case !up:
		v.AddError("password", "must contain at least 1 uppercase letter")
	case !low:
		v.AddError("password", "must contain at least 1 lowercase letter")
	case !digit:
		v.AddError("password", "must contain at least 1 digit")
	case !spec:
		v.AddError("password", "must contain at least 1 special character")
	}
}

func (m UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (name, nickname, email, password_hash, activated) VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, version;
	`
	args := []any{user.Name, user.Nickname, user.Email, user.Password.hashed, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRow(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if strings.Contains(pgErr.ConstraintName, "nickname") {
				return ErrDuplicateNickname
			} else if strings.Contains(pgErr.ConstraintName, "email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m UserModel) GetByNickname(nickname string) (*User, error) {
	query := `
	SELECT * FROM users 
	         WHERE nickname = $1;
`
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRow(ctx, query, nickname).Scan(
		&user.ID,
		&user.Name,
		&user.Nickname,
		&user.Email,
		&user.Password.hashed,
		&user.Activated,
		&user.Version,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1, nickname = $2, email = $3, password_hash = $4, activated = $5, version = version + 1
	WHERE id = $6 AND version = $7
	RETURNING version
`
	args := []any{
		user.Name,
		user.Nickname,
		user.Email,
		user.Password.hashed,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRow(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case strings.Contains(pgErr.ConstraintName, "nickname"):
			return ErrDuplicateNickname
		case strings.Contains(pgErr.ConstraintName, "email"):
			return ErrDuplicateEmail
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
func (m UserModel) Delete(user *User) error {
	query := `DELETE FROM users WHERE id = $1 AND version = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.Exec(ctx, query, user.ID, user.Version)
	return err
}

func (p *password) Set(plaintTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintTextPassword), 12)
	if err != nil {
		return err
	}
	p.plainText = &plaintTextPassword
	p.hashed = hash
	return nil
}
