package input

// ListProductsInput carries validated parameters from the handler to the service.
type ListProductsInput struct {
	OrgID    string
	PageSize int32
	Offset   int32
}
