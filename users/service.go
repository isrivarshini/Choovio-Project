// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	mgauth "github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/authn"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policies"
	"github.com/absmach/magistrala/users/postgres"
	"golang.org/x/sync/errgroup"
)

var (
	errIssueToken            = errors.New("failed to issue token")
	errFailedPermissionsList = errors.New("failed to list permissions")
	errRecoveryToken         = errors.New("failed to generate password recovery token")
	errLoginDisableUser      = errors.New("failed to login in disabled user")
)

type service struct {
	token      magistrala.TokenServiceClient
	clients    postgres.Repository
	idProvider magistrala.IDProvider
	policies   policies.Service
	hasher     Hasher
	email      Emailer
}

// NewService returns a new Users service implementation.
func NewService(token magistrala.TokenServiceClient, crepo postgres.Repository, policyService policies.Service, emailer Emailer, hasher Hasher, idp magistrala.IDProvider) Service {
	return service{
		token:      token,
		clients:    crepo,
		policies:   policyService,
		hasher:     hasher,
		email:      emailer,
		idProvider: idp,
	}
}

func (svc service) RegisterClient(ctx context.Context, session authn.Session, cli mgclients.Client, selfRegister bool) (rc mgclients.Client, err error) {
	if !selfRegister {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return mgclients.Client{}, err
		}
	}

	clientID, err := svc.idProvider.ID()
	if err != nil {
		return mgclients.Client{}, err
	}

	if cli.Credentials.Secret != "" {
		hash, err := svc.hasher.Hash(cli.Credentials.Secret)
		if err != nil {
			return mgclients.Client{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
		}
		cli.Credentials.Secret = hash
	}

	if cli.Status != mgclients.DisabledStatus && cli.Status != mgclients.EnabledStatus {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrMalformedEntity, svcerr.ErrInvalidStatus)
	}
	if cli.Role != mgclients.UserRole && cli.Role != mgclients.AdminRole {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrMalformedEntity, svcerr.ErrInvalidRole)
	}
	cli.ID = clientID
	cli.CreatedAt = time.Now()

	if err := svc.addClientPolicy(ctx, cli.ID, cli.Role); err != nil {
		return mgclients.Client{}, err
	}
	defer func() {
		if err != nil {
			if errRollback := svc.addClientPolicyRollback(ctx, cli.ID, cli.Role); errRollback != nil {
				err = errors.Wrap(errors.Wrap(errors.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	client, err := svc.clients.Save(ctx, cli)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	return client, nil
}

func (svc service) IssueToken(ctx context.Context, identity, secret string) (*magistrala.Token, error) {
	dbUser, err := svc.clients.RetrieveByIdentity(ctx, identity)
	if err != nil {
		return &magistrala.Token{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if err := svc.hasher.Compare(secret, dbUser.Credentials.Secret); err != nil {
		return &magistrala.Token{}, errors.Wrap(svcerr.ErrLogin, err)
	}

	token, err := svc.token.Issue(ctx, &magistrala.IssueReq{UserId: dbUser.ID, Type: uint32(mgauth.AccessKey)})
	if err != nil {
		return &magistrala.Token{}, errors.Wrap(errIssueToken, err)
	}

	return token, err
}

func (svc service) RefreshToken(ctx context.Context, session authn.Session, refreshToken string) (*magistrala.Token, error) {
	dbUser, err := svc.clients.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return &magistrala.Token{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if dbUser.Status == mgclients.DisabledStatus {
		return &magistrala.Token{}, errors.Wrap(svcerr.ErrAuthentication, errLoginDisableUser)
	}

	return svc.token.Refresh(ctx, &magistrala.RefreshReq{RefreshToken: refreshToken})
}

func (svc service) ViewClient(ctx context.Context, session authn.Session, id string) (mgclients.Client, error) {
	client, err := svc.clients.RetrieveByID(ctx, id)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if session.UserID != id {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return mgclients.Client{Name: client.Name, ID: client.ID}, nil
		}
	}

	client.Credentials.Secret = ""

	return client, nil
}

func (svc service) ViewProfile(ctx context.Context, session authn.Session) (mgclients.Client, error) {
	client, err := svc.clients.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	client.Credentials.Secret = ""

	return client, nil
}

func (svc service) ListClients(ctx context.Context, session authn.Session, pm mgclients.Page) (mgclients.ClientsPage, error) {
	if err := svc.checkSuperAdmin(ctx, session); err != nil {
		return mgclients.ClientsPage{}, err
	}

	pm.Role = mgclients.AllRole
	pg, err := svc.clients.RetrieveAll(ctx, pm)
	if err != nil {
		return mgclients.ClientsPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return pg, err
}

func (svc service) SearchUsers(ctx context.Context, pm mgclients.Page) (mgclients.ClientsPage, error) {
	page := mgclients.Page{
		Offset: pm.Offset,
		Limit:  pm.Limit,
		Name:   pm.Name,
		Id:     pm.Id,
		Role:   mgclients.UserRole,
	}

	cp, err := svc.clients.SearchClients(ctx, page)
	if err != nil {
		return mgclients.ClientsPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return cp, nil
}

func (svc service) UpdateClient(ctx context.Context, session authn.Session, cli mgclients.Client) (mgclients.Client, error) {
	if session.UserID != cli.ID {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return mgclients.Client{}, err
		}
	}

	client := mgclients.Client{
		ID:        cli.ID,
		Name:      cli.Name,
		Metadata:  cli.Metadata,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}

	client, err := svc.clients.Update(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) UpdateClientTags(ctx context.Context, session authn.Session, cli mgclients.Client) (mgclients.Client, error) {
	if session.UserID != cli.ID {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return mgclients.Client{}, err
		}
	}

	client := mgclients.Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	client, err := svc.clients.UpdateTags(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	return client, nil
}

func (svc service) UpdateClientIdentity(ctx context.Context, session authn.Session, clientID, identity string) (mgclients.Client, error) {
	if session.UserID != clientID {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return mgclients.Client{}, err
		}
	}

	cli := mgclients.Client{
		ID: clientID,
		Credentials: mgclients.Credentials{
			Identity: identity,
		},
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	cli, err := svc.clients.UpdateIdentity(ctx, cli)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return cli, nil
}

func (svc service) GenerateResetToken(ctx context.Context, email, host string) error {
	client, err := svc.clients.RetrieveByIdentity(ctx, email)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}
	issueReq := &magistrala.IssueReq{
		UserId: client.ID,
		Type:   uint32(mgauth.RecoveryKey),
	}
	token, err := svc.token.Issue(ctx, issueReq)
	if err != nil {
		return errors.Wrap(errRecoveryToken, err)
	}

	return svc.SendPasswordReset(ctx, host, email, client.Name, token.AccessToken)
}

func (svc service) ResetSecret(ctx context.Context, session authn.Session, secret string) error {
	c, err := svc.clients.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	secret, err = svc.hasher.Hash(secret)
	if err != nil {
		return errors.Wrap(svcerr.ErrMalformedEntity, err)
	}
	c = mgclients.Client{
		ID: c.ID,
		Credentials: mgclients.Credentials{
			Identity: c.Credentials.Identity,
			Secret:   secret,
		},
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	if _, err := svc.clients.UpdateSecret(ctx, c); err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}

func (svc service) UpdateClientSecret(ctx context.Context, session authn.Session, oldSecret, newSecret string) (mgclients.Client, error) {
	dbClient, err := svc.clients.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if _, err := svc.IssueToken(ctx, dbClient.Credentials.Identity, oldSecret); err != nil {
		return mgclients.Client{}, err
	}
	newSecret, err = svc.hasher.Hash(newSecret)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
	}
	dbClient.Credentials.Secret = newSecret
	dbClient.UpdatedAt = time.Now()
	dbClient.UpdatedBy = session.UserID

	dbClient, err = svc.clients.UpdateSecret(ctx, dbClient)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	return dbClient, nil
}

func (svc service) SendPasswordReset(_ context.Context, host, email, user, token string) error {
	to := []string{email}
	return svc.email.SendPasswordReset(to, host, user, token)
}

func (svc service) UpdateClientRole(ctx context.Context, session authn.Session, cli mgclients.Client) (mgclients.Client, error) {
	if err := svc.checkSuperAdmin(ctx, session); err != nil {
		return mgclients.Client{}, err
	}
	client := mgclients.Client{
		ID:        cli.ID,
		Role:      cli.Role,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}

	if err := svc.updateClientPolicy(ctx, cli.ID, cli.Role); err != nil {
		return mgclients.Client{}, err
	}

	client, err := svc.clients.UpdateRole(ctx, client)
	if err != nil {
		// If failed to update role in DB, then revert back to platform admin policies in spicedb
		if errRollback := svc.updateClientPolicy(ctx, cli.ID, mgclients.UserRole); errRollback != nil {
			return mgclients.Client{}, errors.Wrap(errRollback, err)
		}
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) EnableClient(ctx context.Context, session authn.Session, id string) (mgclients.Client, error) {
	client := mgclients.Client{
		ID:        id,
		UpdatedAt: time.Now(),
		Status:    mgclients.EnabledStatus,
	}
	client, err := svc.changeClientStatus(ctx, session, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(mgclients.ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, session authn.Session, id string) (mgclients.Client, error) {
	client := mgclients.Client{
		ID:        id,
		UpdatedAt: time.Now(),
		Status:    mgclients.DisabledStatus,
	}
	client, err := svc.changeClientStatus(ctx, session, client)
	if err != nil {
		return mgclients.Client{}, err
	}

	return client, nil
}

func (svc service) changeClientStatus(ctx context.Context, session authn.Session, client mgclients.Client) (mgclients.Client, error) {
	if session.UserID != client.ID {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return mgclients.Client{}, err
		}
	}
	dbClient, err := svc.clients.RetrieveByID(ctx, client.ID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if dbClient.Status == client.Status {
		return mgclients.Client{}, errors.ErrStatusAlreadyAssigned
	}
	client.UpdatedBy = session.UserID

	client, err = svc.clients.ChangeStatus(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) DeleteClient(ctx context.Context, session authn.Session, id string) error {
	client := mgclients.Client{
		ID:        id,
		UpdatedAt: time.Now(),
		Status:    mgclients.DeletedStatus,
	}

	if _, err := svc.changeClientStatus(ctx, session, client); err != nil {
		return err
	}

	return nil
}

func (svc service) ListMembers(ctx context.Context, session authn.Session, objectKind, objectID string, pm mgclients.Page) (mgclients.MembersPage, error) {
	var objectType string
	switch objectKind {
	case policies.ThingsKind:
		objectType = policies.ThingType
	case policies.DomainsKind:
		objectType = policies.DomainType
	case policies.GroupsKind:
		fallthrough
	default:
		objectType = policies.GroupType
	}

	duids, err := svc.policies.ListAllSubjects(ctx, policies.Policy{
		SubjectType: policies.UserType,
		Permission:  pm.Permission,
		Object:      objectID,
		ObjectType:  objectType,
	})
	if err != nil {
		return mgclients.MembersPage{}, errors.Wrap(svcerr.ErrNotFound, err)
	}
	if len(duids.Policies) == 0 {
		return mgclients.MembersPage{
			Page: mgclients.Page{Total: 0, Offset: pm.Offset, Limit: pm.Limit},
		}, nil
	}

	var userIDs []string

	for _, domainUserID := range duids.Policies {
		_, userID := mgauth.DecodeDomainUserID(domainUserID)
		userIDs = append(userIDs, userID)
	}
	pm.IDs = userIDs

	cp, err := svc.clients.RetrieveAll(ctx, pm)
	if err != nil {
		return mgclients.MembersPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	for i, c := range cp.Clients {
		cp.Clients[i] = mgclients.Client{
			ID:        c.ID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Status:    c.Status,
		}
	}

	if pm.ListPerms && len(cp.Clients) > 0 {
		g, ctx := errgroup.WithContext(ctx)

		for i := range cp.Clients {
			// Copying loop variable "i" to avoid "loop variable captured by func literal"
			iter := i
			g.Go(func() error {
				return svc.retrieveObjectUsersPermissions(ctx, session.DomainID, objectType, objectID, &cp.Clients[iter])
			})
		}

		if err := g.Wait(); err != nil {
			return mgclients.MembersPage{}, err
		}
	}

	return mgclients.MembersPage{
		Page:    cp.Page,
		Members: cp.Clients,
	}, nil
}

func (svc service) retrieveObjectUsersPermissions(ctx context.Context, domainID, objectType, objectID string, client *mgclients.Client) error {
	userID := mgauth.EncodeDomainUserID(domainID, client.ID)
	permissions, err := svc.listObjectUserPermission(ctx, userID, objectType, objectID)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	client.Permissions = permissions
	return nil
}

func (svc service) listObjectUserPermission(ctx context.Context, userID, objectType, objectID string) ([]string, error) {
	permissions, err := svc.policies.ListPermissions(ctx, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Object:      objectID,
		ObjectType:  objectType,
	}, []string{})
	if err != nil {
		return []string{}, errors.Wrap(errFailedPermissionsList, err)
	}
	return permissions, nil
}

func (svc *service) checkSuperAdmin(ctx context.Context, session authn.Session) error {
	if !session.SuperAdmin {
		if err := svc.clients.CheckSuperAdmin(ctx, session.UserID); err != nil {
			return errors.Wrap(svcerr.ErrAuthorization, err)
		}
	}

	return nil
}

func (svc service) OAuthCallback(ctx context.Context, client mgclients.Client) (mgclients.Client, error) {
	rclient, err := svc.clients.RetrieveByIdentity(ctx, client.Credentials.Identity)
	if err != nil {
		switch errors.Contains(err, repoerr.ErrNotFound) {
		case true:
			rclient, err = svc.RegisterClient(ctx, authn.Session{}, client, true)
			if err != nil {
				return mgclients.Client{}, err
			}
		default:
			return mgclients.Client{}, err
		}
	}

	return mgclients.Client{
		ID:   rclient.ID,
		Role: rclient.Role,
	}, nil
}

func (svc service) OAuthAddClientPolicy(ctx context.Context, client mgclients.Client) error {
	return svc.addClientPolicy(ctx, client.ID, client.Role)
}

func (svc service) Identify(ctx context.Context, session authn.Session) (string, error) {
	return session.UserID, nil
}

func (svc service) addClientPolicy(ctx context.Context, userID string, role mgclients.Role) error {
	policyList := []policies.Policy{}

	policyList = append(policyList, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Relation:    policies.MemberRelation,
		ObjectType:  policies.PlatformType,
		Object:      policies.MagistralaObject,
	})

	if role == mgclients.AdminRole {
		policyList = append(policyList, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.MagistralaObject,
		})
	}
	err := svc.policies.AddPolicies(ctx, policyList)
	if err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}

	return nil
}

func (svc service) addClientPolicyRollback(ctx context.Context, userID string, role mgclients.Role) error {
	policyList := []policies.Policy{}

	policyList = append(policyList, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Relation:    policies.MemberRelation,
		ObjectType:  policies.PlatformType,
		Object:      policies.MagistralaObject,
	})

	if role == mgclients.AdminRole {
		policyList = append(policyList, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.MagistralaObject,
		})
	}
	err := svc.policies.DeletePolicies(ctx, policyList)
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}

	return nil
}

func (svc service) updateClientPolicy(ctx context.Context, userID string, role mgclients.Role) error {
	switch role {
	case mgclients.AdminRole:
		err := svc.policies.AddPolicy(ctx, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.MagistralaObject,
		})
		if err != nil {
			return errors.Wrap(svcerr.ErrAddPolicies, err)
		}

		return nil
	case mgclients.UserRole:
		fallthrough
	default:
		err := svc.policies.DeletePolicyFilter(ctx, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.MagistralaObject,
		})
		if err != nil {
			return errors.Wrap(svcerr.ErrDeletePolicies, err)
		}

		return nil
	}
}
