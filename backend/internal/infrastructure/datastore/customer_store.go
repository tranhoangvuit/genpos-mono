package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type customerStore struct{}

// NewCustomerReader returns a CustomerReader backed by sqlc.
func NewCustomerReader() gateway.CustomerReader { return &customerStore{} }

// NewCustomerWriter returns a CustomerWriter backed by sqlc.
func NewCustomerWriter() gateway.CustomerWriter { return &customerStore{} }

func (s *customerStore) GetByID(ctx context.Context, id string) (*entity.Customer, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid customer id")
	}
	r, err := sqlc.New(dbtx).GetCustomerByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer not found")
		}
		return nil, errors.Wrap(err, "get customer by id")
	}
	return toCustomerEntity(r.ID, r.OrgID, r.Name, r.Email, r.Phone, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerStore) GetByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	r, err := sqlc.New(dbtx).GetCustomerByEmail(ctx, textOrNull(email))
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer not found")
		}
		return nil, errors.Wrap(err, "get customer by email")
	}
	return toCustomerEntity(r.ID, r.OrgID, r.Name, r.Email, r.Phone, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerStore) GetByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	r, err := sqlc.New(dbtx).GetCustomerByPhone(ctx, textOrNull(phone))
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer not found")
		}
		return nil, errors.Wrap(err, "get customer by phone")
	}
	return toCustomerEntity(r.ID, r.OrgID, r.Name, r.Email, r.Phone, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerStore) ListSummaries(ctx context.Context) ([]*entity.CustomerListItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListCustomerSummaries(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list customer summaries")
	}
	out := make([]*entity.CustomerListItem, 0, len(rows))
	for _, r := range rows {
		names, _ := r.GroupNames.(string)
		out = append(out, &entity.CustomerListItem{
			ID:         uuidString(r.ID),
			Name:       r.Name,
			Email:      textString(r.Email),
			Phone:      textString(r.Phone),
			GroupNames: names,
		})
	}
	return out, nil
}

func (s *customerStore) ListGroupIDsByCustomer(ctx context.Context, customerID string) ([]string, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(customerID)
	if err != nil {
		return nil, errors.BadRequest("invalid customer id")
	}
	rows, err := sqlc.New(dbtx).ListCustomerGroupMembersByCustomer(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list customer group members")
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, uuidString(r.GroupID))
	}
	return out, nil
}

func (s *customerStore) Create(ctx context.Context, params gateway.CreateCustomerParams) (*entity.Customer, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	r, err := sqlc.New(dbtx).CreateCustomer(ctx, sqlc.CreateCustomerParams{
		OrgID: orgID,
		Name:  params.Name,
		Email: textOrNull(params.Email),
		Phone: textOrNull(params.Phone),
		Notes: textOrNull(params.Notes),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create customer")
	}
	return toCustomerEntity(r.ID, r.OrgID, r.Name, r.Email, r.Phone, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerStore) Update(ctx context.Context, params gateway.UpdateCustomerParams) (*entity.Customer, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid customer id")
	}
	r, err := sqlc.New(dbtx).UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
		ID:    id,
		Name:  params.Name,
		Email: textOrNull(params.Email),
		Phone: textOrNull(params.Phone),
		Notes: textOrNull(params.Notes),
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer not found")
		}
		return nil, errors.Wrap(err, "update customer")
	}
	return toCustomerEntity(r.ID, r.OrgID, r.Name, r.Email, r.Phone, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid customer id")
	}
	if err := sqlc.New(dbtx).SoftDeleteCustomer(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete customer")
	}
	return nil
}

func (s *customerStore) ReplaceGroups(ctx context.Context, orgID, customerID string, groupIDs []string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	custUID, err := parseUUID(customerID)
	if err != nil {
		return errors.BadRequest("invalid customer id")
	}
	orgUID, err := parseUUID(orgID)
	if err != nil {
		return errors.BadRequest("invalid org id")
	}
	q := sqlc.New(dbtx)
	if err := q.DeleteCustomerGroupMembersByCustomer(ctx, custUID); err != nil {
		return errors.Wrap(err, "clear customer groups")
	}
	for _, gid := range groupIDs {
		if gid == "" {
			continue
		}
		groupUID, err := parseUUID(gid)
		if err != nil {
			return errors.BadRequest("invalid group id")
		}
		if err := q.InsertCustomerGroupMember(ctx, sqlc.InsertCustomerGroupMemberParams{
			OrgID:      orgUID,
			GroupID:    groupUID,
			CustomerID: custUID,
		}); err != nil {
			return errors.Wrap(err, "insert customer group member")
		}
	}
	return nil
}

func toCustomerEntity(id, orgID pgtype.UUID, name string, email, phone, notes pgtype.Text,
	createdAt, updatedAt pgtype.Timestamptz) *entity.Customer {
	return &entity.Customer{
		ID:        uuidString(id),
		OrgID:     uuidString(orgID),
		Name:      name,
		Email:     textString(email),
		Phone:     textString(phone),
		Notes:     textString(notes),
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}
