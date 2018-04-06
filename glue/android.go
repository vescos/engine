//glue is implemented using mainly this article
//https://developer.nvidia.com/sites/default/files/akamai/mobile/docs/android_lifecycle_app_note.pdf
//and investigating code in android_native_app_glue.c, golang.org/x/mobile

// +build android

package glue

type platform struct {
}

func init() {
	// Redirect Stderr and Stdout to logcat
	enablePrinting()
	log.Print(">>>>> Status: Initializing...")
}

/////////////////////////////////////////////////////////////////
// Logging to logcat - printing to stderr, stdout with fmt.print
// will fail
// copy/paste from golang.org/x/mobile/internal/mobileinit
/////////////////////////////////////////////////////////////////
type infoWriter struct{}

func (infoWriter) Write(p []byte) (n int, err error) {
	cstr := C.CString(string(p))
	C.__android_log_write(C.ANDROID_LOG_INFO, LogTag, cstr)
	C.free(unsafe.Pointer(cstr))
	return len(p), nil
}
func lineLog(f *os.File, priority C.int) {
	r := bufio.NewReaderSize(f, LogSize)
	for {
		line, _, err := r.ReadLine()
		str := string(line)
		if err != nil {
			str += " " + err.Error()
		}
		cstr := C.CString(str)
		C.__android_log_write(priority, LogTag, cstr)
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
