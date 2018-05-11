package assets

import (
	"io"
)

type Asset interface {
	io.ReadSeeker
	io.Closer
	Name() string
	Fd() uintptr
}

type FileManager interface {
	OpenAsset(string) (Asset, error)
}
