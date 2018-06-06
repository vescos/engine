package ui

import (
	"log"

	"github.com/vescos/engine/draw2d/screen"
	"github.com/vescos/engine/geom"
	"github.com/vescos/engine/tobyte"
)

type State struct {
	Size     screen.Size
	elements map[string]*Element
}

type AddElement struct {
	Id      string
	Element *Element
}

type DelElement struct {
	Id string
}

type DrawMode int32

const (
	ModeScale   DrawMode = iota //scale el width against smaller screen dimension
	ModeNoScale                 //keep as is
)

type Element struct {
	X         float32 // % x offset
	Y         float32 // % y offset
	W         float32 // % element width (from smallest dimension)
	R         float32 // % element aspect ratio
	Mode      DrawMode
	Tc        [8]float32
	Rectangle *geom.Rectangle
}

func (s *State) Start() {
	s.elements = make(map[string]*Element)
}

func (s *State) OnSize(w, h float32) {
	s.Size.W = w
	s.Size.H = h
	//resize elements
	for id, _ := range s.elements {
		s.elements[id].Rectangle = GetRectangle(*s.elements[id], s.Size)
	}
}

func (s *State) Ctl(itf interface{}) {
	switch itf.(type) {
	case *AddElement:
		el := itf.(*AddElement)
		if _, ok := s.elements[el.Id]; ok {
			log.Printf("Effects: ui: this element exists - overriding: Id: %v", el.Id)
		}
		s.elements[el.Id] = el.Element
	case *DelElement:
		el := itf.(*DelElement)
		delete(s.elements, el.Id)
	default:
		log.Printf("Effects: ui: Unsupported interface: %T", itf)
	}
}

func (s *State) OnFrame(ind, vert []byte, next uint16) ([]byte, []byte, uint16, bool) {
	for _, el := range s.elements {
		r := el.Rectangle
		vert = append(vert, tobyte.Float32Le([]float32{
			r.Min.GlX, r.Max.GlY, el.Tc[0], el.Tc[1], 1,
			r.Min.GlX, r.Min.GlY, el.Tc[2], el.Tc[3], 1,
			r.Max.GlX, r.Min.GlY, el.Tc[4], el.Tc[5], 1,
			r.Max.GlX, r.Max.GlY, el.Tc[6], el.Tc[7], 1,
		})...)
		ind = append(ind, tobyte.Uint16Le([]uint16{
			next, next + 1, next + 2, next, next + 2, next + 3,
		})...)
		//log.Printf("%+v, %+v", el, el.Rectangle)
		next += 4
	}

	return ind, vert, next, false
}

func GetRectangle(el Element, size screen.Size) *geom.Rectangle {
	var dt, sx, sy, ex, ey, x_add, y_add float32
	w := el.W
	dt = 1.0 / 50.0
	if el.Mode == ModeScale {
		wh := size.W / size.H
		if size.W > size.H {
			w = w / wh
			if el.W != 100 {
				x_add = ((el.W - w) * el.X) / (100 - el.W)
			}
		} else if el.Y != 100 {
			//TODO: optimize math
			s := w * el.R
			y_add = ((s - ((size.W * w * el.R) / size.H)) * el.Y) / (100 - s)
		}
		width := dt * w
		height := dt * ((size.W * w * el.R) / size.H)

		sx = (dt * (el.X + x_add)) - 1.0
		ex = sx + width
		sy = -(dt * (el.Y + y_add)) + 1.0
		ey = sy - height

	} else {
		sx = (dt * el.X) - 1.0
		ex = (dt * (w + el.X)) - 1.0
		sy = -(dt * el.Y) + 1.0
		ey = -(dt * (((size.W * w * el.R) / size.H) + el.Y)) + 1.0
	}

	return &geom.Rectangle{
		Min: geom.Point{
			X:   ((sx + 1) * size.W) / 2,
			Y:   ((1 - sy) * size.H) / 2,
			GlX: sx,
			GlY: sy,
		},
		Max: geom.Point{
			X:   ((ex + 1) * size.W) / 2,
			Y:   ((1 - ey) * size.H) / 2,
			GlX: ex,
			GlY: ey,
		},
	}
}
