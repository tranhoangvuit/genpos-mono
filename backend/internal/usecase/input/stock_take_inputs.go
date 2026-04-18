package input

type CreateStockTakeInput struct {
	OrgID   string
	UserID  string
	StoreID string
	Notes   string
}

type GetStockTakeInput struct {
	ID    string
	OrgID string
}

type StockTakeLineInput struct {
	ItemID     string
	CountedQty string
}

type SaveStockTakeProgressInput struct {
	ID    string
	OrgID string
	Notes string
	Lines []StockTakeLineInput
}

type FinalizeStockTakeInput struct {
	ID     string
	OrgID  string
	UserID string
}

type CancelStockTakeInput struct {
	ID    string
	OrgID string
}

type DeleteStockTakeInput struct {
	ID    string
	OrgID string
}
