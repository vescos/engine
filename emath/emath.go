//Engine math emath

package emath

import (
	"log"
	"math"
)

const (
	PI float32 = 3.141592653589793238462643383
)

func Mat4Multi(m, n []float32) []float32 {
	return []float32{
		m[0]*n[0] + m[4]*n[1] + m[8]*n[2] + m[12]*n[3],
		m[1]*n[0] + m[5]*n[1] + m[9]*n[2] + m[13]*n[3],
		m[2]*n[0] + m[6]*n[1] + m[10]*n[2] + m[14]*n[3],
		m[3]*n[0] + m[7]*n[1] + m[11]*n[2] + m[15]*n[3],
		m[0]*n[4] + m[4]*n[5] + m[8]*n[6] + m[12]*n[7],
		m[1]*n[4] + m[5]*n[5] + m[9]*n[6] + m[13]*n[7],
		m[2]*n[4] + m[6]*n[5] + m[10]*n[6] + m[14]*n[7],
		m[3]*n[4] + m[7]*n[5] + m[11]*n[6] + m[15]*n[7],
		m[0]*n[8] + m[4]*n[9] + m[8]*n[10] + m[12]*n[11],
		m[1]*n[8] + m[5]*n[9] + m[9]*n[10] + m[13]*n[11],
		m[2]*n[8] + m[6]*n[9] + m[10]*n[10] + m[14]*n[11],
		m[3]*n[8] + m[7]*n[9] + m[11]*n[10] + m[15]*n[11],
		m[0]*n[12] + m[4]*n[13] + m[8]*n[14] + m[12]*n[15],
		m[1]*n[12] + m[5]*n[13] + m[9]*n[14] + m[13]*n[15],
		m[2]*n[12] + m[6]*n[13] + m[10]*n[14] + m[14]*n[15],
		m[3]*n[12] + m[7]*n[13] + m[11]*n[14] + m[15]*n[15],
	}
}

func Mat4Identity() []float32 {
	return []float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func Mat4Transpose(m []float32) []float32 {
	return []float32{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
}

func Mat4Translate(x, y, z float32) []float32 {
	return []float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func Mat4Scale(x, y, z float32) []float32 {
	return []float32{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func Mat4RotationX(a float32) []float32 {
	return []float32{
		1, 0, 0, 0,
		0, Cos(a), Sin(a), 0,
		0, -Sin(a), Cos(a), 0,
		0, 0, 0, 1,
	}
}

func Mat4RotationY(a float32) []float32 {
	return []float32{
		Cos(a), 0, -Sin(a), 0,
		0, 1, 0, 0,
		Sin(a), 0, Cos(a), 0,
		0, 0, 0, 1,
	}
}

func Mat4RotationZ(a float32) []float32 {
	return []float32{
		Cos(a), Sin(a), 0, 0,
		-Sin(a), Cos(a), 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func Mat4RotationXYZ(xa, ya, za float32) []float32 {
	return Mat4Multi(Mat4RotationX(xa), Mat4Multi(Mat4RotationY(ya), Mat4RotationZ(za)))
}

func Mat4Frustum(left, right, bottom, top, near, far float32) []float32 {
	return []float32{
		(2 * near) / (right - left), 0, 0, 0,
		0, (2 * near) / (top - bottom), 0, 0,
		(right + left) / (right - left), (top + bottom) / (top - bottom), -((far + near) / (far - near)), -1,
		0, 0, -((2 * far * near) / (far - near)), 0,
	}
}

func Mat4Perspective(angle, ratio, near, far float32) []float32 {
	var top = near * Tan(angle)
	return Mat4Frustum(-top*ratio, top*ratio, -top, top, near, far)
}

func Mat4vec4Multi(m, v []float32) []float32 {
	return []float32{
		m[0]*v[0] + m[4]*v[1] + m[8]*v[2] + m[12]*v[3],
		m[1]*v[0] + m[5]*v[1] + m[9]*v[2] + m[13]*v[3],
		m[2]*v[0] + m[6]*v[1] + m[10]*v[2] + m[14]*v[3],
		m[3]*v[0] + m[7]*v[1] + m[11]*v[2] + m[15]*v[3],
	}
}

///////////////////////////////////////////
// VECTORS
///////////////////////////////////////////

func Vec3Magnitude(v []float32) float32 {
	return RoundPrec(Sqrt(v[0]*v[0]+v[1]*v[1]+v[2]*v[2]), 10)
}
func Vec2Magnitude(v []float32) float32 {
	return RoundPrec(Sqrt(v[0]*v[0]+v[1]*v[1]), 10)
}
func Vec3Substract(v1, v2 []float32) []float32 {
	return []float32{v1[0] - v2[0], v1[1] - v2[1], v1[2] - v2[2]}
}
func Vec2Substract(v1, v2 []float32) []float32 {
	return []float32{v1[0] - v2[0], v1[1] - v2[1]}
}
func Vec3ScalarMulti(v []float32, c float32) []float32 {
	return []float32{v[0] * c, v[1] * c, v[2] * c}
}
func Vec2ScalarMulti(v []float32, c float32) []float32 {
	return []float32{v[0] * c, v[1] * c}
}
func Vec3Normalize(v []float32) []float32 {
	l := Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if l > 0 {
		return []float32{v[0] / l, v[1] / l, v[2] / l}
	}
	log.Print("vec3Normalize: can't normalize zero vector: ", v)
	return v
}
func Vec2Normalize(v []float32) []float32 {
	l := Sqrt(v[0]*v[0] + v[1]*v[1])
	if l > 0 {
		return []float32{v[0] / l, v[1] / l}
	}
	log.Print("vec2Normalize: can't normalize zero vector: ", v)
	return v
}
func Vec3Dot(v1, v2 []float32) float32 {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}
func Vec2Dot(v1, v2 []float32) float32 {
	return v1[0]*v2[0] + v1[1]*v2[1]
}
func Vec3Projection(v1, v2 []float32) []float32 {
	l := Vec3Magnitude(v2)
	return Vec3ScalarMulti(v2, Vec3Dot(v1, v2)/(l*l))
}
func Vec2Projection(v1, v2 []float32) []float32 {
	l := Vec2Magnitude(v2)
	return Vec2ScalarMulti(v2, Vec2Dot(v1, v2)/(l*l))
}
func Vec3Distance(v1, v2 []float32) float32 {
	return Vec3Magnitude(Vec3Substract(v1, v2))
}
func Vec2Distance(v1, v2 []float32) float32 {
	return Vec2Magnitude(Vec2Substract(v1, v2))
}

func Vec3SurfaceNormal(s []float32, reverse int32) []float32 {
	var u, v []float32
	if reverse == 1 {
		u = []float32{s[3] - s[6], s[4] - s[7], s[5] - s[8]}
		v = []float32{s[0] - s[6], s[1] - s[7], s[2] - s[8]}
	} else {
		u = []float32{s[3] - s[0], s[4] - s[1], s[5] - s[2]}
		v = []float32{s[6] - s[0], s[7] - s[1], s[8] - s[2]}
	}
	return Vec3Normalize([]float32{
		(u[1] * v[2]) - (u[2] * v[1]),
		(u[2] * v[0]) - (u[0] * v[2]),
		(u[0] * v[1]) - (u[1] * v[0]),
	})
}

func Vec2Angle3Points(v1, v2, v3 []float32) float32 {
	a := Vec2Distance(v1, v2)
	b := Vec2Distance(v2, v3)
	c := Vec2Distance(v1, v3)
	return Acos((a*a + b*b - c*c) / (2 * a * b))
}

func Vec2InConvex(x, z float32, pl []float32) bool {
	if len(pl) < 6 {
		return false
	}
	l := len(pl)
	positive := 0
	c := float32(0)
	for i := 0; i < l; i += 2 {
		if i == l-2 {
			c = (x-pl[i])*(z-pl[1]) - (z-pl[i+1])*(x-pl[0])
		} else {
			c = (x-pl[i])*(z-pl[i+3]) - (z-pl[i+1])*(x-pl[i+2])
		}

		if positive == 0 {
			if c > 0 {
				positive = 1
			} else if c < 0 {
				positive = -1
			}
		}
		if (c > 0 && positive < 0) || (c < 0 && positive > 0) {
			return false
		}
	}
	return true
}

func Line2Intersection(d []float32) ([]float32, bool) {
	var m, n, b, c float32 = 0, 0, 0, 0
	var mb, nb bool
	var ic = make([]float32, 2)
	if RoundPrec(d[0]-d[2], 10) != 0.0 {
		m = (d[1] - d[3]) / (d[0] - d[2])
		mb = true
	}
	if RoundPrec(d[4]-d[6], 10) != 0 {
		n = (d[5] - d[7]) / (d[4] - d[6])
		nb = true
	}
	if (!mb && !nb) || (m == n && mb && nb) {
		return ic, false
	}
	if !mb {
		ic[0] = d[0]
		c = d[5] - n*d[4]
		ic[1] = n*d[0] + c
	} else if !nb {
		ic[0] = d[4]
		b = d[1] - m*d[0]
		ic[1] = m*d[4] + b
	} else {
		b = d[1] - m*d[0]
		c = d[5] - n*d[4]
		ic[0] = (c - b) / (m - n)
		ic[1] = m*ic[0] + b
	}
	return ic, true
}

func Vec3TexCoord(p []float32, scaleX, scaleY float32) []float32 {
	vec1 := Vec3Substract(p[0:3], p[3:6])
	length := Vec3Magnitude(vec1)
	t := []float32{
		0, 0,
		0, length * scaleY,
	}
	for i := 6; i < len(p); i += 3 {
		vec2 := Vec3Substract(p[3:6], p[i:i+3])
		projv := Vec3Projection(vec2, vec1)
		t = append(t,
			Vec3Distance(projv, vec2)*scaleX,
			(length-Vec3Magnitude(projv))*scaleY)
	}
	return t
}
func Vec2TexCoord(p []float32, scaleX, scaleY float32) []float32 {
	vec1 := Vec2Substract(p[0:2], p[2:4])
	length := Vec2Magnitude(vec1)
	t := []float32{
		0, 0,
		0, length * scaleY,
	}
	for i := 4; i < len(p); i += 2 {
		vec2 := Vec2Substract(p[2:4], p[i:i+2])
		projv := Vec2Projection(vec2, vec1)
		t = append(t,
			Vec2Distance(projv, vec2)*scaleX,
			(length-Vec2Magnitude(projv))*scaleY)
	}
	return t
}

// 32 bit Wrapers

func Abs(j float32) float32 {
	if j < 0 {
		return -j
	}
	return j
}

func Max(x, y float32) float32 {
	if x >= y {
		return x
	}
	return y
}

func RoundPrec(x float32, prec int) float32 {
	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := float64(x) * pow
	_, frac := math.Modf(intermed)

	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return float32(rounder / pow * sign)
}

//Wrappes to math functions
func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}
func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}
func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}
func Tan(x float32) float32 {
	return float32(math.Tan(float64(x)))
}
func Acos(x float32) float32 {
	return float32(math.Acos(float64(x)))
}
func Pow(x, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}
