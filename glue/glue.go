package glue

import (
	//"log"
	"time"

	"graphs/engine/input/keys"
	"graphs/engine/input/size"
	"graphs/engine/input/touch"
)

type State interface {
	// Lifetime events
	InitState()
	Load()
	InitGL()
	Resume()
	StartDrawing()
	Draw()
	StopDrawing()
	Pause()
	Destroy()
	// Input Events
	Size(size.Event)
	Touch(touch.Event)
	Key(keys.Event)
}

type Glue struct {
	Config         map[string]string
	MaxFbWidth     int
	FbWidth        int
	FbHeight       int
	RefreshRate    float32
	PlatformString string
	///////////////////////
	// platform specific
	LinuxConfigFile string
	// eg. android shared preferences
	AndroidConfigFile string
	//////////////////////
	// unexported fields
	platform
	fpsTime  time.Time
	fpsCount int
}
