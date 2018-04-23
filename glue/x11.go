// +build linux,!android

package glue

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -lX11 -lXrandr

#include "x11.h"
*/
import "C"
import (
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"graphs/engine/assets/cfd"
	"graphs/engine/assets/gofd"
	"graphs/engine/input/keys"
	"graphs/engine/input/size"
	"graphs/engine/input/touch"
)

type platform struct {
	cRefs *C.cRefs
}

func init() {
	log.Print(">>>>> Status: Initializing...")
}

func (g *Glue) InitPlatform(s State) {
	g.cRefs = (*C.cRefs)(C.cRefsPtr())
	g.PlatformString = runtime.GOARCH + "/" + runtime.GOOS

	log.Printf(">>>>> Platform: %v", g.PlatformString)

	g.FbWidth = 1280
	g.FbHeight = 720
	// Parse flags of type -flag=string
	g.Flags = make(map[string]string)
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
		g.Flags[key] = sp[1]
	}
}

func (g *Glue) StartMainLoop(s State) {
	s.InitState()
	runtime.LockOSThread()
	s.Load()
	g.RefreshRate = float32(C.createWindow(C.int(g.FbWidth), C.int(g.FbHeight), g.cRefs))
	s.InitGL()
	sz := size.Event{
		WidthPx:     int(g.FbWidth),
		HeightPx:    int(g.FbHeight),
		WidthPt:     float32(g.FbWidth),
		HeightPt:    float32(g.FbHeight),
		PixelsPerPt: 1.0,
	}
	s.Resume()
	s.StartDrawing()
	s.Size(sz)
	for {
		g.processEvents(s)
		s.Draw()
		// Swap buffers will block till vsync
		if C.eglSwapBuffers(g.cRefs.eglDisplay, g.cRefs.eglSurface) == C.EGL_FALSE {
			log.Printf("eglSwapBuffers failed  - exiting. EGL error: %v", eglGetError())
			g.AppExit(s)
		}
	}
}

func (g *Glue) AppExit(s State) {
	s.StopDrawing()
	s.Pause()
	s.Destroy()
	C.free(unsafe.Pointer(g.cRefs))
	os.Exit(0)
	//time.AfterFunc(time.Millisecond*50, func() { os.Exit(0) })
}

func (g *Glue) CFdHandle(path string) *cfd.State {
	return &cfd.State{AssetsPath: path}
}

func (g *Glue) GoFdHandle(path string) *gofd.State {
	return &gofd.State{AssetsPath: path}
}

func (g *Glue) processEvents(s State) {
	maxEvents := 50
	eventCounter := 0
	for C.XPending(g.cRefs.xDisplay) > 0 && eventCounter < maxEvents {
		eventCounter += 1
		var event C.XEvent
		C.XNextEvent(g.cRefs.xDisplay, &event)

		anyEvent := (*C.XAnyEvent)(unsafe.Pointer(&event))
		switch anyEvent._type {
		case C.ClientMessage:
			cmEvent := (*C.XClientMessageEvent)(unsafe.Pointer(&event))
			data := uint32(cmEvent.data[0]) | (uint32(cmEvent.data[1]) << 8) |
				(uint32(cmEvent.data[2]) << 16) | (uint32(cmEvent.data[3]) << 24)
			if g.cRefs.wmDeleteWindow != C.None && C.ulong(data) == g.cRefs.wmDeleteWindow {
				g.AppExit(s)
				return
			}
		case C.ButtonPress:
			bpEvent := (*C.XButtonPressedEvent)(unsafe.Pointer(&event))
			s.Touch(touch.Event{
				X:        float32(bpEvent.x),
				Y:        float32(bpEvent.y),
				Sequence: touch.Sequence(bpEvent.button),
				Type:     touch.TypeBegin,
			})
		case C.ButtonRelease:
			brEvent := (*C.XButtonReleasedEvent)(unsafe.Pointer(&event))
			s.Touch(touch.Event{
				X:        float32(brEvent.x),
				Y:        float32(brEvent.y),
				Sequence: touch.Sequence(brEvent.button),
				Type:     touch.TypeEnd,
			})
		case C.MotionNotify:
			mnEvent := (*C.XPointerMovedEvent)(unsafe.Pointer(&event))
			// Move apply to all pressed buttons
			// Call Touch for every pressed button
			for i := 1; i < 6; i += 1 {
				if ((1 << uint(7+i)) & int(mnEvent.state)) > 0 {
					s.Touch(touch.Event{
						X:        float32(mnEvent.x),
						Y:        float32(mnEvent.y),
						Sequence: touch.Sequence(i),
						Type:     touch.TypeMove,
					})
				}
			}
		case C.ConfigureNotify:
			cnEvent := (*C.XConfigureEvent)(unsafe.Pointer(&event))
			sz := size.Event{
				WidthPx:     int(cnEvent.width),
				HeightPx:    int(cnEvent.height),
				WidthPt:     float32(cnEvent.width),
				HeightPt:    float32(cnEvent.height),
				PixelsPerPt: 1.0,
			}
			g.FbWidth = int(cnEvent.width)
			g.FbHeight = int(cnEvent.height)
			s.Size(sz)
			log.Printf(">>>>> WindowConfig: %vx%vpx at %vfps, density %v",
				sz.WidthPx, sz.HeightPx, g.RefreshRate, sz.PixelsPerPt)
		case C.KeyPress:
			kpEvent := (*C.XKeyPressedEvent)(unsafe.Pointer(&event))
			s.Key(keys.Event{
				Code: int(kpEvent.keycode),
				Type: keys.Press,
			})
		case C.KeyRelease:
			krEvent := (*C.XKeyPressedEvent)(unsafe.Pointer(&event))
			//check for repeat
			if int(C.XEventsQueued(g.cRefs.xDisplay, C.QueuedAfterReading)) > 0 {
				var nev C.XEvent
				C.XPeekEvent(g.cRefs.xDisplay, &nev)
				anyNev := (*C.XAnyEvent)(unsafe.Pointer(&nev))
				if anyNev._type == C.KeyPress {
					kpNev := (*C.XKeyPressedEvent)(unsafe.Pointer(&nev))
					if kpNev.time == krEvent.time && kpNev.keycode == krEvent.keycode {
						//Clear next event
						C.XNextEvent(g.cRefs.xDisplay, &event)
						s.Key(keys.Event{
							Code: int(krEvent.keycode),
							Type: keys.Repeat,
						})
						break
					}
				}
			}
			s.Key(keys.Event{
				Code: int(krEvent.keycode),
				Type: keys.Release,
			})
			// End Switch
		}
	}
}