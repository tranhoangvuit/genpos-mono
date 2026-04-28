package input

// ----- Payment methods ------------------------------------------------------

type PaymentMethodInput struct {
	Name      string
	Type      string
	IsActive  bool
	SortOrder int32
}

type CreatePaymentMethodInput struct {
	OrgID  string
	Method PaymentMethodInput
}

type UpdatePaymentMethodInput struct {
	ID     string
	OrgID  string
	Method PaymentMethodInput
}

type DeletePaymentMethodInput struct {
	ID    string
	OrgID string
}

// ----- Tax rates ------------------------------------------------------------

type TaxRateInput struct {
	Name        string
	Rate        string
	IsInclusive bool
	IsDefault   bool
}

type CreateTaxRateInput struct {
	OrgID string
	Rate  TaxRateInput
}

type UpdateTaxRateInput struct {
	ID    string
	OrgID string
	Rate  TaxRateInput
}

type DeleteTaxRateInput struct {
	ID    string
	OrgID string
}

// ----- Tax classes ----------------------------------------------------------

type TaxClassRateInput struct {
	TaxRateID  string
	Sequence   int32
	IsCompound bool
}

type TaxClassInput struct {
	Name        string
	Description string
	IsDefault   bool
	SortOrder   int32
	Rates       []TaxClassRateInput
}

type ListTaxClassesInput struct {
	OrgID string
}

type GetTaxClassInput struct {
	OrgID string
	ID    string
}

type CreateTaxClassInput struct {
	OrgID string
	Class TaxClassInput
}

type UpdateTaxClassInput struct {
	ID    string
	OrgID string
	Class TaxClassInput
}

type DeleteTaxClassInput struct {
	ID    string
	OrgID string
}

// ----- Members --------------------------------------------------------------

type CreateMemberInput struct {
	OrgID     string
	Name      string
	Email     string
	Phone     string
	RoleID    string
	Password  string
	AllStores bool
	StoreIDs  []string
}

type UpdateMemberInput struct {
	ID        string
	OrgID     string
	Name      string
	Phone     string
	RoleID    string
	Status    string
	AllStores bool
	StoreIDs  []string
}

type DeleteMemberInput struct {
	ID            string
	OrgID         string
	CurrentUserID string
}
