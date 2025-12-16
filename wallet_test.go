package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/joshuayildiz/wallet/chain"
	"github.com/joshuayildiz/wallet/chain/trongrid"
	"github.com/joshuayildiz/wallet/tronusdt"
	"github.com/joshuayildiz/wallet/trx"
	"github.com/stretchr/testify/assert"
)

func TestTRXWalletIsWallet(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trongrid := trongrid.New(chain.Mainnet, "")

	trxW, err := trx.New(trongrid)
	assert.NoError(t, err)
	assert.NotNil(t, trxW)

	var w Wallet = trxW

	balance, err := w.Balance(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), balance)
}

func TestTRONUSDTWalletIsWallet(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trongrid := trongrid.New(chain.Mainnet, "")

	tronusdtW, err := tronusdt.New(trongrid)
	assert.NoError(t, err)
	assert.NotNil(t, tronusdtW)

	var w Wallet = tronusdtW

	balance, err := w.Balance(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), balance)
}
