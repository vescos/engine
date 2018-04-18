// +build android

package key

/*

#include <android/keycodes.h>
*/
import "C"

const (
	Down   Type = C.AKEY_EVENT_ACTION_DOWN
	Up          = C.AKEY_EVENT_ACTION_UP
	Repeat      = C.AKEY_EVENT_ACTION_MULTIPLE
)

const (
	CodeEnter int = C.AKEYCODE_ENTER
	CodeLeft   = C.AKEYCODE_DPAD_LEFT
	CodeRight  = C.AKEYCODE_DPAD_RIGHT
	CodeUp     = C.AKEYCODE_DPAD_UP
	CodeDown   = C.AKEYCODE_DPAD_DOWN
	CodeBack  = C.AKEYCODE_DEL
	
	//this is haedcoded against labyrinth game
	Code  = C.AKEYCODE_DPAD_CENTER
)
