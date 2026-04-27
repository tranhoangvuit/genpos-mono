package entity

import "time"

// OrderSummary is the list row for the daily sales report.
type OrderSummary struct {
	ID            string
	OrderNumber   string
	Status        string // open|completed|voided|refunded|partially_refunded
	Subtotal      string // decimal string
	TaxTotal      string
	DiscountTotal string
	Total         string
	StoreID       string
	StoreName     string
	RegisterID    string
	UserID        string
	UserName      string
	CustomerID    string
	CustomerName  string
	CreatedAt     time.Time
	Source        string // pos|online_store|shopify|woocommerce|manual|import
	ExternalID    string
}

// Order is the detail view with nested items + payments.
type Order struct {
	ID            string
	OrderNumber   string
	Status        string
	Subtotal      string
	TaxTotal      string
	DiscountTotal string
	Total         string
	Notes         string
	StoreID       string
	StoreName     string
	RegisterID    string
	UserID        string
	UserName      string
	CustomerID    string
	CustomerName  string
	CreatedAt     time.Time
	CompletedAt   time.Time
	Source        string
	ExternalID    string
	LineItems     []*OrderLineItem
	Payments      []*OrderPayment
}

type OrderLineItem struct {
	ID             string
	VariantID      string
	ProductName    string
	VariantName    string
	SKU            string
	Quantity       string
	UnitPrice      string
	TaxRate        string
	TaxAmount      string
	DiscountAmount string
	LineTotal      string
	Notes          string
}

type OrderPayment struct {
	ID                string
	PaymentMethodID   string
	PaymentMethodName string
	Amount            string
	Tendered          string
	ChangeAmount      string
	Reference         string
	Status            string
	CreatedAt         time.Time
}
