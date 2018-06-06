// +build android

package gofd

/*
#cgo LDFLAGS: -landroid


#include <android/configuration.h>
#include <android/native_activity.h>

#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/vescos/engine/assets"
)

type State struct {
	AssetManager unsafe.Pointer //*C.AAssetManager
}

/////////////////////////////////////////////////////////////////
// copy/paste from mobile/app with mods
/////////////////////////////////////////////////////////////////
func (s *State) OpenAsset(name string) (assets.Asset, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	a := &asset{
		ptr:  C.AAssetManager_open((*C.AAssetManager)(s.AssetManager), cname, C.AASSET_MODE_UNKNOWN),
		name: name,
	}
	if a.ptr == nil {
		return nil, a.errorf("open", "bad asset")
	}
	return a, nil
}

type asset struct {
	ptr  *C.AAsset
	name string
}

func (a *asset) errorf(op string, format string, v ...interface{}) error {
	return &os.PathError{
		Op:   op,
		Path: a.name,
		Err:  fmt.Errorf(format, v...),
	}
}

func (a *asset) Read(p []byte) (n int, err error) {
	n = int(C.AAsset_read(a.ptr, unsafe.Pointer(&p[0]), C.size_t(len(p))))
	if n == 0 && len(p) > 0 {
		return 0, io.EOF
	}
	if n < 0 {
		return 0, a.errorf("read", "negative bytes: %d", n)
	}
	return n, nil
}

func (a *asset) Seek(offset int64, whence int) (int64, error) {
	off := C.AAsset_seek(a.ptr, C.off_t(offset), C.int(whence))
	if off == -1 {
		return 0, a.errorf("seek", "bad result for offset=%d, whence=%d", offset, whence)
	}
	return int64(off), nil
}

func (a *asset) Close() error {
	C.AAsset_close(a.ptr)
	return nil
}

func (a *asset) Name() string {
	return a.name
}

func (a *asset) Fd() uintptr {
	return uintptr(unsafe.Pointer(a.ptr))
}
