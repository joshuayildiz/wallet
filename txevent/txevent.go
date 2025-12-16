package txevent

type E struct {
	Block    uint
	Hash     string
	Currency Currency
	Sender   string
	Receiver string
	Amount   int
	Fee      int
}

type Currency string

const (
	TRX       Currency = "TRX"
	TRON_USDT Currency = "TRON_USDT"
)
