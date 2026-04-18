package input

import "time"

// Supplier inputs ------------------------------------------------------------

type SupplierInput struct {
	Name        string
	ContactName string
	Email       string
	Phone       string
	Address     string
	Notes       string
}

type CreateSupplierInput struct {
	OrgID    string
	Supplier SupplierInput
}

type UpdateSupplierInput struct {
	ID       string
	OrgID    string
	Supplier SupplierInput
}

type DeleteSupplierInput struct {
	ID    string
	OrgID string
}

// Purchase order inputs ------------------------------------------------------

type PurchaseOrderItemInput struct {
	VariantID       string
	QuantityOrdered string
	CostPrice       string
}

type PurchaseOrderInput struct {
	StoreID      string
	SupplierName string
	Notes        string
	ExpectedAt   time.Time
	Items        []PurchaseOrderItemInput
}

type CreatePurchaseOrderInput struct {
	OrgID         string
	UserID        string
	PurchaseOrder PurchaseOrderInput
}

type UpdatePurchaseOrderInput struct {
	ID            string
	OrgID         string
	PurchaseOrder PurchaseOrderInput
}

type GetPurchaseOrderInput struct {
	ID    string
	OrgID string
}

type DeletePurchaseOrderInput struct {
	ID    string
	OrgID string
}

type SubmitPurchaseOrderInput struct {
	ID    string
	OrgID string
}

type CancelPurchaseOrderInput struct {
	ID    string
	OrgID string
}

type ReceivePurchaseOrderLineInput struct {
	ItemID             string
	QuantityToReceive  string
}

type ReceivePurchaseOrderInput struct {
	ID     string
	OrgID  string
	UserID string
	Lines  []ReceivePurchaseOrderLineInput
}
