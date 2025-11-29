package trongrid

import (
	"context"
	"testing"
	"time"

	"github.com/joshuayildiz/wallet/chain"
	"github.com/stretchr/testify/assert"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor := &memCursor{curr: 60128600}

	watcher := Watch(ctx, chain.Testnet, cursor)
	assert.NotNil(t, watcher)

	<-watcher.EventCh
}

type memCursor struct {
	curr uint
}

func (r *memCursor) Curr() uint {
	return r.curr
}

func (r *memCursor) Adv(by uint) error {
	r.curr += by
	return nil
}
