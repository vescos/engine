//EGL Golang wrapper(partial and not by specification)
//mainly using these sources
//https://github.com/mortdeus/egles/tree/master/egl
//https://github.com/remogatto/egl

package egl

//#include <EGL/egl.h>
import "C"
import "unsafe"

type (
	Display unsafe.Pointer
	Context unsafe.Pointer
	Surface unsafe.Pointer
	Config  unsafe.Pointer
)

//not by spec
func GetDefaultDisplay() Display {
	return Display(C.eglGetDisplay(C.EGL_DEFAULT_DISPLAY))
}

//not by spec
func InitDisplay(d Display) bool {
	var r C.EGLBoolean
	r = int(C.eglInitialize(C.EGLDisplay(unsafe.Pointer(d))))
	if r == 0 {
		return false
	}
	return true
}
