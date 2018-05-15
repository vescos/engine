// +build linux,!android

package cfd

/*

#include <stdlib.h>
#include <stdio.h>
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"

	"graphs/engine/assets"
)

type State struct {
	AssetsPath string
}

func (s *State) OpenAsset(name string) (assets.Asset, error) {

	//dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//dir = filepath.Join(dir, s.AssetsPath)
	name = filepath.Join(s.AssetsPath, name)
	cname := C.CString(name)
	cmode := C.CString("rb")
	defer C.free(unsafe.Pointer(cname))
	a := &cFile{
		//TODO: handle error beter way
		ptr:  C.fopen(cname, cmode),
		name: name,
	}

	if a.ptr == nil {
		return nil, a.errorf("open", "can't fopen file")
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
