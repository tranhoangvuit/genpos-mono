package entity

import "time"

// TaxRate is a VAT/sales-tax rate configured by an org.
type TaxRate struct {
	ID          string
	OrgID       string
	Name        string
	// Decimal string, fraction form (e.g. "0.1000" = 10%).
	Rate        string
	IsInclusive bool
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
