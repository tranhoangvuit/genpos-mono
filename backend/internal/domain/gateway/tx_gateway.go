package gateway

import "context"

//go:generate mockgen -source=tx_gateway.go -destination=mock/mock_tx_gateway.go -package=mock

// TxManager abstracts database transactions for use cases that need atomicity.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
