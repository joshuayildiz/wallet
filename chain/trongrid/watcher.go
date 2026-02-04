package trongrid

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/joshuayildiz/wallet/cursor"
	"github.com/joshuayildiz/wallet/txevent"
)

type Watcher struct {
	trongrid *Client
	EventCh  chan txevent.E
	ErrCh    chan error
}

func Watch(ctx context.Context, trongrid *Client, c cursor.Cursor, filter func(hash, sender, receiver string) bool) *Watcher {
	self := &Watcher{
		trongrid: trongrid,
		EventCh:  make(chan txevent.E),
		ErrCh:    make(chan error, 1),
	}

	go self.watch(ctx, c, filter)

	return self
}

func (r *Watcher) watch(ctx context.Context, c cursor.Cursor, filter func(hash, sender, receiver string) bool) {
	defer close(r.EventCh)
	defer close(r.ErrCh)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Poll immediately on first iteration, then use ticker
	immediate := make(chan struct{}, 1)
	immediate <- struct{}{}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop

		case <-immediate:
			if err := r.poll(ctx, c, filter); err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					break loop
				}
				r.ErrCh <- err
				break loop
			}

		case <-ticker.C:
			if err := r.poll(ctx, c, filter); err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					break loop
				}
				r.ErrCh <- err
				break loop
			}
		}
	}
}

func (r *Watcher) poll(ctx context.Context, c cursor.Cursor, filter func(hash, sender, receiver string) bool) error {
	now, err := r.trongrid.Now(ctx)
	if err != nil {
		return fmt.Errorf("watcher: fetching now block: %w", err)
	}

	latest := now.BlockHeader.RawData.Number
	if c.Curr() == latest || latest == 0 {
		return nil
	}

	for c.Curr() < latest {
		b, err := r.trongrid.BlockByNum(ctx, c.Curr())
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		} else if err != nil {
			// Non-context error fetching block, skip this poll cycle
			return nil
		}

		if err := r.doBlock(ctx, b, filter); err != nil {
			return fmt.Errorf("watcher: %w", err)
		}

		if err := c.Adv(); err != nil {
			return fmt.Errorf("advancing cursor: %w", err)
		}
	}

	return nil
}

func (r *Watcher) doBlock(ctx context.Context, b *Block, filter func(hash, sender, receiver string) bool) error {
	txInfoList, err := r.trongrid.TxInfoByBlockNum(ctx, b.BlockHeader.RawData.Number)
	if err != nil {
		return err
	}

	txInfoMap := make(map[string]TxInfo, len(txInfoList))
	for _, i := range txInfoList {
		txInfoMap[i.ID] = i
	}

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
			hash := tx.TxID
			from, err := decodeTransferAddr(first.Parameter.Value.OwnerAddress)
			if err != nil {
				return fmt.Errorf("decoding owner address: %w", err)
			}
			to, err := decodeTransferAddr(first.Parameter.Value.ToAddress)
			if err != nil {
				return fmt.Errorf("decoding to address: %w", err)
			}
			amt := first.Parameter.Value.Amount

			if !filter(hash, from, to) {
				continue
			}

			info, ok := txInfoMap[tx.TxID]
			if !ok {
				return fmt.Errorf("tx info not found: %s", tx.TxID)
			}

			r.EventCh <- txevent.E{
				Block:    b.BlockHeader.RawData.Number,
				Currency: txevent.TRX,
				Hash:     hash,
				Sender:   from,
				Receiver: to,
				Amount:   amt,
				Fee:      info.Fee,
			}

		case "TriggerSmartContract":
			info, ok := txInfoMap[tx.TxID]
			if !ok {
				return fmt.Errorf("tx info not found: %s", tx.TxID)
			}

			if info.Receipt.Result != "SUCCESS" {
				continue
			}

			for _, l := range info.Log {
				if l.Address != encodedUSDTContractAddr(r.trongrid.Net) {
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
				from, err := decodeTopicAddr(r.trongrid.Net, l.Topics[1])
				if err != nil {
					return fmt.Errorf("decoding topic sender address: %w", err)
				}
				to, err := decodeTopicAddr(r.trongrid.Net, l.Topics[2])
				if err != nil {
					return fmt.Errorf("decoding topic receiver address: %w", err)
				}
				amt, err := strconv.ParseInt(l.Data, 16, 64)
				if err != nil {
					return fmt.Errorf("parsing transfer amount %q: %w", l.Data, err)
				}

				if !filter(hash, from, to) {
					continue
				}

				r.EventCh <- txevent.E{
					Block:    b.BlockHeader.RawData.Number,
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
