package cursor

type Cursor interface {
	Curr() uint
	Adv(by uint) error
}
