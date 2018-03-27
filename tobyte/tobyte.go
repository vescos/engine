//convert slices to byte slices

package tobyte

import (
	"log"
	"math"
)

func Float32Le(arr []float32) []byte {
	b := make([]byte, 4*len(arr))
	n := 0
	for _, v := range arr {
		u := math.Float32bits(v)
		b[n] = byte(u)
		b[n+1] = byte(u >> 8)
		b[n+2] = byte(u >> 16)
		b[n+3] = byte(u >> 24)
		n += 4
	}
	return b
}

func Uint16Le(arr []uint16) []byte {
	b := make([]byte, 2*len(arr))
	n := 0
	for _, v := range arr {
		b[n] = byte(v)
		b[n+1] = byte(v >> 8)
		n += 2
	}
	return b
}

func ByteLeToInt16(arr []byte) []int16 {
	l := len(arr)
	if l <= 0 || l%2 != 0 {
		log.Print("ByteLeToUint16: len(arr) is expected as even number > 0: len: ", l)
		return nil
	}
	u := make([]int16, l/2)
	for i, j := 0, 0; i < l; i, j = i+2, j+1 {
		u[j] = int16(arr[i]) | (int16(arr[i+1]) << 8)
	}
	return u
}
