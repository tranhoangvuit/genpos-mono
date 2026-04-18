package input

type StoreInput struct {
	Name     string
	Address  string
	Phone    string
	Email    string
	Timezone string
	Status   string
}

type CreateStoreInput struct {
	OrgID string
	Store StoreInput
}

type UpdateStoreInput struct {
	ID    string
	OrgID string
	Store StoreInput
}

type DeleteStoreInput struct {
	ID    string
	OrgID string
}
