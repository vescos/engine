package glue

import (
//"log"

//"graphs/engine/ext/size"
)

type State interface {
	InitState()
	Create()
	InitGL()
	Size()
	Resume()
	StartDrawing()
	Draw()
	StopDrawing()
	Pause()
	Destroy()
}

type App struct {
	State
	Flags       map[string]string
	FbWidth     int
	FbHeight    int
	RefreshRate float32

	platform

	//all
	//commandQueue    chan int
	//drawCmdQueue    chan int
	//drawTerm        chan int
	//windowConfig    chan size.Event

}

func (a *App) StartMainLoop() {

	a.InitState()
	//a.inputBlock =     make(chan int)
	//a.drawCmdBlock=    make(chan int)
	//inputQueue = make(chan *C.AInputQueue)

	//a.commandQueue=   make(chan int, 100)
	//a.drawCmdQueue=    make(chan int, 100)
	//a.drawTerm=        make(chan int, 5)
	//a.windowConfig=    make(chan size.Event, 100)

	a.mainLoop()

}
