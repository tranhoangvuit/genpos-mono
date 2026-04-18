package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=customer_gateway.go -destination=mock/mock_customer_gateway.go -package=mock

// ----- Customer --------------------------------------------------------------

type CreateCustomerParams struct {
	OrgID string
	Name  string
	Email string
	Phone string
	Notes string
}

type UpdateCustomerParams struct {
	ID    string
	Name  string
	Email string
	Phone string
	Notes string
}

// CustomerReader fetches customers.
type CustomerReader interface {
	GetByID(ctx context.Context, id string) (*entity.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*entity.Customer, error)
	ListGroupIDsByCustomer(ctx context.Context, customerID string) ([]string, error)
	ListSummaries(ctx context.Context) ([]*entity.CustomerListItem, error)
}

// CustomerWriter mutates customers and their group memberships.
type CustomerWriter interface {
	Create(ctx context.Context, params CreateCustomerParams) (*entity.Customer, error)
	Update(ctx context.Context, params UpdateCustomerParams) (*entity.Customer, error)
	SoftDelete(ctx context.Context, id string) error

	ReplaceGroups(ctx context.Context, orgID, customerID string, groupIDs []string) error
}

// ----- CustomerGroup ---------------------------------------------------------

type CreateCustomerGroupParams struct {
	OrgID         string
	Name          string
	Description   string
	DiscountType  string
	DiscountValue string
}

type UpdateCustomerGroupParams struct {
	ID            string
	Name          string
	Description   string
	DiscountType  string
	DiscountValue string
}

// CustomerGroupReader lists and retrieves customer groups.
type CustomerGroupReader interface {
	List(ctx context.Context) ([]*entity.CustomerGroup, error)
	GetByID(ctx context.Context, id string) (*entity.CustomerGroup, error)
	GetByName(ctx context.Context, name string) (*entity.CustomerGroup, error)
}

// CustomerGroupWriter mutates customer groups.
type CustomerGroupWriter interface {
	Create(ctx context.Context, params CreateCustomerGroupParams) (*entity.CustomerGroup, error)
	Update(ctx context.Context, params UpdateCustomerGroupParams) (*entity.CustomerGroup, error)
	SoftDelete(ctx context.Context, id string) error
}
