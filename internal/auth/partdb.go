package auth

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var ErrUsernamePasswordMismatch = errors.New("invalid username or password")

type PartDBAuthenticator struct {
	db *sql.DB
}

func NewPartDBAuthenticator(db *sql.DB) *PartDBAuthenticator {
	return &PartDBAuthenticator{
		db: db,
	}
}

func (a *PartDBAuthenticator) Authenticate(ctx context.Context, username, password string) error {
	var passwordHash string
	err := a.db.QueryRowContext(ctx, `SELECT password FROM users WHERE name = ?`, username).
		Scan(&passwordHash)
	userNotFound := errors.Is(err, sql.ErrNoRows)
	if !userNotFound && err != nil {
		return err
	}

	// Use a dummy hash in case the user couldn't be found. This will prevent timing attacks to enumerate users
	if userNotFound && passwordHash == "" {
		passwordHash = "$2a$10$PuzrdtSJWGVOCwpMA5bIReejK/nfO1Bj8mwxJhJZdydRYvqnN87Oy"
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if userNotFound {
		return ErrUsernamePasswordMismatch
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrUsernamePasswordMismatch
	}
	if err != nil {
		return err
	}
	return nil
}
