package entity

import "time"

// Customer represents a customer record.
type Customer struct {
	ID        string
	OrgID     string
	Name      string
	Email     string
	Phone     string
	Notes     string
	GroupIDs  []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CustomerGroup represents a pricing/discount group.
type CustomerGroup struct {
	ID            string
	OrgID         string
	Name          string
	Description   string
	DiscountType  string // "percentage" | "fixed" | ""
	DiscountValue string // decimal string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CustomerGroupMember is the join row linking a customer to a group.
type CustomerGroupMember struct {
	ID         string
	OrgID      string
	GroupID    string
	CustomerID string
	CreatedAt  time.Time
}

// CustomerListItem is the summary row for the customers list page.
type CustomerListItem struct {
	ID         string
	Name       string
	Email      string
	Phone      string
	GroupNames string // comma-separated
}
