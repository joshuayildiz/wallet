package trongrid

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/joshuayildiz/wallet/chain"
	"github.com/joshuayildiz/wallet/cursor"
	"github.com/joshuayildiz/wallet/txevent"
)

type Watcher struct {
	net      chain.Network
	trongrid *Client
	EventCh  chan txevent.E
}

func Watch(ctx context.Context, net chain.Network, c cursor.Cursor) *Watcher {
	self := &Watcher{
		net:      net,
		trongrid: New(net),
		EventCh:  make(chan txevent.E),
	}

	go self.watch(ctx, c)

	return self
}

func (r *Watcher) watch(ctx context.Context, c cursor.Cursor) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop

		case <-time.After(3 * time.Second):
			now, err := r.trongrid.Now(ctx)
			if errors.Is(err, context.DeadlineExceeded) {
				break loop
			} else if err != nil {
				panic(fmt.Errorf("watcher: fetching now block: %w", err))
			}

			latest := now.BlockHeader.RawData.Number
			if c.Curr() == latest || latest == 0 {
				continue
			}

			diff := latest - c.Curr()
			if diff > 10 {
				diff = 5
			}

			err = r.do(ctx, c.Curr(), c.Curr()+diff)
			if errors.Is(err, context.DeadlineExceeded) {
				break loop
			} else if err != nil {
				panic(fmt.Errorf("watcher: %w", err))
			}

			err = c.Adv(diff)
			if err != nil {
				panic(fmt.Errorf("advancing cursor: %w", err))
			}
		}
	}

	close(r.EventCh)
}

func (r *Watcher) do(ctx context.Context, from, to uint) error {
	for i := from; i <= to; i++ {
		b, err := r.trongrid.BlockByNum(ctx, i)
		if err != nil {
			return fmt.Errorf("fetching block %d: %w", i, err)
		}

		err = r.doBlock(ctx, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Watcher) doBlock(ctx context.Context, b *Block) error {
	for _, tx := range b.Transactions {
		if len(tx.RawData.Contract) == 0 {
			continue
		}

		// first contract type determines the transaction type
		// TransferContract     : trx transfer
		// TriggerSmartContract : may be trx, trc10 or trc20 (includes usdt) transfer
		first := tx.RawData.Contract[0]
		switch first.Type {
		case "TransferContract":
			info, err := r.trongrid.TxInfoByID(ctx, tx.TxID)
			if err != nil {
				return err // todo: should just retry a couple seconds later
			}

			hash := tx.TxID
			from := decodeTransferAddr(first.Parameter.Value.OwnerAddress)
			to := decodeTransferAddr(first.Parameter.Value.ToAddress)
			amt := first.Parameter.Value.Amount
			r.EventCh <- txevent.E{
				Currency: txevent.TRX,
				Hash:     hash,
				Sender:   from,
				Receiver: to,
				Amount:   amt,
				Fee:      info.Fee,
			}

		case "TriggerSmartContract":
			info, err := r.trongrid.TxInfoByID(ctx, tx.TxID)
			if err != nil {
				return err // todo: should just retry a couple seconds later
			}

			if info.Receipt.Result != "SUCCESS" {
				continue
			}

			for _, l := range info.Log {
				if l.Address != encodedUSDTContractAddr(r.net) {
					continue
				}

				if len(l.Topics) != 3 {
					continue
				}

				encodedEvent := l.Topics[0]
				if encodedEvent != encodedTransferEvent {
					continue
				}

				hash := tx.TxID
				from := decodeTopicAddr(r.net, l.Topics[1])
				to := decodeTopicAddr(r.net, l.Topics[2])
				amt, _ := strconv.ParseInt(l.Data, 16, 64)

				r.EventCh <- txevent.E{
					Currency: txevent.TRON_USDT,
					Hash:     hash,
					Sender:   from,
					Receiver: to,
					Amount:   int(amt),
					Fee:      info.Fee,
				}
			}
		}
	}

	return nil
}
