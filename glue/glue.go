package glue

import (
	//"log"
)

type State interface{
	InitState()
}

type App struct {
	State
	Flags	map[string]string
	
	
	// Unexported fields
	inputBlock      chan int
	commandQueue    chan int
	drawCmdQueue    chan int
	drawTerm        chan int
	drawCmdBlock    chan int
	//windowConfig    chan size.Event
	//inputSendEvents chan touchEvent
	//inputQueue chan *C.AInputQueue
}

func (a *App) Start () {
	
	a.InitState()
	
	a.inputBlock =     make(chan int)
	a.commandQueue=   make(chan int, 100)
	a.drawCmdQueue=    make(chan int, 100)
	a.drawTerm=        make(chan int, 5)
	a.drawCmdBlock=    make(chan int)
	//a.windowConfig=    make(chan size.Event, 100)
	//a.inputSendEvents= make(chan touchEvent, 50)
	//inputQueue = make(chan *C.AInputQueue)
	
	//log.Printf("Glue Start %+v", a)
	
}
