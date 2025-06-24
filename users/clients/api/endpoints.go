// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/users/clients"
)

func registrationEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createClientReq)
		if err := req.validate(); err != nil {
			return createClientRes{}, err
		}
		client, err := svc.RegisterClient(ctx, req.token, req.client)
		if err != nil {
			return createClientRes{}, err
		}
		ucr := createClientRes{
			Client:  client,
			created: true,
		}

		return ucr, nil
	}
}

func viewClientEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewClientReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		client, err := svc.ViewClient(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return viewClientRes{Client: client}, nil
	}
}

func viewProfileEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewProfileReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		client, err := svc.ViewProfile(ctx, req.token)
		if err != nil {
			return nil, err
		}
		return viewClientRes{
			Client: client,
		}, nil
	}
}

func listClientsEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listClientsReq)
		if err := req.validate(); err != nil {
			return mfclients.ClientsPage{}, err
		}

		pm := mfclients.Page{
			SharedBy: req.sharedBy,
			Status:   req.status,
			Offset:   req.offset,
			Limit:    req.limit,
			Owner:    req.owner,
			Name:     req.name,
			Tag:      req.tag,
			Metadata: req.metadata,
			Identity: req.identity,
		}
		page, err := svc.ListClients(ctx, req.token, pm)
		if err != nil {
			return mfclients.ClientsPage{}, err
		}

		res := clientsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Clients: []viewClientRes{},
		}
		for _, client := range page.Clients {
			res.Clients = append(res.Clients, viewClientRes{Client: client})
		}

		return res, nil
	}
}

func listMembersEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMembersReq)
		if err := req.validate(); err != nil {
			return memberPageRes{}, err
		}
		page, err := svc.ListMembers(ctx, req.token, req.groupID, req.Page)
		if err != nil {
			return memberPageRes{}, err
		}
		return buildMembersResponse(page), nil
	}
}

func updateClientEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		client := mfclients.Client{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		client, err := svc.UpdateClient(ctx, req.token, client)
		if err != nil {
			return nil, err
		}
		return updateClientRes{Client: client}, nil
	}
}

func updateClientTagsEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientTagsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		client := mfclients.Client{
			ID:   req.id,
			Tags: req.Tags,
		}
		client, err := svc.UpdateClientTags(ctx, req.token, client)
		if err != nil {
			return nil, err
		}
		return updateClientRes{Client: client}, nil
	}
}

func updateClientIdentityEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientIdentityReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		client, err := svc.UpdateClientIdentity(ctx, req.token, req.id, req.Identity)
		if err != nil {
			return nil, err
		}
		return updateClientRes{Client: client}, nil
	}
}

// Password reset request endpoint.
// When successful password reset link is generated.
// Link is generated using MF_TOKEN_RESET_ENDPOINT env.
// and value from Referer header for host.
// {Referer}+{MF_TOKEN_RESET_ENDPOINT}+{token=TOKEN}
// http://mainflux.com/reset-request?token=xxxxxxxxxxx.
// Email with a link is being sent to the user.
// When user clicks on a link it should get the ui with form to
// enter new password, when form is submitted token and new password
// must be sent as PUT request to 'password/reset' passwordResetEndpoint.
func passwordResetRequestEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwResetReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.GenerateResetToken(ctx, req.Email, req.Host); err != nil {
			return nil, err
		}

		return passwResetReqRes{Msg: MailSent}, nil
	}
}

// This is endpoint that actually sets new password in password reset flow.
// When user clicks on a link in email finally ends on this endpoint as explained in
// the comment above.
func passwordResetEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetTokenReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.ResetSecret(ctx, req.Token, req.Password); err != nil {
			return nil, err
		}
		return passwChangeRes{}, nil
	}
}

func updateClientSecretEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		client, err := svc.UpdateClientSecret(ctx, req.token, req.OldSecret, req.NewSecret)
		if err != nil {
			return nil, err
		}
		return updateClientRes{Client: client}, nil
	}
}

func updateClientOwnerEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientOwnerReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		client := mfclients.Client{
			ID:    req.id,
			Owner: req.Owner,
		}

		client, err := svc.UpdateClientOwner(ctx, req.token, client)
		if err != nil {
			return nil, err
		}
		return updateClientRes{Client: client}, nil
	}
}

func issueTokenEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loginClientReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		token, err := svc.IssueToken(ctx, req.Identity, req.Secret)
		if err != nil {
			return nil, err
		}
		return tokenRes{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			AccessType:   token.AccessType,
		}, nil
	}
}

func refreshTokenEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(tokenReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		token, err := svc.RefreshToken(ctx, req.RefreshToken)
		if err != nil {
			return nil, err
		}

		return tokenRes{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			AccessType:   token.AccessType,
		}, nil
	}
}

func enableClientEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeClientStatusReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		client, err := svc.EnableClient(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return deleteClientRes{Client: client}, nil
	}
}

func disableClientEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeClientStatusReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		client, err := svc.DisableClient(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return deleteClientRes{Client: client}, nil
	}
}

func buildMembersResponse(cp mfclients.MembersPage) memberPageRes {
	res := memberPageRes{
		pageRes: pageRes{
			Total:  cp.Total,
			Offset: cp.Offset,
			Limit:  cp.Limit,
		},
		Members: []viewMembersRes{},
	}
	for _, client := range cp.Members {
		res.Members = append(res.Members, viewMembersRes{Client: client})
	}
	return res
}
