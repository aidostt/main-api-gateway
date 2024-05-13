package domain

import "errors"

var (
	ErrUserNotFound         = errors.New("user doesn't exists")
	ErrUserAlreadyExists    = errors.New("user with such email already exists")
	ErrTokenExpired         = errors.New("token is expired")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrTokenInvalidElements = errors.New("token has xxx elements")
)
