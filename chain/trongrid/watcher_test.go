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
	trongrid := New(chain.Testnet, "")

	watcher := Watch(ctx, trongrid, cursor, func(hash, sender, receiver string) bool {
		return true
	})
	assert.NotNil(t, watcher)

	<-watcher.EventCh
}

type memCursor struct {
	curr uint
}

func (r *memCursor) Curr() uint {
	return r.curr
}

func (r *memCursor) Adv() error {
	r.curr++
	return nil
}
