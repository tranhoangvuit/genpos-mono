package entity

import "time"

// TaxClass is a named bundle of tax rates assignable to product variants.
// Editing a class re-rates every variant in it without touching variants.
type TaxClass struct {
	ID          string
	OrgID       string
	Name        string
	Description string
	IsDefault   bool
	SortOrder   int32
	Rates       []*TaxClassRate
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaxClassRate is one rate inside a class. Sequence drives application
// order at sale time; IsCompound governs whether this rate's base is
// (taxable_base + previously applied taxes) or just taxable_base.
type TaxClassRate struct {
	ID         string
	TaxRateID  string
	Sequence   int32
	IsCompound bool
}
