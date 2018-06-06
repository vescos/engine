// Veselin Kostov Jan 2018

// draw 2D snow effect - INCOMPLETE

package snow

import (
	"log"
	"math/rand"
	"time"

	"github.com/vescos/engine/tobyte"
)

type State struct {
	Particles     uint16
	Tc            [][8]float32 // texture coordinates
	StartDuration time.Duration
	Paused        bool
	AspectRatio   float32
	Xr            float32
	Yr            float32
	Persp         float32

	parts     []part
	startTime time.Time
	rand      *rand.Rand
	iSizeX    float32
	iSizeY    float32

	dx   float32
	dy   float32
	move bool

	lp           [3]layerProps
	nextLayer    layer
	currLayerNum uint
}

type layer uint8

const (
	frontLayer layer = iota
	midleLayer
	backLayer
)

type layerProps struct {
	sizeCoef  float32
	fallSpeed float32
	dx        float32
	dist      uint // layer distribution in units eg 3:2:1 between layers
}

const (
	fallSpeed float32 = 0.005
	xMax      float32 = 1.1
	xMax2     float32 = xMax * 2
	yMax      float32 = 1.1
	yMax2     float32 = yMax * 2
	weightCf  float32 = 1.0 / 1.0
	driftCf   float32 = 1.0 / 2.0
)

type Change struct {
	Xr   float32
	Yr   float32
	Move bool
}

type part struct {
	t            int // texture index
	x, y         float32
	sizeX, sizeY float32
	dx, dy       float32
	alpha        float32
	layer        layer
	weight       float32 // small random dy add
	drift        float32 // small random dx add
}

func (s *State) Start() {
	s.startTime = time.Now()
	s.parts = make([]part, 0, s.Particles)
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.iSizeX = 0.025
	s.iSizeY = s.iSizeX * s.AspectRatio
	// lauers props
	// HARDCODED for now
	s.lp[frontLayer] = layerProps{
		sizeCoef:  1.6,
		fallSpeed: 1.5 * fallSpeed,
		dx:        0.001,
		dist:      1,
	}
	s.lp[midleLayer] = layerProps{
		sizeCoef:  1.3,
		fallSpeed: 1.1 * fallSpeed,
		dx:        -0.0005,
		dist:      2,
	}
	s.lp[backLayer] = layerProps{
		sizeCoef:  1,
		fallSpeed: 0.8 * fallSpeed,
		dx:        0.0002,
		dist:      3,
	}

}

func (s *State) OnSize(w, h float32) {
	s.AspectRatio = w / h
	s.iSizeY = s.iSizeX * s.AspectRatio
	//TODO: update parts sizeX, sizeY values
}

func (s *State) Ctl(itf interface{}) {
	switch itf.(type) {
	case *Change:
		c := itf.(*Change)
		s.dy = ((s.Xr - c.Xr) / s.Persp) * 2
		s.Xr = c.Xr
		s.dx = ((s.Yr - c.Yr) / (s.Persp * s.AspectRatio)) * 2
		s.Yr = c.Yr
		s.move = c.Move
	default:
		log.Printf("Effects: snow: Unsupported interface: %T", itf)
	}
}

func (s *State) OnFrame(ind, vert []byte, next uint16) ([]byte, []byte, uint16, bool) {
	if len(s.parts) < int(s.Particles) {
		tp := int((float32(time.Now().Sub(s.startTime).Nanoseconds()) * float32(s.Particles)) /
			float32(s.StartDuration.Nanoseconds()))
		for i := len(s.parts); i < tp && len(s.parts) < int(s.Particles); i += 1 {
			// choose random texture for this part
			t := s.rand.Intn(len(s.Tc))
			// x from -1.1 to 1.1
			x := (s.rand.Float32() * xMax2) - xMax
			// y from 1.0 to 1.1
			y := (s.rand.Float32() * (yMax - 1.0)) + 1.0

			// distribute parts between layers
			lr := layer(s.nextLayer)
			s.currLayerNum += 1
			if s.currLayerNum >= s.lp[lr].dist {
				s.nextLayer += 1
				if s.nextLayer > backLayer {
					s.nextLayer = frontLayer
				}
				s.currLayerNum = 0
			}
			// for 0 to weightCf part of layer fallSpeed
			weight := s.lp[lr].fallSpeed * s.rand.Float32() * weightCf
			// frand(0 to  +- layer dx / driftCf)
			var sign float32 = 1.0
			if s.rand.Float32() > 0.5 {
				sign = -1.0
			}
			drift := s.lp[lr].dx * s.rand.Float32() * driftCf * sign
			sizeX := s.iSizeX * s.lp[lr].sizeCoef
			sizeY := sizeX * s.AspectRatio
			s.parts = append(s.parts, part{t: t, x: x, y: y, sizeX: sizeX, sizeY: sizeY, layer: lr, weight: weight, drift: drift})
		}
	}

	for key, val := range s.parts {
		// animate particles
		lr := val.layer

		// TODO: merge multiple asigments into one
		val.y -= s.lp[lr].fallSpeed
		val.y -= s.dy
		val.y -= val.weight

		val.x += s.lp[lr].dx
		val.x += s.dx
		val.x += val.drift

		if val.y > yMax {
			val.y -= yMax2
		}
		if val.y < -yMax-val.sizeY {
			val.y += yMax2 - val.sizeY
		}
		if val.x < -xMax {
			val.x += xMax2
		}
		if val.x > xMax+val.sizeX {
			val.x -= xMax2 - val.sizeX
		}

		v := tobyte.Float32Le([]float32{
			val.x, val.y, s.Tc[val.t][0], s.Tc[val.t][1], 0.8,
			val.x + val.sizeX, val.y, s.Tc[val.t][2], s.Tc[val.t][3], 0.8,
			val.x + val.sizeX, val.y + val.sizeY, s.Tc[val.t][4], s.Tc[val.t][5], 0.8,
			val.x, val.y + val.sizeY, s.Tc[val.t][6], s.Tc[val.t][7], 0.8,
		})
		vert = append(vert, v...)

		i := tobyte.Uint16Le([]uint16{
			next, next + 1, next + 2, next, next + 2, next + 3,
		})
		ind = append(ind, i...)
		next += 4

		s.parts[key] = val
	}
	s.dx, s.dy = 0.0, 0.0
	s.move = false
	return ind, vert, next, false
}
