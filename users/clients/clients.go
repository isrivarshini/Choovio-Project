// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"

	"github.com/mainflux/mainflux/pkg/clients"
)

// ClientService specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type ClientService interface {
	// RegisterClient creates new client. In case of the failed registration, a
	// non-nil error value is returned.
	RegisterClient(ctx context.Context, token string, client clients.Client) (clients.Client, error)

	// ViewClient retrieves client info for a given client ID and an authorized token.
	ViewClient(ctx context.Context, token, id string) (clients.Client, error)

	// ViewProfile retrieves client info for a given token.
	ViewProfile(ctx context.Context, token string) (clients.Client, error)

	// ListClients retrieves clients list for a valid auth token.
	ListClients(ctx context.Context, token string, pm clients.Page) (clients.ClientsPage, error)

	// ListMembers retrieves everything that is assigned to a group identified by groupID.
	ListMembers(ctx context.Context, token, groupID string, pm clients.Page) (clients.MembersPage, error)

	// UpdateClient updates the client's name and metadata.
	UpdateClient(ctx context.Context, token string, client clients.Client) (clients.Client, error)

	// UpdateClientTags updates the client's tags.
	UpdateClientTags(ctx context.Context, token string, client clients.Client) (clients.Client, error)

	// UpdateClientIdentity updates the client's identity.
	UpdateClientIdentity(ctx context.Context, token, id, identity string) (clients.Client, error)

	// GenerateResetToken email where mail will be sent.
	// host is used for generating reset link.
	GenerateResetToken(ctx context.Context, email, host string) error

	// UpdateClientSecret updates the client's secret.
	UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (clients.Client, error)

	// ResetSecret change users secret in reset flow.
	// token can be authentication token or secret reset token.
	ResetSecret(ctx context.Context, resetToken, secret string) error

	// SendPasswordReset sends reset password link to email.
	SendPasswordReset(ctx context.Context, host, email, user, token string) error

	// UpdateClientOwner updates the client's owner.
	UpdateClientOwner(ctx context.Context, token string, client clients.Client) (clients.Client, error)

	// EnableClient logically enableds the client identified with the provided ID.
	EnableClient(ctx context.Context, token, id string) (clients.Client, error)

	// DisableClient logically disables the client identified with the provided ID.
	DisableClient(ctx context.Context, token, id string) (clients.Client, error)

	// Identify returns the client id from the given token.
	Identify(ctx context.Context, tkn string) (string, error)
}
