package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/users"
	"google.golang.org/grpc"
)

var _ mainflux.UsersServiceClient = (*usersServiceMock)(nil)

type usersServiceMock struct {
	users map[string]string
}

// NewUsersService creates mock of users service.
func NewUsersService(users map[string]string) mainflux.UsersServiceClient {
	return &usersServiceMock{users}
}

func (svc usersServiceMock) Identify(ctx context.Context, in *mainflux.Token, opts ...grpc.CallOption) (*mainflux.Identity, error) {
	if id, ok := svc.users[in.Value]; ok {
		return &mainflux.Identity{id}, nil
	}
	return nil, users.ErrUnauthorizedAccess
}
