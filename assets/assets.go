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

type OpenAsset interface {
	OpenAsset(string) (Asset, error)
}
