// +build linux,!android
package key

const (
	Down Type = iota
	Up
	Repeat
)

const (
	CodeEnter  int = 36  //enter
	CodeLeft   = 113 //left arrow
	CodeRight  = 114 //right arrow
	CodeUp     = 111 //up arrow
	CodeDown   = 116 //down arrow
	CodeBack  = 22  //backspace
	
	//this is haedcoded against labyrinth game
	CodeMove   = 38  //a
)
