//Dec 2017

package draw2d

import (
	//"time"

	"github.com/vescos/engine/draw2d/effects"
	"github.com/vescos/engine/draw2d/effects/ui"
	"github.com/vescos/engine/draw2d/screen"
	"github.com/vescos/engine/geom"
)

const ZINDEX_CNT = 8

type DrawCmd int32

const (
	CmdSendBuff DrawCmd = iota
	CmdWriteBuff
	CmdExit
)

type Buffs struct {
	Ind  []byte
	Vert []byte
	next uint16 //next index - used internaly by engine
}

type Zindex uint8

const (
	Z0 Zindex = iota //top - last draw
	Z1
	Z2
	Z3
	Z4
	Z5
	Z6
	Z7 //bottom - first draw
)

type Handle struct {
	in  chan interface{}
	out chan []*Buffs
}

func (d *Handle) Start() {
	d.in = make(chan interface{}, 100)
	d.out = make(chan []*Buffs)
	go func() {
		poller(d)
	}()
}

func (d *Handle) OnSize(w, h float32) {
	d.in <- &screen.Size{w, h}
}

// Get Buffers - blocks until buffers are ready
func (d *Handle) GetBuffs() []*Buffs {
	d.in <- CmdSendBuff
	return <-d.out
}

// Start writing buffers
func (d *Handle) WriteBuffs() {
	d.in <- CmdWriteBuff
}

func (d *Handle) StartEffect(id string, e effects.Effect, z Zindex, override bool) {
	d.in <- &effect{id: id, e: e, z: z, override: override}
}

func (d *Handle) CtlEffect(id string, ctl interface{}) {
	d.in <- &ctleff{id: id, ctl: ctl}
}

// Delete effect from 2D engine
func (d *Handle) DelEffect(id string) {
	d.in <- &deleff{id: id}
}

func (d *Handle) Stop() {
	d.in <- CmdExit
}

// Effects specific api
// ui effect
func (d *Handle) UiAddElement(effectId, elementId string, x, y, w, r float32,
	m ui.DrawMode, tc [8]float32, size screen.Size) *geom.Rectangle {
	el := ui.Element{X: x, Y: y, W: w, R: r, Mode: m, Tc: tc}
	el.Rectangle = ui.GetRectangle(el, size)
	d.CtlEffect(effectId, &ui.AddElement{elementId, &el})
	// return copy of rectangle struct
	cp := *(el.Rectangle)
	return &cp
}

func (d *Handle) UiDelElement(effectId, elementId string) {
	d.CtlEffect(effectId, &ui.DelElement{elementId})
}

func (d *Handle) UiGetRectangle(x, y, w, r float32, m ui.DrawMode, s screen.Size) *geom.Rectangle {
	return ui.GetRectangle(ui.Element{X: x, Y: y, W: w, R: r, Mode: m}, s)
}
