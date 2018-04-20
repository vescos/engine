package main

import (
	"log"
	"strings"
	"strconv"
	"time"

	"graphs/engine/input/keys"
	"graphs/engine/input/size"
	"graphs/engine/input/touch"
	"graphs/engine/gles2"
	"graphs/engine/glprog"
	"graphs/engine/glue"
)

const (
	stOESETC1ExtStr string = "GL_OES_compressed_ETC1_RGB8_texture"
)

func main() {
	s := &State{}
	s.Glue = &glue.Glue{}
	// glue call to setup platform dependent code 
	// x11 backend will parse cmd flags
	s.InitPlatform(s)
	// set initial FB size in px
	// on android FB size will be overriden by actual device screen size
	var err error
	if fbWidth, ok := s.Flags["fbWidth"]; ok {
		s.FbWidth, err = strconv.Atoi(fbWidth)
		if err != nil {
			log.Printf("Can't parse fbWidth, Err: %v, fbWidth: %v", err, fbWidth)
			log.Print("FbWidth set to: 0")
		}
	}
	if fbHeight, ok := s.Flags["fbHeight"]; ok {
		s.FbHeight, err = strconv.Atoi(fbHeight)
		if err != nil {
			log.Printf("Can't parse fbHeight, Err: %v, fbHeight: %v", err, fbHeight)
			log.Print("FbHeight set to: 0")
		}
	}
	s.StartMainLoop(s)
}

// State implements glue.Engine interface
type State struct {
	*glue.Glue
	programs map[string]*glprog.Prog
	fpsTime  time.Time
	ETC1Ext  bool
}

func (s *State) InitState() {
	s.fpsTime = time.Now()
}

func (s *State) Load() {}

func (s *State) InitGL() {
	log.Print(">>>>> ", "GLES version: ", gles2.GetString(gles2.VERSION))
	dst := gles2.GetString(gles2.EXTENSIONS)
	if strings.Contains(dst, stOESETC1ExtStr) {
		s.ETC1Ext = true
	}
	log.Print(">>>>> ", stOESETC1ExtStr, ": ", s.ETC1Ext)
}

func (s *State) Size(sz size.Event) {
	gles2.Viewport(0, 0, int(s.FbWidth), int(s.FbHeight))
}

func (s *State) Resume() {}

func (s *State) StartDrawing() {}

func (s *State) Draw() {
	gles2.ClearColor(0.0, 1.0, 0.0, 1.0)
	gles2.ClearDepthf(1.0)
	gles2.Clear(gles2.COLOR_BUFFER_BIT | gles2.DEPTH_BUFFER_BIT)
}

func (s *State) StopDrawing() {}

func (s *State) Pause() {}

func (s *State) Destroy() {}

func (s *State) Touch(t touch.Event) {}

func (s *State) Key(k keys.Event) {
	//log.Print(k)
}
