package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/internal/postgres"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	mfgroups "github.com/mainflux/mainflux/pkg/groups"
)

var _ mfgroups.Repository = (*groupRepository)(nil)

type groupRepository struct {
	db postgres.Database
}

// New instantiates a PostgreSQL implementation of group
// repository.
func New(db postgres.Database) mfgroups.Repository {
	return &groupRepository{
		db: db,
	}
}

// TODO - check parent group write access.
func (repo groupRepository) Save(ctx context.Context, g mfgroups.Group) (mfgroups.Group, error) {
	q := `INSERT INTO groups (name, description, id, owner_id, parent_id, metadata, created_at, status)
		VALUES (:name, :description, :id, :owner_id, :parent_id, :metadata, :created_at, :status)
		RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, status;`
	dbg, err := toDBGroup(g)
	if err != nil {
		return mfgroups.Group{}, err
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbg)
	if err != nil {
		return mfgroups.Group{}, postgres.HandleError(err, errors.ErrCreateEntity)
	}

	defer row.Close()
	row.Next()
	dbg = dbGroup{}
	if err := row.StructScan(&dbg); err != nil {
		return mfgroups.Group{}, err
	}

	return toGroup(dbg)
}

func (repo groupRepository) Memberships(ctx context.Context, clientID string, gm mfgroups.GroupsPage) (mfgroups.MembershipsPage, error) {
	var q string
	query, err := buildQuery(gm)
	if err != nil {
		return mfgroups.MembershipsPage{}, err
	}
	if gm.ID != "" {
		q = buildHierachy(gm)
	}
	if gm.ID == "" {
		q = `SELECT g.id, g.owner_id, COALESCE(g.parent_id, '') AS parent_id, g.name, g.description,
		g.metadata, g.created_at, g.updated_at, g.updated_by, g.status FROM groups g`
	}
	aq := ""
	// If not admin, the client needs to have a g_list action on the group or they are the owner.
	if gm.Subject != "" {
		aq = `AND policies.object IN (SELECT object FROM policies WHERE subject = :subject AND :action=ANY(actions)) OR g.owner_id = :subject`
	}
	q = fmt.Sprintf(`%s INNER JOIN policies ON g.id=policies.object %s AND policies.subject = :client_id %s
			ORDER BY g.updated_at LIMIT :limit OFFSET :offset;`, q, query, aq)

	dbPage, err := toDBGroupPage(gm)
	if err != nil {
		return mfgroups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
	}
	dbPage.ClientID = clientID
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return mfgroups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
	}
	defer rows.Close()

	var items []mfgroups.Group
	for rows.Next() {
		dbg := dbGroup{}
		if err := rows.StructScan(&dbg); err != nil {
			return mfgroups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
		}
		group, err := toGroup(dbg)
		if err != nil {
			return mfgroups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
		}
		items = append(items, group)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM groups g INNER JOIN policies
        ON g.id=policies.object %s AND policies.subject = :client_id`, query)

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return mfgroups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
	}
	page := mfgroups.MembershipsPage{
		Memberships: items,
		Page: mfgroups.Page{
			Total: total,
		},
	}

	return page, nil
}

func (repo groupRepository) Update(ctx context.Context, g mfgroups.Group) (mfgroups.Group, error) {
	var query []string
	var upq string
	if g.Name != "" {
		query = append(query, "name = :name,")
	}
	if g.Description != "" {
		query = append(query, "description = :description,")
	}
	if g.Metadata != nil {
		query = append(query, "metadata = :metadata,")
	}
	if len(query) > 0 {
		upq = strings.Join(query, " ")
	}
	g.Status = mfclients.EnabledStatus
	q := fmt.Sprintf(`UPDATE groups SET %s updated_at = :updated_at, updated_by = :updated_by
		WHERE id = :id AND status = :status
		RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, updated_at, updated_by, status`, upq)

	dbu, err := toDBGroup(g)
	if err != nil {
		return mfgroups.Group{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}

	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return mfgroups.Group{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return mfgroups.Group{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	dbu = dbGroup{}
	if err := row.StructScan(&dbu); err != nil {
		return mfgroups.Group{}, errors.Wrap(err, errors.ErrUpdateEntity)
	}
	return toGroup(dbu)
}

func (repo groupRepository) ChangeStatus(ctx context.Context, group mfgroups.Group) (mfgroups.Group, error) {
	qc := `UPDATE groups SET status = :status WHERE id = :id
	RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, updated_at, updated_by, status`

	dbg, err := toDBGroup(group)
	if err != nil {
		return mfgroups.Group{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}
	row, err := repo.db.NamedQueryContext(ctx, qc, dbg)
	if err != nil {
		return mfgroups.Group{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return mfgroups.Group{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	dbg = dbGroup{}
	if err := row.StructScan(&dbg); err != nil {
		return mfgroups.Group{}, errors.Wrap(err, errors.ErrUpdateEntity)
	}

	return toGroup(dbg)
}

func (repo groupRepository) RetrieveByID(ctx context.Context, id string) (mfgroups.Group, error) {
	q := `SELECT id, name, owner_id, COALESCE(parent_id, '') AS parent_id, description, metadata, created_at, updated_at, updated_by, status FROM groups
	    WHERE id = :id`

	dbg := dbGroup{
		ID: id,
	}

	row, err := repo.db.NamedQueryContext(ctx, q, dbg)
	if err != nil {
		if err == sql.ErrNoRows {
			return mfgroups.Group{}, errors.Wrap(errors.ErrNotFound, err)
		}
		return mfgroups.Group{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	defer row.Close()
	row.Next()
	dbg = dbGroup{}
	if err := row.StructScan(&dbg); err != nil {
		return mfgroups.Group{}, errors.Wrap(errors.ErrNotFound, err)
	}

	return toGroup(dbg)
}

func (repo groupRepository) RetrieveAll(ctx context.Context, gm mfgroups.GroupsPage) (mfgroups.GroupsPage, error) {
	var q string
	query, err := buildQuery(gm)
	if err != nil {
		return mfgroups.GroupsPage{}, err
	}

	if gm.ID != "" {
		q = buildHierachy(gm)
	}
	if gm.ID == "" {
		q = `SELECT DISTINCT g.id, g.owner_id, COALESCE(g.parent_id, '') AS parent_id, g.name, g.description,
		g.metadata, g.created_at, g.updated_at, g.updated_by, g.status FROM groups g`
	}
	q = fmt.Sprintf("%s %s ORDER BY g.updated_at LIMIT :limit OFFSET :offset;", q, query)

	dbPage, err := toDBGroupPage(gm)
	if err != nil {
		return mfgroups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return mfgroups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	defer rows.Close()

	items, err := repo.processRows(rows)
	if err != nil {
		return mfgroups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}

	cq := "SELECT COUNT(*) FROM groups g"
	if query != "" {
		cq = fmt.Sprintf(" %s %s", cq, query)
	}

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return mfgroups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}

	page := gm
	page.Groups = items
	page.Total = total

	return page, nil
}

func buildHierachy(gm mfgroups.GroupsPage) string {
	query := ""
	switch {
	case gm.Direction >= 0: // ancestors
		query = `WITH RECURSIVE groups_cte as (
			SELECT id, COALESCE(parent_id, '') AS parent_id, owner_id, name, description, metadata, created_at, updated_at, updated_by, status, 0 as level from groups WHERE id = :id
			UNION SELECT x.id, COALESCE(x.parent_id, '') AS parent_id, x.owner_id, x.name, x.description, x.metadata, x.created_at, x.updated_at, x.updated_by, x.status, level - 1 from groups x
			INNER JOIN groups_cte a ON a.parent_id = x.id
		) SELECT * FROM groups_cte g`

	case gm.Direction < 0: // descendants
		query = `WITH RECURSIVE groups_cte as (
			SELECT id, COALESCE(parent_id, '') AS parent_id, owner_id, name, description, metadata, created_at, updated_at, updated_by, status, 0 as level, CONCAT('', '', id) as path from groups WHERE id = :id
			UNION SELECT x.id, COALESCE(x.parent_id, '') AS parent_id, x.owner_id, x.name, x.description, x.metadata, x.created_at, x.updated_at, x.updated_by, x.status, level + 1, CONCAT(path, '.', x.id) as path from groups x
			INNER JOIN groups_cte d ON d.id = x.parent_id
		) SELECT * FROM groups_cte g`
	}
	return query
}
func buildQuery(gm mfgroups.GroupsPage) (string, error) {
	queries := []string{}

	if gm.Name != "" {
		queries = append(queries, "g.name = :name")
	}
	if gm.Status != mfclients.AllStatus {
		queries = append(queries, "g.status = :status")
	}
	if gm.OwnerID != "" {
		queries = append(queries, "g.owner_id = :owner_id")
	}
	if gm.Tag != "" {
		queries = append(queries, ":tag = ANY(c.tags)")
	}
	if gm.Subject != "" {
		queries = append(queries, "(g.owner_id = :owner_id OR id IN (SELECT object as id FROM policies WHERE subject = :subject AND :action=ANY(actions)))")
	}
	if len(gm.Metadata) > 0 {
		queries = append(queries, "'g.metadata @> :metadata'")
	}
	if len(queries) > 0 {
		return fmt.Sprintf("WHERE %s", strings.Join(queries, " AND ")), nil
	}
	return "", nil
}

type dbGroup struct {
	ID          string           `db:"id"`
	ParentID    *string          `db:"parent_id,omitempty"`
	OwnerID     string           `db:"owner_id,omitempty"`
	Name        string           `db:"name"`
	Description string           `db:"description,omitempty"`
	Level       int              `db:"level"`
	Path        string           `db:"path,omitempty"`
	Metadata    []byte           `db:"metadata,omitempty"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   sql.NullTime     `db:"updated_at,omitempty"`
	UpdatedBy   *string          `db:"updated_by,omitempty"`
	Status      mfclients.Status `db:"status"`
}

func toDBGroup(g mfgroups.Group) (dbGroup, error) {
	data := []byte("{}")
	if len(g.Metadata) > 0 {
		b, err := json.Marshal(g.Metadata)
		if err != nil {
			return dbGroup{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		data = b
	}
	var parentID *string
	if g.Parent != "" {
		parentID = &g.Parent
	}
	var updatedAt sql.NullTime
	if !g.UpdatedAt.IsZero() {
		updatedAt = sql.NullTime{Time: g.UpdatedAt, Valid: true}
	}
	var updatedBy *string
	if g.UpdatedBy != "" {
		updatedBy = &g.UpdatedBy
	}
	return dbGroup{
		ID:          g.ID,
		Name:        g.Name,
		ParentID:    parentID,
		OwnerID:     g.Owner,
		Description: g.Description,
		Metadata:    data,
		Path:        g.Path,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   updatedAt,
		UpdatedBy:   updatedBy,
		Status:      g.Status,
	}, nil
}

func toGroup(g dbGroup) (mfgroups.Group, error) {
	var metadata mfclients.Metadata
	if g.Metadata != nil {
		if err := json.Unmarshal([]byte(g.Metadata), &metadata); err != nil {
			return mfgroups.Group{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}
	var parentID string
	if g.ParentID != nil {
		parentID = *g.ParentID
	}
	var updatedAt time.Time
	if g.UpdatedAt.Valid {
		updatedAt = g.UpdatedAt.Time
	}
	var updatedBy string
	if g.UpdatedBy != nil {
		updatedBy = *g.UpdatedBy
	}

	return mfgroups.Group{
		ID:          g.ID,
		Name:        g.Name,
		Parent:      parentID,
		Owner:       g.OwnerID,
		Description: g.Description,
		Metadata:    metadata,
		Level:       g.Level,
		Path:        g.Path,
		UpdatedAt:   updatedAt,
		UpdatedBy:   updatedBy,
		CreatedAt:   g.CreatedAt,
		Status:      g.Status,
	}, nil
}

func toDBGroupPage(pm mfgroups.GroupsPage) (dbGroupPage, error) {
	level := mfgroups.MaxLevel
	if pm.Level < mfgroups.MaxLevel {
		level = pm.Level
	}
	data := []byte("{}")
	if len(pm.Metadata) > 0 {
		b, err := json.Marshal(pm.Metadata)
		if err != nil {
			return dbGroupPage{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		data = b
	}
	return dbGroupPage{
		ID:       pm.ID,
		Name:     pm.Name,
		Metadata: data,
		Path:     pm.Path,
		Level:    level,
		Total:    pm.Total,
		Offset:   pm.Offset,
		Limit:    pm.Limit,
		ParentID: pm.ID,
		OwnerID:  pm.OwnerID,
		Subject:  pm.Subject,
		Action:   pm.Action,
		Status:   pm.Status,
	}, nil
}

type dbGroupPage struct {
	ClientID string           `db:"client_id"`
	ID       string           `db:"id"`
	Name     string           `db:"name"`
	ParentID string           `db:"parent_id"`
	OwnerID  string           `db:"owner_id"`
	Metadata []byte           `db:"metadata"`
	Path     string           `db:"path"`
	Level    uint64           `db:"level"`
	Total    uint64           `db:"total"`
	Limit    uint64           `db:"limit"`
	Offset   uint64           `db:"offset"`
	Subject  string           `db:"subject"`
	Action   string           `db:"action"`
	Status   mfclients.Status `db:"status"`
}

func (gr groupRepository) processRows(rows *sqlx.Rows) ([]mfgroups.Group, error) {
	var items []mfgroups.Group
	for rows.Next() {
		dbg := dbGroup{}
		if err := rows.StructScan(&dbg); err != nil {
			return items, err
		}
		group, err := toGroup(dbg)
		if err != nil {
			return items, err
		}
		items = append(items, group)
	}
	return items, nil
}
