package storage // код работы с бд

import "errors"

var (
	ErrUserExists   = errors.New("User already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)
