package input

// Category inputs -----------------------------------------------------------

type CreateCategoryInput struct {
	OrgID     string
	Name      string
	ParentID  string
	Color     string
	SortOrder int32
}

type UpdateCategoryInput struct {
	ID        string
	OrgID     string
	Name      string
	ParentID  string
	Color     string
	SortOrder int32
}

type DeleteCategoryInput struct {
	ID    string
	OrgID string
}

// Product inputs ------------------------------------------------------------

type OptionInput struct {
	Name   string
	Values []string
}

type VariantInput struct {
	Name         string
	SKU          string
	Barcode      string
	Price        string
	CostPrice    string
	TrackStock   bool
	IsActive     bool
	SortOrder    int32
	OptionValues []string // aligns with OptionInput.Name positionally
	TaxClassID   string   // optional FK to tax_classes; empty = no auto tax
}

type ProductImageInput struct {
	URL       string
	SortOrder int32
}

type ProductInput struct {
	Name        string
	Description string
	CategoryID  string
	IsActive    bool
	SortOrder   int32
	Options     []OptionInput
	Variants    []VariantInput
	Images      []ProductImageInput
}

type CreateProductInput struct {
	OrgID   string
	Product ProductInput
}

type UpdateProductInput struct {
	ID      string
	OrgID   string
	Product ProductInput
}

type GetProductInput struct {
	ID    string
	OrgID string
}

type DeleteProductInput struct {
	ID    string
	OrgID string
}

// CSV import inputs ---------------------------------------------------------

type ParseImportCsvInput struct {
	OrgID   string
	CsvData []byte
}

type CsvVariantRow struct {
	SKU               string
	Barcode           string
	Option1Name       string
	Option1Value      string
	Option2Name       string
	Option2Value      string
	Option3Name       string
	Option3Value      string
	Price             string
	CostPrice         string
	InventoryQuantity string
}

type CsvProductRow struct {
	Name        string
	Description string
	Status      string
	Variants    []CsvVariantRow
	Errors      []string
	Exists      bool
	ExistingID  string
}

type ParseImportCsvResult struct {
	Rows       []CsvProductRow
	ValidCount int32
	ErrorCount int32
	Warnings   []string
}

type ImportProductItem struct {
	Row               CsvProductRow
	OverrideExisting  bool
	ExistingID        string
}

type ImportProductsInput struct {
	OrgID string
	Items []ImportProductItem
}

type ImportProductsResult struct {
	Created int32
	Updated int32
	Skipped int32
	Errors  []string
}
