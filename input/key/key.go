package key

type Event struct {
	Code int
	Type Type
	// TODO: modifiers
}

type Type int 

const (
	Down Type = iota
	Up
	Repeat
)
