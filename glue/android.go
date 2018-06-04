// glue is implemented using mainly this article
// https://developer.nvidia.com/sites/default/files/akamai/mobile/docs/android_lifecycle_app_note.pdf
// and investigating code in android_native_app_glue.c, golang.org/x/mobile

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
	"strings"
	"sync"
	"time"
	"unsafe"

	"graphs/engine/assets"
	"graphs/engine/glue/internal/assets/cfd"
	"graphs/engine/glue/internal/assets/gofd"
	"graphs/engine/glue/internal/callfn"
	"graphs/engine/input/keys"
	"graphs/engine/input/size"
	"graphs/engine/input/touch"
)

// Used to link android onCreate with mainLoop
// linkGlue and linkGlueNextId are global variables readed/written from two threads
// Use mutex.Lock/Unlock on read/write
type linkGlue struct {
	cRefs       *C.cRefs
	comm        chan interface{}
	blockInput  chan int
	blockWindow chan int
	linked      bool
}

var linkGlueMap map[int]*linkGlue
var linkGlueNextId int
var linkGlueMutex sync.Mutex

// armeabi version passed from Makefile(6 or 7)
var goarm string = "0"

type platform struct {
	cRefs  *C.cRefs
	linkId int
	// draw flags
	resumed       bool
	hasFocus      bool
	glInitialized bool
	surfaceReady  bool
	drawing       bool
	callSize      bool
	exitMain      bool
	//window config
	windowConfig size.Event
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
	cmdConfigChange
	cmdInputQueueDestroyed
	cmdUnbindContext
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

	g.Config = make(map[string]string)
	cPkgName := C.getPackageName(g.cRefs.aActivity)
	defer C.free(unsafe.Pointer(cPkgName))
	pkgName := C.GoString(cPkgName)
	rml := len(pkgName) + 1

	cExtras := C.getIntentExtras(g.cRefs.aActivity)
	defer C.free(unsafe.Pointer(cExtras))
	if cExtras != nil {
		extras := C.GoString(cExtras)
		opts := strings.Split(extras, "\n")
		for _, val := range opts {
			kv := strings.SplitN(val, "=", 2)
			if len(kv) != 2 {
				log.Printf("InitPlatform: parseIntent: can't parse element: %v, skipping", val)
				continue
			}
			// remove leading pkgname.
			if len(kv[0]) <= rml || !strings.Contains(kv[0][:rml], pkgName+".") {
				log.Printf("InitPlatform: parseIntent: key don't start with pkgName(dot): %v, skipping", val)
				continue
			}
			g.Config[kv[0][rml:]] = kv[1]
		}
	}

	// Read config from sharedPreferences
	// Not too much useful envvars in android but let expandEnv
	prefs := os.ExpandEnv(g.AndroidConfigFile)
	// Overwrite AndroidConfigFile if available in Intent extras line params
	if fn, ok := g.Config["AndroidConfigFile"]; ok {
		prefs = os.ExpandEnv(fn)
	}
	cprefs := C.CString(prefs)
	defer C.free(unsafe.Pointer(cprefs))
	cCfg := C.getSharedPrefs(g.cRefs.aActivity, cprefs)
	if cCfg != nil {
		defer C.free(unsafe.Pointer(cCfg))
		cfg := C.GoString(cCfg)
		opts := strings.Split(cfg, "\n")
		for _, val := range opts {
			kv := strings.SplitN(val, "=", 2)
			if len(kv) != 2 {
				log.Printf("InitPlatform: parseSharedConf: can't parse element: %v, skipping", val)
				continue
			}
			// intent params have priority
			if _, ok := g.Config[kv[0]]; !ok {
				g.Config[kv[0]] = kv[1]
			}
		}
	} else {
		log.Printf("InitPlatform: can't open shared prefs file: %v", prefs)
	}
}

func (g *Glue) StartMainLoop(s State) {
	// sleep for 10ms if not drawing
	// or block on eglSwapBuffers if drawing
	var maxSleep time.Duration = 10
	s.InitState()
	runtime.LockOSThread()
	g.RefreshRate = float32(C.getRefreshRate(g.cRefs.aActivity))
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
		g.processInputQueue(s)
		// Destroy received - try to unlink and exit main loop
		if g.exitMain {
			linkGlueMutex.Lock()
			if linkGlueMap[g.linkId].linked {
				delete(linkGlueMap, g.linkId)
			}
			linkGlueMutex.Unlock()
			return
		}
		// can we draw now???
		if g.hasFocus && g.resumed && g.surfaceReady && g.windowConfig.WidthPx > 0 && g.windowConfig.HeightPx > 0 {
			//YES
			if !g.drawing {
				g.drawing = true
				s.StartDrawing()
			}
			if g.callSize {
				s.Size(g.windowConfig)
				//log.Printf(">>>>> WindowConfig: %vx%vpt(%vx%vpx) at %vfps, density %v",
				//v.ctWidth, v.ctHeight, v.ctWidth*v.ctPixelsPerPt, v.ctHeight*v.ctPixelsPerPt,
				//v.ctRefreshRate, v.ctPixelsPerPt)
				log.Printf(">>>>> FbSize set to: %vx%vpx", g.FbWidth, g.FbHeight)
				g.callSize = false
			}
			s.Draw()
			// this can realy blocks for some time waiting for vsync
			// TODO: investigate if EGL_CONTEXT_LOST is possible and reinitialize
			// "eglSwapBuffers performs an implicit flush operation on the context"
			if C.eglSwapBuffers(g.cRefs.eglDisplay, g.cRefs.eglSurface) == C.EGL_FALSE {
				log.Printf("eglSwapBuffers failed  - exiting. EGL error: %v", eglGetError())
				g.AppExit(s)
			}
		} else {
			if g.drawing {
				g.drawing = false
				// Not safe to use gl calls here - context can be unbound
				s.StopDrawing()
			}
		}
		if !g.drawing {
			time.Sleep(time.Millisecond * maxSleep)
		}
	}
}

func (g *Glue) AppExit(s State) {
	C.ANativeActivity_finish(g.cRefs.aActivity)
}

func (g *Glue) SaveConfig(cfg map[string]string) bool {
	// TODO:
	prefsName := os.ExpandEnv(g.AndroidConfigFile)
	// Overwrite AndroidConfigFile if available in Intent extras line params
	if fn, ok := g.Config["AndroidConfigFile"]; ok {
		prefsName = os.ExpandEnv(fn)
	}
	if prefsName == "" {
		log.Print("SaveConfig: empty string AndroidConfigFile: skipping.")
		return false
	}
	if len(cfg) == 0 {
		return true
	}
	cprefsName := C.CString(prefsName)
	defer C.free(unsafe.Pointer(cprefsName))
	prefsStr := ""
	for key, val := range cfg {
		// check for equal sign(=) or \n into key
		if strings.ContainsAny(key, "=\n") {
			log.Printf("SaveConfig: equal sign(=) and newline is not allowed in config map keys, key: %v, ignoring", key)
			continue
		}
		// check for \n in value
		if strings.Contains(val, "\n") {
			log.Printf("SaveConfig: newline is not allowed in config map vals, val: %v, ignoring", val)
			continue
		}
		prefsStr += key + "=" + val + "\n"
	}
	if len(prefsStr) > 0 {
		cPrefsStr := C.CString(prefsStr)
		defer C.free(unsafe.Pointer(cPrefsStr))
		C.saveSharedPrefs(g.cRefs.aActivity, cprefsName, cPrefsStr)
	}
	
	return true
}

func (g *Glue) CFdHandle(path string) assets.FileManager {
	return &cfd.State{AssetManager: unsafe.Pointer(g.cRefs.aActivity.assetManager)}
}

func (g *Glue) GoFdHandle(path string) assets.FileManager {
	return &gofd.State{AssetManager: unsafe.Pointer(g.cRefs.aActivity.assetManager)}
}

// can return 0 if w = ACONFIGURATION_SCREEN_WIDTH_DP_ANY
func (g *Glue) ScreenWidth() int {
	dpi := getDpi(g.cRefs.aConfig)
	w := int(C.AConfiguration_getScreenWidthDp(g.cRefs.aConfig))
	return w * dpi
}

// can return 0 if h = ACONFIGURATION_SCREEN_HEIGHT_DP_ANY
func (g *Glue) ScreenHeight() int {
	dpi := getDpi(g.cRefs.aConfig)
	w := int(C.AConfiguration_getScreenHeightDp(g.cRefs.aConfig))
	return w * dpi
}

func (g *Glue) WindowWidth() int {
	if g.cRefs.aWindow == nil {
		return 0
	}
	return int(C.ANativeWindow_getWidth(g.cRefs.aWindow))
}

func (g *Glue) WindowHeight() int {
	if g.cRefs.aWindow == nil {
		return 0
	}
	return int(C.ANativeWindow_getHeight(g.cRefs.aWindow))
}

// Return reference to C struct and give access to
// Activity and other android specific stuff.
// Not cross platform code. eg. use import "C" and
// // +build android
func (g *Glue) HackPlatform() *C.cRefs {
	return g.cRefs
}

func (g *Glue) processEvents(s State) {
	eventCounter := 0
	for eventCounter < maxEvents {
		eventCounter += 1
		select {
		default:
			return
		case ev := <-linkGlueMap[g.linkId].comm:
			switch ev.(type) {
			case *size.Event:
				g.windowConfig = *(ev.(*size.Event))
				g.callSize = true
			case *C.ANativeWindow:
				g.cRefs.aWindow = ev.(*C.ANativeWindow)
				g.bindContext(s)
			case *C.AInputQueue:
				g.processInputQueue(s)
				g.cRefs.aInputQueue = ev.(*C.AInputQueue)
				g.processInputQueue(s)
				linkGlueMap[g.linkId].blockInput <- 1
			case androidCmd:
				cmd := ev.(androidCmd)
				switch cmd {
				case cmdUnbindContext:
					g.surfaceReady = false
					if C.eglMakeCurrent(g.cRefs.eglDisplay, nil, nil, nil) == C.EGL_FALSE {
						log.Printf("Glue: eglMakeCurrent: %s", eglGetError())
						g.AppExit(s)
					} else {
						if C.eglDestroySurface(g.cRefs.eglDisplay, g.cRefs.eglSurface) == C.EGL_FALSE {
							log.Printf("Glue: eglDestroySurface failed: %s", eglGetError())
							g.AppExit(s)
						} else {
							g.cRefs.eglSurface = nil
						}
					}
					linkGlueMap[g.linkId].blockWindow <- 1
					g.cRefs.aWindow = nil
				case cmdInputQueueDestroyed:
					g.processInputQueue(s)
					g.cRefs.aInputQueue = nil
					linkGlueMap[g.linkId].blockInput <- 1
				case cmdConfigChange:
					// Not implemented
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
					g.exitMain = true
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

func (g *Glue) bindContext(s State) {
	if g.cRefs.eglSurface != nil {
		linkGlueMap[g.linkId].blockWindow <- 1
		return
	}
	if err := int(C.createGLContext(g.cRefs)); err > 0 {
		log.Printf("glue.draw.createGLContext: (%s)", eglGetError())
		linkGlueMap[g.linkId].blockWindow <- 1
		g.AppExit(s)
		return
	}
	w := int(C.ANativeWindow_getWidth(g.cRefs.aWindow))
	h := int(C.ANativeWindow_getHeight(g.cRefs.aWindow))
	// limit surface size to s.MaxFbWidth
	if g.MaxFbWidth > 0 && w > g.MaxFbWidth {
		h = (g.MaxFbWidth * h) / w
		w = g.MaxFbWidth
	}
	if g.FbWidth <= 0 {
		g.FbWidth = w
	}
	if g.FbHeight <= 0 {
		g.FbHeight = h
	}
	if err := int(C.bindGLContext(g.cRefs, C.int(g.FbWidth), C.int(g.FbHeight))); err > 0 {
		log.Printf("glue.draw.bindGLContext: (%s)", eglGetError())
		linkGlueMap[g.linkId].blockWindow <- 1
		g.AppExit(s)
		return
	}
	g.surfaceReady = true
	// onInitGL can take long and generate ANR so unblock Java thread before call
	linkGlueMap[g.linkId].blockWindow <- 1
	if !g.glInitialized {
		g.glInitialized = true
		s.InitGL()
		// load gl resources(textures etc)
	}
}

func (g *Glue) processInputQueue(s State) {
	if g.cRefs.aInputQueue == nil {
		return
	}
	var event *C.AInputEvent
	for {
		handled := 0
		if C.AInputQueue_hasEvents(g.cRefs.aInputQueue) < 1 {
			return
		}
		if C.AInputQueue_getEvent(g.cRefs.aInputQueue, &event) < 0 {
			return
		}
		if C.AInputQueue_preDispatchEvent(g.cRefs.aInputQueue, event) == 0 {
			handled = processInputEvent(s, event)
			C.AInputQueue_finishEvent(g.cRefs.aInputQueue, event, C.int(handled))
		} else if C.AConfiguration_getSdkVersion(g.cRefs.aConfig) < 16 {
			C.AInputQueue_finishEvent(g.cRefs.aInputQueue, event, C.int(handled))
		}
	}
}

func processInputEvent(s State, e *C.AInputEvent) int {
	handled := 0
	switch C.AInputEvent_getType(e) {
	case C.AINPUT_EVENT_TYPE_KEY:
		c := int(C.AKeyEvent_getKeyCode(e))
		t := keys.Type(C.AKeyEvent_getAction(e))
		s.Key(keys.Event{
			Code: int(c),
			Type: t,
		})
		handled = 1
	case C.AINPUT_EVENT_TYPE_MOTION:
		// At most one of the events in this batch is an up or down event; get its index and change.
		upDownIndex := C.size_t(C.AMotionEvent_getAction(e)&C.AMOTION_EVENT_ACTION_POINTER_INDEX_MASK) >> C.AMOTION_EVENT_ACTION_POINTER_INDEX_SHIFT
		upDownType := touch.TypeMove
		switch C.AMotionEvent_getAction(e) & C.AMOTION_EVENT_ACTION_MASK {
		case C.AMOTION_EVENT_ACTION_DOWN, C.AMOTION_EVENT_ACTION_POINTER_DOWN:
			upDownType = touch.TypeBegin
		case C.AMOTION_EVENT_ACTION_UP, C.AMOTION_EVENT_ACTION_POINTER_UP:
			upDownType = touch.TypeEnd
		}

		for i, n := C.size_t(0), C.AMotionEvent_getPointerCount(e); i < n; i++ {
			t := touch.TypeMove
			if i == upDownIndex {
				t = upDownType
			}
			ev := touch.Event{
				X:        float32(C.AMotionEvent_getX(e, i)),
				Y:        float32(C.AMotionEvent_getY(e, i)),
				Sequence: touch.Sequence(C.AMotionEvent_getPointerId(e, i)),
				Type:     t,
			}
			s.Touch(ev)
			handled = 1
		}
	default:
		//log.Printf("unknown input event, type=%d", C.AInputEvent_getType(e))
	}
	return handled
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
	linkGlueMap[tid].comm <- cmdConfigChange
}

//export onNativeWindowCreated
func onNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- window
	<-linkGlueMap[tid].blockWindow
}

//export onNativeWindowDestroyed
func onNativeWindowDestroyed(activity *C.ANativeActivity, window *C.ANativeWindow) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdUnbindContext
	<-linkGlueMap[tid].blockWindow
}

//export onNativeWindowRedrawNeeded
func onNativeWindowRedrawNeeded(activity *C.ANativeActivity, window *C.ANativeWindow) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- getWindowConfig(activity, window)
}

//export onInputQueueCreated
func onInputQueueCreated(activity *C.ANativeActivity, queue *C.AInputQueue) {
	if queue != nil {
		tid := getThreadId()
		linkGlueMap[tid].comm <- queue
		<-linkGlueMap[tid].blockInput
	}
}

//export onInputQueueDestroyed
func onInputQueueDestroyed(activity *C.ANativeActivity, queue *C.AInputQueue) {
	tid := getThreadId()
	linkGlueMap[tid].comm <- cmdInputQueueDestroyed
	<-linkGlueMap[tid].blockInput
}

// start main.main thread(goroutine) to listen for events comming from above callbacks
// derived from gomobile
//export callMain
func callMain(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize int, mainPC uintptr) {
	// Set env vars - just a few vars are available on android
	// TODO: try to implement getNextEnv with uintptr + i
	n := C.getNextEnv(0)
	for i := 1; n != nil; i += 1 {
		split := strings.SplitN(C.GoString(n), "=", 2)
		if len(split) == 2 {
			os.Setenv(split[0], split[1])
			n = C.getNextEnv(C.int(i))
		}
	}
	// copy/paste from golang.org/x/mobile/app
	// is this is required to init go runtime before call main???
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
	// TODO: instead of tid use TLS
	tid := getThreadId()
	linkGlueMutex.Lock()
	linkGlueMap[tid] = &linkGlue{
		cRefs:       c,
		comm:        make(chan interface{}, 1000),
		blockInput:  make(chan int),
		blockWindow: make(chan int),
	}
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
// window configuration
/////////////////////////////////////////////////////////////////
type windowConfig struct {
	orientation size.Orientation
	pixelsPerPt float32
}

func getWindowConfig(a *C.ANativeActivity, w *C.ANativeWindow) *size.Event {
	cfg := windowConfigRead(a)
	widthPx := int(C.ANativeWindow_getWidth(w))
	heightPx := int(C.ANativeWindow_getHeight(w))
	return &size.Event{
		WidthPx:     widthPx,
		HeightPx:    heightPx,
		WidthPt:     float32(float32(widthPx) / cfg.pixelsPerPt),
		HeightPt:    float32(float32(heightPx) / cfg.pixelsPerPt),
		PixelsPerPt: cfg.pixelsPerPt,
		Orientation: cfg.orientation,
	}
}

//copy/paste from mobile/app
func windowConfigRead(activity *C.ANativeActivity) windowConfig {
	aconfig := C.AConfiguration_new()
	defer C.AConfiguration_delete(aconfig)
	C.AConfiguration_fromAssetManager(aconfig, activity.assetManager)
	orient := C.AConfiguration_getOrientation(aconfig)
	o := size.OrientationUnknown
	switch orient {
	case C.ACONFIGURATION_ORIENTATION_PORT:
		o = size.OrientationPortrait
	case C.ACONFIGURATION_ORIENTATION_LAND:
		o = size.OrientationLandscape
	}

	return windowConfig{
		orientation: o,
		pixelsPerPt: float32(getDpi(aconfig)) / 72,
	}
}

func getDpi(aConfig *C.AConfiguration) int {
	density := C.AConfiguration_getDensity(aConfig)
	var dpi int
	switch density {
	case C.ACONFIGURATION_DENSITY_DEFAULT:
		dpi = 160
	case C.ACONFIGURATION_DENSITY_LOW,
		C.ACONFIGURATION_DENSITY_MEDIUM,
		213, // C.ACONFIGURATION_DENSITY_TV
		C.ACONFIGURATION_DENSITY_HIGH,
		320, // ACONFIGURATION_DENSITY_XHIGH
		480, // ACONFIGURATION_DENSITY_XXHIGH
		640: // ACONFIGURATION_DENSITY_XXXHIGH
		dpi = int(density)
	case C.ACONFIGURATION_DENSITY_NONE:
		log.Print("android device reports no screen density")
		dpi = 72
	default:
		log.Printf("android device reports unknown density: %d", density)
		// All we can do is guess.
		if density > 0 {
			dpi = int(density)
		} else {
			dpi = 72
		}
	}
	return dpi
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
