//glue is implemented using mainly this article
//https://developer.nvidia.com/sites/default/files/akamai/mobile/docs/android_lifecycle_app_note.pdf
//and investigating code in android_native_app_glue.c, golang.org/x/mobile

// +build android

package glue

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -landroid -llog

#include "android.h"
*/
import "C"
import (
	"bufio"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"graphs/engine/assets/cfd"
	"graphs/engine/assets/gofd"
	"graphs/engine/glue/internal/callfn"
)

// Used to link android onCreate with mainLoop
// linkGlue and linkGlueNextId are global variables readed/written from two threads
// Use mutex.Lock/Unlock on read/write
type linkGlue struct {
	cRefs  *C.cRefs
	comm   chan interface{}
	linked bool
}

var linkGlueMap map[int]*linkGlue
var linkGlueNextId int
var linkGlueMutex sync.Mutex

// armeabi version passed from Makefile(6 or 7)
var goarm string = "0"

type platform struct {
	cRefs  *C.cRefs
	linkId int
	comm   chan interface{}
	// draw flags
	resumed       bool
	hasFocus      bool
	glInitialized bool
	surfaceReady  bool
	drawing       bool
}

type androidCmd int

const (
	cmdStart androidCmd = iota
	cmdResume
	cmdPause
	cmdStop
	cmdDestroy
	cmdFocusOn
	cmdFocusOff
	cmdLowMemory
)

func init() {
	linkGlueMap = make(map[int]*linkGlue)

	// Redirect Stderr and Stdout to logcat
	enablePrinting()
	log.Print(">>>>> Status: Initializing...")
}

func (g *Glue) InitPlatform(s State) {
	// Traverse entire linkGlueMap and link to first not linked entry
	linkGlueMutex.Lock()
	for key, val := range linkGlueMap {
		if !val.linked {
			// TODO: delete this key on onDestroy
			g.linkId = key
			g.cRefs = val.cRefs
			val.linked = true
		}
	}
	linkGlueMutex.Unlock()

	g.PlatformString = runtime.GOARCH + "/" + runtime.GOOS
	log.Printf(">>>>> Platform: %v", g.PlatformString)
	log.Printf(">>>>> SDK version: %v", C.AConfiguration_getSdkVersion(g.cRefs.aConfig))
	if runtime.GOARCH == "arm" {
		log.Printf(">>>>> ARM abi: %v", goarm)
	}
	// TODO: read sharedPreferences and set Flags
}

func (g *Glue) StartMainLoop(s State) {
	s.InitState()
	runtime.LockOSThread()
	s.Load()

	// get/init default display
	err := int(C.getDisplay(g.cRefs))
	if err > 0 {
		log.Printf("glue.draw.getDisplay: (%s)", eglGetError())
		g.AppExit(s)
	} else {
		// setEGLConfig and createGLContext
		err = int(C.setEGLConfig(g.cRefs))
		if err > 0 {
			log.Printf("glue.draw.setEGLConfig: (%s)", eglGetError())
			g.AppExit(s)
		}
	}
	for {
		g.processEvents(s)
	}
}

func (g *Glue) AppExit(s State) {
	C.ANativeActivity_finish(g.cRefs.aActivity)
}

func (g *Glue) CFdHandle(path string) *cfd.State {
	return &cfd.State{AssetManager: unsafe.Pointer(g.cRefs.aActivity.assetManager)}
}

func (g *Glue) GoFdHandle(path string) *gofd.State {
	return &gofd.State{AssetManager: unsafe.Pointer(g.cRefs.aActivity.assetManager)}
}

func (g *Glue) processEvents(s State) {
	eventCounter := 0
	for eventCounter < maxEvents {
		eventCounter += 1
		select {
		default:
			return
		case ev := <-g.comm:
			switch ev.(type) {
			case androidCmd:
				cmd := ev.(androidCmd)
				switch cmd {
				case cmdLowMemory:
					// Not implemented
				case cmdFocusOn:
					g.hasFocus = true
				case cmdFocusOff:
					g.hasFocus = false
				case cmdStart:
					// Not implemented
				case cmdResume:
					s.Resume()
					g.resumed = true
				case cmdPause:
					s.Pause()
					g.resumed = false
				case cmdStop:
					g.glInitialized = false
					if int(C.eglDestroyContext(g.cRefs.eglDisplay, g.cRefs.eglContext)) == 0 {
						log.Printf("Glue: eglDestroyContext: (%s)", eglGetError())
						g.AppExit(s)
					}
				case cmdDestroy:
					s.Destroy()
					delete(linkGlueMap, g.linkId)
					return
				default:
					log.Print("Glue: unknown command:")
				}
			default:
				log.Printf("Glue: unknown event type.")
			}
		}
	}
}

/////////////////////////////////////////////////////////////////
// Android callbacks
/////////////////////////////////////////////////////////////////
//export onDestroy
func onDestroy(activity *C.ANativeActivity) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdDestroy
}

//export onStart
func onStart(activity *C.ANativeActivity) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdStart
}

//export onResume
func onResume(activity *C.ANativeActivity) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdResume
}

//export onPause
func onPause(activity *C.ANativeActivity) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdPause
	// TODO: save prefs
}

//export onStop
func onStop(activity *C.ANativeActivity) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdStop
}

//export onLowMemory
func onLowMemory(activity *C.ANativeActivity) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdLowMemory
}

//export onWindowFocusChanged
func onWindowFocusChanged(activity *C.ANativeActivity, hasFocus C.int) {
	tid := getThreadId()
	if int(hasFocus) > 0 {
		linkGlueMap[tid].comm <- cmdFocusOn
	} else {
		linkGlueMap[tid].comm <- cmdFocusOff
	}
}

//export onSaveInstanceState
func onSaveInstanceState(activity *C.ANativeActivity, outSize *C.size_t) unsafe.Pointer {
	return nil
}

//export onConfigurationChanged
func onConfigurationChanged(activity *C.ANativeActivity) {
	tid := getThreadId()
	C.AConfiguration_fromAssetManager(linkGlueMap[tid].cRefs.aConfig, activity.assetManager)
}

//export onNativeWindowCreated
func onNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {

}

//export onNativeWindowDestroyed
func onNativeWindowDestroyed(activity *C.ANativeActivity, window *C.ANativeWindow) {
}

//export onNativeWindowRedrawNeeded
func onNativeWindowRedrawNeeded(activity *C.ANativeActivity, window *C.ANativeWindow) {
}

//export onInputQueueCreated
func onInputQueueCreated(activity *C.ANativeActivity, queue *C.AInputQueue) {
}

//export onInputQueueDestroyed
func onInputQueueDestroyed(activity *C.ANativeActivity, queue *C.AInputQueue) {
}

// start main.main thread(goroutine) to listen for events comming from above callbacks
// derived from gomobile
//export callMain
func callMain(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize int, mainPC uintptr) {
	// copy/paste from golang.org/x/mobile/app
	// is this is required to init go runtime before call main???
	for _, name := range []string{"TMPDIR", "PATH", "LD_LIBRARY_PATH"} {
		n := C.CString(name)
		os.Setenv(name, C.GoString(C.getenv(n)))
		C.free(unsafe.Pointer(n))
	}
	var curtime C.time_t
	var curtm C.struct_tm
	C.time(&curtime)
	C.localtime_r(&curtime, &curtm)
	tzOffset := int(curtm.tm_gmtoff)
	tz := C.GoString(curtm.tm_zone)
	time.Local = time.FixedZone(tz, tzOffset)

	// TODO: savedState is dropped here - find platform independent method to store prefs
	c := (*C.cRefs)(C.cRefsPtr())
	c.aActivity = activity
	c.aConfig = C.AConfiguration_new()
	C.AConfiguration_fromAssetManager(c.aConfig, c.aActivity.assetManager)
	tid := getThreadId()
	linkGlueMutex.Lock()
	linkGlueMap[tid] = &linkGlue{cRefs: c, comm: make(chan interface{}, 1000)}
	linkGlueMutex.Unlock()
	go func() {
		callfn.CallFn(mainPC)
		C.AConfiguration_delete(c.aConfig)
		C.free(unsafe.Pointer(c))
	}()
}

// Helper/wrapper function to get thread Id
func getThreadId() int {
	return int(C.gettid())
}

/////////////////////////////////////////////////////////////////
// Logging to logcat - printing to stderr, stdout with fmt.print
// will fail
// copy/paste from golang.org/x/mobile/internal/mobileinit
/////////////////////////////////////////////////////////////////
type infoWriter struct{}

func (infoWriter) Write(p []byte) (n int, err error) {
	cstr := C.CString(string(p))
	defer C.free(unsafe.Pointer(cstr))
	tag := C.CString(LogTag)
	defer C.free(unsafe.Pointer(tag))
	C.__android_log_write(C.ANDROID_LOG_INFO, tag, cstr)
	return len(p), nil
}
func lineLog(f *os.File, priority C.int) {
	r := bufio.NewReaderSize(f, LogSize)
	tag := C.CString(LogTag)
	defer C.free(unsafe.Pointer(tag))
	for {
		line, _, err := r.ReadLine()
		str := string(line)
		if err != nil {
			str += " " + err.Error()
		}
		cstr := C.CString(str)
		C.__android_log_write(priority, tag, cstr)
		C.free(unsafe.Pointer(cstr))
		if err != nil {
			break
		}
	}
}
func enablePrinting() {
	log.SetOutput(infoWriter{})
	// android logcat includes all of log.LstdFlags
	log.SetFlags(log.Flags() &^ log.LstdFlags)

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stderr = w
	go lineLog(r, C.ANDROID_LOG_ERROR)

	r, w, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w
	go lineLog(r, C.ANDROID_LOG_INFO)
}
