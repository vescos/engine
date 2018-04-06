// +build linux,!android

package glue

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -lX11 -lXrandr

#include <EGL/egl.h>

short createWindow(int w, int h, Display *xDpy, EGLDisplay eDpy, EGLSurface eSurf);
void processEvents(void);
void swapBuffers(void);
*/
import "C"
import (
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type platform struct {
	xDpy  *C.Display
	eDpy  C.EGLDisplay
	eSurf C.EGLSurface
}

func init() {
	log.Print(">>>>> Status: Initializing...")
}

func (a *App) InitPlatform() {
	log.Printf(">>>>> Platform: %v/%v", runtime.GOARCH, runtime.GOOS)

	// Parse flags of type -flag=string
	a.Flags = make(map[string]string)
	for _, v := range os.Args[1:] {
		sp := strings.SplitN(v, "=", 2)
		if len(sp) < 2 {
			log.Printf("Can't parse flag: %v. Use -flag=string", v)
			continue
		}
		// remove leading -
		if sp[0][0] != []byte("-")[0] {
			log.Printf("Missing '-' in flag definition: %v. Use -flag=string", v)
			continue
		}
		key := sp[0][1:]
		a.Flags[key] = sp[1]
	}
}

func (a *App) mainLoop() {
	runtime.LockOSThread()
	a.Create()
	a.RefreshRate = float32(C.createWindow(C.int(a.FbWidth), C.int(a.FbHeight), a.xDpy, a.eDpy, a.eSurf))
	for {
		time.Sleep(time.Millisecond * 10)
	}
}

func (a *App) appExit() {
	time.AfterFunc(time.Millisecond*50, func() { os.Exit(0) })
}

//export JustTest
func JustTest() {

}
