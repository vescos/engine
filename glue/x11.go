// +build linux,!android

package glue

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -lX11 -lXrandr

#include "x11.h"
*/
import "C"
import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/vescos/engine/assets"
	"github.com/vescos/engine/glue/internal/assets/cfd"
	"github.com/vescos/engine/glue/internal/assets/gofd"
	"github.com/vescos/engine/input/keys"
	"github.com/vescos/engine/input/size"
	"github.com/vescos/engine/input/touch"
)

type platform struct {
	cRefs        *C.cRefs
	windowWidth  int
	windowHeight int
	exitMain     bool
}

func init() {
	log.Print(">>>>> Status: Initializing...")
}

func (g *Glue) InitPlatform(s State) {
	g.cRefs = (*C.cRefs)(C.cRefsPtr())
	g.PlatformString = runtime.GOARCH + "/" + runtime.GOOS

	log.Printf(">>>>> Platform: %v", g.PlatformString)

	g.cRefs.xDisplay = C.XOpenDisplay(nil)
	if g.cRefs.xDisplay == nil {
		log.Print("InitPlatform: XOpenDisplay failed")
		g.AppExit(s)
	}

	// Parse flags of type -flag=string
	g.Config = make(map[string]string)
	for _, v := range os.Args[1:] {
		sp := strings.SplitN(v, "=", 2)
		if len(sp) < 2 {
			log.Printf("InitPlatform: Can't parse flag: %v. Use -flag=string", v)
			continue
		}
		// remove leading -
		if sp[0][0] != []byte("-")[0] {
			log.Printf("InitPlatform: Missing '-' in flag definition: %v. Use -flag=string, ignoring", v)
			continue
		}
		key := sp[0][1:]
		g.Config[key] = sp[1]
	}
	// Read config
	fname := filepath.Clean(os.ExpandEnv(g.LinuxConfigFile))
	// Overwrite LinuxConfigFile if available in command line params
	if fn, ok := g.Config["LinuxConfigFile"]; ok {
		fname = filepath.Clean(os.ExpandEnv(fn))
	}
	file, err := os.Open(fname)
	if err != nil {
		if g.LinuxConfigFile != "" {
			log.Printf("InitPlatform: Can't open config file: %v, error: %v", g.LinuxConfigFile, err)
		}
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			t := scanner.Text()
			sp := strings.SplitN(t, "=", 2)
			if len(sp) != 2 {
				log.Print("InitPlatform: Bad Config key: %v", sp)
				continue
			}
			key, val := sp[0], sp[1]
			// cmd params have priority, add key only if not exists
			if _, ok := g.Config[key]; !ok {
				g.Config[key] = val
			}
		}
	}
}

func (g *Glue) StartMainLoop(s State) {
	s.InitState()
	runtime.LockOSThread()
	if g.FbWidth <= 0 {
		g.FbWidth = g.ScreenWidth()
	}
	if g.FbHeight <= 0 {
		g.FbHeight = g.ScreenHeight()
	}
	g.RefreshRate = float32(C.createWindow(C.int(g.FbWidth), C.int(g.FbHeight), g.cRefs))
	g.windowWidth = g.FbWidth
	g.windowHeight = g.FbHeight
	s.Load()
	s.InitGL()
	sz := size.Event{
		WidthPx:     int(g.FbWidth),
		HeightPx:    int(g.FbHeight),
		WidthPt:     float32(g.FbWidth),
		HeightPt:    float32(g.FbHeight),
		PixelsPerPt: 1.0,
	}
	s.Resume()
	g.fpsTime = time.Now()
	g.fpsCount = 0
	s.StartDrawing()
	s.Size(sz)
	for {
		if g.exitMain {
			s.StopDrawing()
			s.Pause()
			s.Destroy()

			// Block and give time to other goroutines to exit.
			time.Sleep(time.Millisecond * 50)
			//time.AfterFunc(time.Millisecond*50, func() { os.Exit(0) })
			C.free(unsafe.Pointer(g.cRefs))
			// Return to main.main or caller.
			return
		}
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
	g.exitMain = true
}

// All keys that don't present in cfg will be deleted from config
// TODO: above is not ok.
func (g *Glue) SaveConfig(cfg map[string]string) bool {
	fname := filepath.Clean(os.ExpandEnv(g.LinuxConfigFile))
	// Overwrite LinuxConfigFile if available in command line params
	if fn, ok := g.Config["LinuxConfigFile"]; ok {
		fname = filepath.Clean(os.ExpandEnv(fn))
	}
	if fname == "" {
		log.Printf("SaveConfig: empty string LinuxConfigFile, skipping, fname:", fname)
		return false
	}
	fdir := filepath.Dir(fname)
	os.MkdirAll(fdir, 0777)
	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	oldCfg := make(map[string]string)
	if err != nil {
		log.Printf("SaveConfig: can't open file: %v, err: %v", fname, err)
		return false
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			t := scanner.Text()
			sp := strings.SplitN(t, "=", 2)
			if len(sp) != 2 {
				log.Print("SaveConfig: Bad Config key: %v", sp)
				continue
			}
			key, val := sp[0], sp[1]
			oldCfg[key] = val
		}
	}
	file.Truncate(0)
	file.Seek(0, 0)
	for k, v := range cfg {
		// check for equal sign(=) or \n into key
		if strings.ContainsAny(k, "=\n") {
			log.Printf("SaveConfig: equal sign(=) and newline is not allowed in config map keys, key: %v, ignoring", k)
			continue
		}
		// check for \n in value
		if strings.Contains(v, "\n") {
			log.Printf("SaveConfig: newline is not allowed in config map vals, val: %v, ignoring", v)
			continue
		}
		_, err = file.WriteString(k + "=" + v + "\n")
		if err != nil {
			log.Print("SaveConfig: can't save config file")
			return false
		}
	}
	for k, v := range oldCfg {
		// write old entries only if not present in cfg
		if _, ok := cfg[k]; ok {
			continue
		}
		if strings.ContainsAny(k, "=\n") {
			log.Printf("SaveConfig: equal sign(=) and newline is not allowed in config map keys, key: %v, ignoring", k)
			continue
		}
		// check for \n in value
		if strings.Contains(v, "\n") {
			log.Printf("SaveConfig: newline is not allowed in config map vals, val: %v, ignoring", v)
			continue
		}
		_, err = file.WriteString(k + "=" + v + "\n")
		if err != nil {
			log.Print("SaveConfig: can't save config file")
			return false
		}
	}
	return true
}

func (g *Glue) CFdHandle(path string) assets.FileManager {
	return &cfd.State{AssetsPath: path}
}

func (g *Glue) GoFdHandle(path string) assets.FileManager {
	return &gofd.State{AssetsPath: path}
}

// Will work after InitPlatform call
// and for screen 0
func (g *Glue) ScreenWidth() int {
	if g.cRefs.xDisplay == nil {
		log.Print("ScreenWidth: xDisplay is nil!!!")
		return 0
	}
	return int(C.XDisplayWidth(g.cRefs.xDisplay, 0))
}

func (g *Glue) ScreenHeight() int {
	if g.cRefs.xDisplay == nil {
		log.Print("ScreenHeight: xDisplay is nil!!!")
		return 0
	}
	return int(C.XDisplayHeight(g.cRefs.xDisplay, 0))
}

func (g *Glue) WindowWidth() int {
	return g.windowWidth
}

func (g *Glue) WindowHeight() int {
	return g.windowHeight
}

// Return reference to C struct.
// Allow access to X11.
// Not crossPlatform code. Use import "C" and
// // +build linux,!android
func (g *Glue) HackPlatform() *C.cRefs {
	return g.cRefs
}

func (g *Glue) processEvents(s State) {
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
			g.windowWidth = g.FbWidth
			g.windowHeight = g.FbHeight
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
