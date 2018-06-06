//Veselin Kostov Jan 2018

//draw 2d red cover effect

//TODO: sized polygons
//TODO: Reverse alpha - from alpha to color
//TODO: zero Duration to last till Stop()
//TODO: multiple alpha coeficient calculation formulas

package cover

import (
	//"log"
	"time"

	"github.com/vescos/engine/tobyte"
)

type State struct {
	Duration  time.Duration
	Tc        [8]float32 //texture coordinates
	startTime time.Time
	endTime   time.Time
}

func (s *State) Start() {
	s.startTime = time.Now()
	s.endTime = time.Now().Add(s.Duration)
}

func (s *State) OnSize(w, h float32) {}

func (s *State) Ctl(itf interface{}) {}

func (s *State) OnFrame(ind, vert []byte, next uint16) ([]byte, []byte, uint16, bool) {
	//alpha coefficient (1 - (timeFromStart / Duration))
	alphaC := 1 - (float32(time.Now().Sub(s.startTime).Nanoseconds()) / float32(s.Duration.Nanoseconds()))
	vert = append(vert, tobyte.Float32Le([]float32{
		-1, 1, s.Tc[0], s.Tc[1], alphaC,
		1, 1, s.Tc[2], s.Tc[3], alphaC,
		1, -1, s.Tc[4], s.Tc[5], alphaC,
		-1, -1, s.Tc[6], s.Tc[7], alphaC,
	})...)

	ind = append(ind, tobyte.Uint16Le([]uint16{
		next, next + 1, next + 2, next, next + 2, next + 3,
	})...)

	completed := false
	if s.endTime.Before(time.Now()) {
		completed = true
	}

	return ind, vert, next + 4, completed
}
