package cursor

type Cursor interface {
	Curr() uint
	Adv() error
}
