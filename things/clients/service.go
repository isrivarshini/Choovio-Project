// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	mfgroups "github.com/mainflux/mainflux/pkg/groups"
	tpolicies "github.com/mainflux/mainflux/things/policies"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

const (
	MyKey             = "mine"
	thingsObjectKey   = "things"
	addRelationKey    = "g_add"
	updateRelationKey = "c_update"
	listRelationKey   = "c_list"
	deleteRelationKey = "c_delete"
	groupEntityType   = "group"
	clientEntityType  = "client"
)

var AdminRelationKey = []string{updateRelationKey, listRelationKey, deleteRelationKey}

type service struct {
	uauth       upolicies.AuthServiceClient
	policies    tpolicies.Service
	clients     mfclients.Repository
	clientCache Cache
	idProvider  mainflux.IDProvider
	grepo       mfgroups.Repository
}

// NewService returns a new Clients service implementation.
func NewService(uauth upolicies.AuthServiceClient, policies tpolicies.Service, c mfclients.Repository, grepo mfgroups.Repository, tcache Cache, idp mainflux.IDProvider) Service {
	return service{
		uauth:       uauth,
		policies:    policies,
		clients:     c,
		grepo:       grepo,
		clientCache: tcache,
		idProvider:  idp,
	}
}

func (svc service) CreateThings(ctx context.Context, token string, clis ...mfclients.Client) ([]mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return []mfclients.Client{}, err
	}
	var clients []mfclients.Client
	for _, cli := range clis {
		if cli.ID == "" {
			clientID, err := svc.idProvider.ID()
			if err != nil {
				return []mfclients.Client{}, err
			}
			cli.ID = clientID
		}
		if cli.Credentials.Secret == "" {
			key, err := svc.idProvider.ID()
			if err != nil {
				return []mfclients.Client{}, err
			}
			cli.Credentials.Secret = key
		}
		if cli.Owner == "" {
			cli.Owner = userID
		}
		if cli.Status != mfclients.DisabledStatus && cli.Status != mfclients.EnabledStatus {
			return []mfclients.Client{}, apiutil.ErrInvalidStatus
		}
		cli.CreatedAt = time.Now()
		clients = append(clients, cli)
	}
	return svc.clients.Save(ctx, clients...)
}

func (svc service) ViewClient(ctx context.Context, token string, id string) (mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, userID, id, listRelationKey); err != nil {
		return mfclients.Client{}, errors.Wrap(errors.ErrNotFound, err)
	}
	return svc.clients.RetrieveByID(ctx, id)
}

func (svc service) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}

	// If the user is admin, fetch all things from database.
	if err := svc.checkAdmin(ctx, userID, thingsObjectKey, listRelationKey); err == nil {
		pm.Owner = ""
		pm.SharedBy = ""
		return svc.clients.RetrieveAll(ctx, pm)
	}

	// If the user is not admin, check 'sharedby' parameter from page metadata.
	// If user provides 'sharedby' key, fetch things from policies. Otherwise,
	// fetch things from the database based on thing's 'owner' field.
	if pm.SharedBy == MyKey {
		pm.SharedBy = userID
	}
	if pm.Owner == MyKey {
		pm.Owner = userID
	}
	pm.Action = "c_list"

	return svc.clients.RetrieveAll(ctx, pm)
}

func (svc service) UpdateClient(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, userID, cli.ID, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:        cli.ID,
		Name:      cli.Name,
		Metadata:  cli.Metadata,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.Update(ctx, client)
}

func (svc service) UpdateClientTags(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, userID, cli.ID, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.UpdateTags(ctx, client)
}

func (svc service) UpdateClientSecret(ctx context.Context, token, id, key string) (mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, userID, id, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID: id,
		Credentials: mfclients.Credentials{
			Secret: key,
		},
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.UpdateSecret(ctx, client)
}

func (svc service) UpdateClientOwner(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, userID, cli.ID, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:        cli.ID,
		Owner:     cli.Owner,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.UpdateOwner(ctx, client)
}

func (svc service) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	client := mfclients.Client{
		ID:        id,
		Status:    mfclients.EnabledStatus,
		UpdatedAt: time.Now(),
	}
	client, err := svc.changeClientStatus(ctx, token, client)
	if err != nil {
		return mfclients.Client{}, errors.Wrap(mfclients.ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	client := mfclients.Client{
		ID:        id,
		Status:    mfclients.DisabledStatus,
		UpdatedAt: time.Now(),
	}
	client, err := svc.changeClientStatus(ctx, token, client)
	if err != nil {
		return mfclients.Client{}, errors.Wrap(mfclients.ErrDisableClient, err)
	}

	if err := svc.clientCache.Remove(ctx, client.ID); err != nil {
		return client, err
	}

	return client, nil
}

func (svc service) changeClientStatus(ctx context.Context, token string, client mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, userID, client.ID, deleteRelationKey); err != nil {
		return mfclients.Client{}, err
	}
	dbClient, err := svc.clients.RetrieveByID(ctx, client.ID)
	if err != nil {
		return mfclients.Client{}, err
	}
	if dbClient.Status == client.Status {
		return mfclients.Client{}, mfclients.ErrStatusAlreadyAssigned
	}
	client.UpdatedBy = userID
	return svc.clients.ChangeStatus(ctx, client)
}

func (svc service) ListClientsByGroup(ctx context.Context, token, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	userID, err := svc.identify(ctx, token)
	if err != nil {
		return mfclients.MembersPage{}, err
	}
	// If the user is admin, fetch all things connected to the channel.
	if err := svc.checkAdmin(ctx, userID, thingsObjectKey, listRelationKey); err == nil {
		return svc.clients.Members(ctx, groupID, pm)
	}
	pm.Owner = userID

	return svc.clients.Members(ctx, groupID, pm)
}

func (svc service) Identify(ctx context.Context, key string) (string, error) {
	id, err := svc.clientCache.ID(ctx, key)
	if err == nil {
		return id, nil
	}
	client, err := svc.clients.RetrieveBySecret(ctx, key)
	if err != nil {
		return "", err
	}
	if err := svc.clientCache.Save(ctx, key, client.ID); err != nil {
		return "", err
	}
	return client.ID, nil
}

// ShareClient shares a thing with a user.
// We assume the user has already created the things anf group.
func (svc service) ShareClient(ctx context.Context, token, userID, groupID, thingID string, actions []string) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	areq := tpolicies.AccessRequest{Subject: id, Object: groupID, Action: addRelationKey, Entity: groupEntityType}
	if _, err := svc.policies.Authorize(ctx, areq); err != nil {
		return fmt.Errorf("cannot share things using group %s to user %s: %s", groupID, userID, err)
	}

	areq = tpolicies.AccessRequest{Subject: id, Object: thingID, Action: addRelationKey, Entity: clientEntityType}
	if _, err := svc.policies.Authorize(ctx, areq); err != nil {
		return fmt.Errorf("cannot share thing %s with user %s: %s", thingID, userID, err)
	}

	policy := tpolicies.Policy{
		Subject: userID,
		Object:  groupID,
		Actions: actions,
	}
	if _, err = svc.policies.AddPolicy(ctx, token, policy); err != nil {
		return err
	}

	return nil
}

func (svc service) identify(ctx context.Context, token string) (string, error) {
	req := &upolicies.Token{Value: token}
	res, err := svc.uauth.Identify(ctx, req)
	if err != nil {
		return "", errors.Wrap(errors.ErrAuthorization, err)
	}
	return res.GetId(), nil
}

func (svc service) authorize(ctx context.Context, subject, object, action string) error {
	// If the user is admin, skip authorization.
	if err := svc.checkAdmin(ctx, subject, thingsObjectKey, action); err == nil {
		return nil
	}

	policy := tpolicies.AccessRequest{Subject: subject, Object: object, Action: action, Entity: clientEntityType}

	_, err := svc.policies.Authorize(ctx, policy)
	return err
}

// TODO : Only accept token as parameter since object and action are irrelevant.
func (svc service) checkAdmin(ctx context.Context, subject, object, action string) error {
	req := &upolicies.AuthorizeReq{
		Sub:        subject,
		Obj:        object,
		Act:        action,
		EntityType: clientEntityType,
	}
	res, err := svc.uauth.Authorize(ctx, req)
	if err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	if !res.GetAuthorized() {
		return errors.ErrAuthorization
	}
	return nil
}
