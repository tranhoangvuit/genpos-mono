package input

// Customer inputs ----------------------------------------------------------

type CustomerInput struct {
	Name     string
	Email    string
	Phone    string
	Notes    string
	GroupIDs []string
}

type CreateCustomerInput struct {
	OrgID    string
	Customer CustomerInput
}

type UpdateCustomerInput struct {
	ID       string
	OrgID    string
	Customer CustomerInput
}

type DeleteCustomerInput struct {
	ID    string
	OrgID string
}

type GetCustomerInput struct {
	ID    string
	OrgID string
}

// CustomerGroup inputs ------------------------------------------------------

type CustomerGroupInput struct {
	Name          string
	Description   string
	DiscountType  string
	DiscountValue string
}

type CreateCustomerGroupInput struct {
	OrgID string
	Group CustomerGroupInput
}

type UpdateCustomerGroupInput struct {
	ID    string
	OrgID string
	Group CustomerGroupInput
}

type DeleteCustomerGroupInput struct {
	ID    string
	OrgID string
}

// CSV import inputs ---------------------------------------------------------

type ParseImportCustomerCsvInput struct {
	OrgID   string
	CsvData []byte
}

type CsvCustomerRow struct {
	Name       string
	Email      string
	Phone      string
	Notes      string
	Groups     string // comma-separated group names
	Errors     []string
	Exists     bool
	ExistingID string
}

type ParseImportCustomerCsvResult struct {
	Rows       []CsvCustomerRow
	ValidCount int32
	ErrorCount int32
	Warnings   []string
}

type ImportCustomerItem struct {
	Row              CsvCustomerRow
	OverrideExisting bool
	ExistingID       string
}

type ImportCustomersInput struct {
	OrgID string
	Items []ImportCustomerItem
}

type ImportCustomersResult struct {
	Created int32
	Updated int32
	Skipped int32
	Errors  []string
}
