// +build linux,!android

package gofd

import (
	"os"
	"path/filepath"

	"github.com/vescos/engine/assets"
)

type State struct {
	AssetsPath string
}

func (s *State) OpenAsset(name string) (assets.Asset, error) {

	//dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//dir = filepath.Join(dir, s.AssetsPath)
	name = filepath.Join(s.AssetsPath, name)
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}
