// Veselin Kostov Jun, 2017
//
// Part of "Labyrinth lost gems" game engine android/linux version.
// 2D engine

// Create buffers for 2d draws
// Only rectangles are supported
// TODO: solid colors instead of texture atlass

// TODO: BUGFIX: pass Size as parameter to start() to avoid crash when stop-start is triggered withowt onSize happens
// TODO: pass initFunction pointer to be called when started

package draw2d

import (
	"log"

	"graphs/engine/draw2d/effects"
	"graphs/engine/draw2d/screen"
)

const (
	// initial buff/ind size
	buffSize = 80 * 60 // 60 rectangles
	indSize  = 12 * 60
)

type effect struct {
	id string
	e  effects.Effect
	z  Zindex
	// override or skip if effect with same id exists.
	override bool
}

type ctleff struct {
	id  string
	ctl interface{}
}

type deleff struct {
	id string
}

func poller(draw *Handle) {
	log.Print(">>>>> draw2d: Start.")
	var (
		size *screen.Size
	)

	buff := make([]*Buffs, ZINDEX_CNT, ZINDEX_CNT)
	effects := make(map[string]*effect)
	// TODO: Optimisation - use front and back buffers to avoid allocate buffers
	// on every frame but careful to avoid DATA_RACE with main
	for i := 0; i < ZINDEX_CNT; i += 1 {
		buff[i] = &Buffs{
			Ind:  make([]byte, 0, indSize),
			Vert: make([]byte, 0, buffSize),
			next: 0,
		}
	}
	for e := range draw.in {
		switch e.(type) {
		case DrawCmd:
			command := e.(DrawCmd)
			switch command {
			case CmdExit:
				log.Print("<<<<< draw2d: Stop.")
				return
			case CmdWriteBuff:
				// start generating new buffers for next frame.
				buff = make([]*Buffs, ZINDEX_CNT, ZINDEX_CNT)
				for i := 0; i < ZINDEX_CNT; i += 1 {
					buff[i] = &Buffs{
						Ind:  make([]byte, 0, indSize),
						Vert: make([]byte, 0, buffSize),
						next: 0,
					}
				}
				// effects
				for key, value := range effects {
					var r bool
					buff[value.z].Ind, buff[value.z].Vert, buff[value.z].next, r = value.e.OnFrame(buff[value.z].Ind, buff[value.z].Vert, buff[value.z].next)
					if r {
						delete(effects, key)
					}
				}
			case CmdSendBuff:
				// Send 2D buffers
				draw.out <- buff
			default:
				log.Printf("draw2d: Unknown command: %v", command)
			}
		case *screen.Size:
			size = e.(*screen.Size)
			// effects
			for _, value := range effects {
				value.e.OnSize(size.W, size.H)
			}
		case *effect:
			eff := e.(*effect)
			if _, ok := effects[eff.id]; ok && !eff.override {
				log.Printf("draw2d: StartEffect: this id is allredy in use - skipping. ID: %v", eff.id)
			} else {
				effects[eff.id] = eff
				eff.e.Start()
			}
		case *ctleff:
			ctl := e.(*ctleff)
			if _, ok := effects[ctl.id]; !ok {
				log.Printf("draw2d: CtlEffect: can't find id. ID: %v", ctl.id)
			} else {
				effects[ctl.id].e.Ctl(ctl.ctl)
			}
		case *deleff:
			ctl := e.(*deleff)
			delete(effects, ctl.id)
		default:
			log.Printf("draw2d: unsuported type %T", e)
		}
	}
}
