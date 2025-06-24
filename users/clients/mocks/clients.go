// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/clients/postgres"
	"github.com/stretchr/testify/mock"
)

const WrongID = "wrongID"

var _ postgres.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

func (m *Repository) ChangeStatus(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	if client.Status != mfclients.EnabledStatus && client.Status != mfclients.DisabledStatus {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) Members(ctx context.Context, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	ret := m.Called(ctx, groupID, pm)
	if groupID == WrongID {
		return mfclients.MembersPage{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.MembersPage), ret.Error(1)
}

func (m *Repository) RetrieveAll(ctx context.Context, pm mfclients.Page) (mfclients.ClientsPage, error) {
	ret := m.Called(ctx, pm)

	return ret.Get(0).(mfclients.ClientsPage), ret.Error(1)
}

func (m *Repository) RetrieveByID(ctx context.Context, id string) (mfclients.Client, error) {
	ret := m.Called(ctx, id)

	if id == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) RetrieveByIdentity(ctx context.Context, identity string) (mfclients.Client, error) {
	ret := m.Called(ctx, identity)

	if identity == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) Save(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)
	if client.Owner == WrongID {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}
	if client.Credentials.Secret == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return client, ret.Error(1)
}

func (m *Repository) Update(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}
	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) UpdateIdentity(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Identity == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) UpdateSecret(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Secret == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) UpdateTags(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *Repository) UpdateOwner(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (*Repository) RetrieveBySecret(ctx context.Context, key string) (mfclients.Client, error) {
	return mfclients.Client{}, nil
}
