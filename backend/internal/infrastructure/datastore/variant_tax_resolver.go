package datastore

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type variantTaxResolver struct{}

// NewVariantTaxResolver returns a VariantTaxResolver backed by sqlc.
func NewVariantTaxResolver() gateway.VariantTaxResolver { return &variantTaxResolver{} }

func (r *variantTaxResolver) RatesForVariants(ctx context.Context, variantIDs []string) ([]gateway.VariantTaxRate, error) {
	if len(variantIDs) == 0 {
		return nil, nil
	}
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uids, err := uuidArray(variantIDs)
	if err != nil {
		return nil, errors.BadRequest("invalid variant id")
	}
	rows, err := sqlc.New(dbtx).ListTaxRatesForVariants(ctx, uids)
	if err != nil {
		return nil, errors.Wrap(err, "list tax rates for variants")
	}
	out := make([]gateway.VariantTaxRate, 0, len(rows))
	for _, row := range rows {
		out = append(out, gateway.VariantTaxRate{
			VariantID:    uuidString(row.VariantID),
			TaxRateID:    uuidString(row.TaxRateID),
			NameSnapshot: row.NameSnapshot,
			Rate:         row.Rate,
			IsInclusive:  row.IsInclusive,
			IsCompound:   row.IsCompound,
			Sequence:     row.Sequence,
		})
	}
	return out, nil
}
