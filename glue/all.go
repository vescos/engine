package glue

/*

#include <EGL/egl.h>
*/
import "C"
import (
	"fmt"
)

const (
	LogTag  string = "LabyrinthEngine"
	LogSize int    = 1024
	// max events to process on frame
	maxEvents int = 50
)

/////////////////////////////////////////////////////////////////
// eglError copy/paste from gomobile/app
/////////////////////////////////////////////////////////////////
func eglGetError() string {
	errNum := C.eglGetError()
	switch errNum {
	case C.EGL_SUCCESS:
		return "EGL_SUCCESS"
	case C.EGL_NOT_INITIALIZED:
		return "EGL_NOT_INITIALIZED"
	case C.EGL_BAD_ACCESS:
		return "EGL_BAD_ACCESS"
	case C.EGL_BAD_ALLOC:
		return "EGL_BAD_ALLOC"
	case C.EGL_BAD_ATTRIBUTE:
		return "EGL_BAD_ATTRIBUTE"
	case C.EGL_BAD_CONTEXT:
		return "EGL_BAD_CONTEXT"
	case C.EGL_BAD_CONFIG:
		return "EGL_BAD_CONFIG"
	case C.EGL_BAD_CURRENT_SURFACE:
		return "EGL_BAD_CURRENT_SURFACE"
	case C.EGL_BAD_DISPLAY:
		return "EGL_BAD_DISPLAY"
	case C.EGL_BAD_SURFACE:
		return "EGL_BAD_SURFACE"
	case C.EGL_BAD_MATCH:
		return "EGL_BAD_MATCH"
	case C.EGL_BAD_PARAMETER:
		return "EGL_BAD_PARAMETER"
	case C.EGL_BAD_NATIVE_PIXMAP:
		return "EGL_BAD_NATIVE_PIXMAP"
	case C.EGL_BAD_NATIVE_WINDOW:
		return "EGL_BAD_NATIVE_WINDOW"
	case C.EGL_CONTEXT_LOST:
		return "EGL_CONTEXT_LOST"
	default:
		return fmt.Sprintf("Unknown EGL err: %d", errNum)
	}
}
