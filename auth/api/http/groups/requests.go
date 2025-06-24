package groups

import (
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/internal/apiutil"
)

type createGroupReq struct {
	token       string
	Name        string                 `json:"name,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req createGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if len(req.Name) > maxNameSize || req.Name == "" {
		return apiutil.ErrNameSize
	}

	return nil
}

type updateGroupReq struct {
	token       string
	id          string
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type listGroupsReq struct {
	token string
	id    string
	level uint64
	// - `true`  - result is JSON tree representing groups hierarchy,
	// - `false` - result is JSON array of groups.
	tree     bool
	metadata auth.GroupMetadata
}

func (req listGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.level > auth.MaxLevel || req.level < auth.MinLevel {
		return apiutil.ErrMaxLevelExceeded
	}

	return nil
}

type listMembersReq struct {
	token     string
	id        string
	groupType string
	offset    uint64
	limit     uint64
	tree      bool
	metadata  auth.GroupMetadata
}

func (req listMembersReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type listMembershipsReq struct {
	token    string
	id       string
	offset   uint64
	limit    uint64
	metadata auth.GroupMetadata
}

func (req listMembershipsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type assignReq struct {
	token   string
	groupID string
	Type    string   `json:"type,omitempty"`
	Members []string `json:"members"`
}

func (req assignReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.Type == "" {
		return apiutil.ErrMissingMemberType
	}

	if req.groupID == "" {
		return apiutil.ErrMissingID
	}

	if len(req.Members) == 0 {
		return apiutil.ErrEmptyList
	}

	return nil
}

type shareGroupAccessReq struct {
	token        string
	userGroupID  string
	ThingGroupID string `json:"thing_group_id"`
}

func (req shareGroupAccessReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.ThingGroupID == "" || req.userGroupID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type unassignReq struct {
	assignReq
}

func (req unassignReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.groupID == "" {
		return apiutil.ErrMissingID
	}

	if len(req.Members) == 0 {
		return apiutil.ErrEmptyList
	}

	return nil
}

type groupReq struct {
	token string
	id    string
}

func (req groupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}
