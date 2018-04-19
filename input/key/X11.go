// +build linux,!android
package key

const (
	Press Type = iota
	Release
	Repeat
)

const (
	CodeEnter int = 36  //enter
	CodeLeft      = 113 //left arrow
	CodeRight     = 114 //right arrow
	CodeUp        = 111 //up arrow
	CodeDown      = 116 //down arrow
	CodeBack      = 22  //backspace

	//this is hardcoded against labyrinth game
	CodeMove = 38 //a
)
