// Veselin Kostov 21 Feb, 2018
// Decode vorbis file using stb_vorbis.c decoder from
// http://nothings.org/stb_vorbis/

package oggvorbis

/*
#cgo LDFLAGS: -lm

#include "stb_vorbis.h"
#include <stdlib.h>
*/
import "C"
import (
	"log"
	"unsafe"

	"github.com/vescos/engine/assets"
	//"github.com/vescos/tobyte"
)

type State struct{}

func (s *State) Decode(file assets.Asset) []int16 {
	buff_size := 4096 - 32
	fd := (*C.FILE)(unsafe.Pointer(file.Fd()))
	err := C.int(0)
	//C.stb_vorbis
	sv := C.stb_vorbis_open_file(fd, 0, &err, nil)
	if sv == nil {
		log.Printf("Oggvorbis: can't open file: %v, err: VORBIS_file_open_failure", file.Name())
		return nil
	}
	defer C.stb_vorbis_close(sv)
	//info := C.stb_vorbis_get_info(sv)
	//channels := info.channels
	//log.Printf("%+v", info)

	//TODO: find raw data size to avoid realocation on append
	rbuff := make([]int16, 0, 2*buff_size*20)
	cbuff := (*C.short)(C.malloc(C.size_t(2 * buff_size)))
	defer C.free(unsafe.Pointer(cbuff))
	for {
		// n - number of samples read per channel
		// for 2 channels 16bit audio samples read is n*2, bytes read is n*4
		n := C.stb_vorbis_get_samples_short_interleaved(sv, 2, cbuff, C.int(buff_size))
		if n == 0 {
			break
		}
		// this copy data twice
		//tmp_buff := tobyte.ByteLeToInt16(C.GoBytes(unsafe.Pointer(cbuff), C.int(n*4)))

		// this is much faster
		// see https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
		// 1 << 30 is too large for 32bit target
		length := n * 2
		tmp_buff := (*[1 << 20]int16)(unsafe.Pointer(cbuff))[:length:length]

		rbuff = append(rbuff, tmp_buff...)
	}

	return rbuff
}
