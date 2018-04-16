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
	Size(new, old size.Event)
	Resume()
	StartDrawing()
	Draw()
	StopDrawing()
	Pause()
	Destroy()
	// Input Events
	Touch(t touch.Event)
	Key(k key.Event)
}

type Glue struct {
	Flags       map[string]string
	FbWidth     int
	FbHeight    int
	RefreshRate float32
	// unexported fields
	platform
}
