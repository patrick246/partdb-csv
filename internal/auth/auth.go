package auth

import "context"

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) error
}
