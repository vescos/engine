package glue

import (
	//"log"

	"graphs/engine/ext/key"
	"graphs/engine/ext/size"
	"graphs/engine/ext/touch"
)

type State interface {
	// Lifetime events
	InitState()
	Create()
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
	Key(key.Event)
}

type Glue struct {
	Flags       map[string]string
	FbWidth     int
	FbHeight    int
	RefreshRate float32
	PlatformString string
	// unexported fields
	platform
}
