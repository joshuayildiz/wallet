package wallet

import "context"

type Wallet interface {
	// Hex encoded private key.
	PrivKeyHex() string

	// Human readable wallet address.
	Addr() string

	// Balance of wallet.
	Balance(ctx context.Context) (uint, error)

	// Returns the transaction hash
	Send(ctx context.Context, to string, amt uint) (string, error)
}
