// +build android

package cfd

// C code from here
//http://www.50ply.com/blog/2013/01/19/loading-compressed-android-assets-with-file-pointer/#comment-1850768990

/*
#cgo LDFLAGS: -landroid


#include <android/configuration.h>
#include <android/native_activity.h>

#include <stdlib.h>
#include <stdio.h>

static inline int android_read(void* cookie, char* buf, int size) {
	return AAsset_read((AAsset*)cookie, buf, size);
}

static inline int android_write(void* cookie, const char* buf, int size) {
	return -1; // can't provide write access to the apk
}

static inline fpos_t android_seek(void* cookie, fpos_t offset, int whence) {
	return AAsset_seek((AAsset*)cookie, offset, whence);
}

static inline int android_close(void* cookie) {
	AAsset_close((AAsset*)cookie);
	return 0;
}

static inline FILE* android_fopen(const char* fname, AAssetManager* am) {

	AAsset* asset = AAssetManager_open(am, fname, AASSET_MODE_UNKNOWN);
	if(!asset) return NULL;

	return funopen(asset, android_read, android_write, android_seek, android_close);
}

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

func (s *State) OpenAsset(name string) (assets.Asset, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	a := &cFile{
		//TODO: handle error beter way
		ptr:  C.android_fopen(cname, (*C.AAssetManager)(s.AssetManager)),
		name: name,
	}

	if a.ptr == nil {
		return nil, a.errorf("open", "bad asset")
	}
	return a, nil
}

type cFile struct {
	ptr  *C.FILE
	name string
}

func (a *cFile) errorf(op string, format string, v ...interface{}) error {
	return &os.PathError{
		Op:   op,
		Path: a.name,
		Err:  fmt.Errorf(format, v...),
	}
}

//TODO: not tested and not clear what next funcs are doing
func (a *cFile) Read(p []byte) (n int, err error) {
	n = int(C.fread(unsafe.Pointer(&p[0]), 1, C.size_t(len(p)), a.ptr))

	if n == 0 && len(p) > 0 {
		return 0, io.EOF
	}
	if n < 0 {
		return 0, a.errorf("read", "negative bytes: %d", n)
	}
	return n, nil
}

func (a *cFile) Seek(offset int64, whence int) (int64, error) {
	off := C.fseek(a.ptr, C.long(offset), C.int(whence))
	if off < 0 {
		return int64(off), a.errorf("seek", "bad result for offset=%d, whence=%d", offset, whence)
	}
	return int64(off), nil
}

func (a *cFile) Close() error {
	C.fclose(a.ptr)
	return nil
}

func (a *cFile) Name() string {
	return a.name
}

func (a *cFile) Fd() uintptr {
	return uintptr(unsafe.Pointer(a.ptr))
}
