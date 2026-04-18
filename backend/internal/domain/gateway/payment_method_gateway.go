package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=payment_method_gateway.go -destination=mock/mock_payment_method_gateway.go -package=mock

type CreatePaymentMethodParams struct {
	OrgID     string
	Name      string
	Type      string
	IsActive  bool
	SortOrder int32
}

type UpdatePaymentMethodParams struct {
	ID        string
	Name      string
	Type      string
	IsActive  bool
	SortOrder int32
}

// PaymentMethodReader reads payment methods.
type PaymentMethodReader interface {
	List(ctx context.Context) ([]*entity.PaymentMethod, error)
}

// PaymentMethodWriter mutates payment methods within a tenant-scoped tx.
type PaymentMethodWriter interface {
	Create(ctx context.Context, params CreatePaymentMethodParams) (*entity.PaymentMethod, error)
	Update(ctx context.Context, params UpdatePaymentMethodParams) (*entity.PaymentMethod, error)
	SoftDelete(ctx context.Context, id string) error
}
