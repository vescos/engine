package geom

type Point struct {
	X   float32
	Y   float32
	GlX float32
	GlY float32
}
type Rectangle struct {
	Min Point
	Max Point
}

func IsPointInRectangle(p *Point, r *Rectangle) bool {
	if p == nil || r == nil {
		return false
	}
	if p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y {
		return true
	}
	return false
}
