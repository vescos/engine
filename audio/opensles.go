// OpenSL ES backend(android)

// +build android

package audio

//#cgo LDFLAGS: -lOpenSLES
//#include "opensles.h"
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

func openDevice(aParams *AudioParams) unsafe.Pointer {
	aParams.frameSize = aParams.SampleSize * aParams.Channels
	aParams.frameRate = aParams.SampleRate / aParams.Channels
	aParams.periodSize = int(float32(aParams.frameRate) * (float32(aParams.PeriodTime) / 1000000))
	aParams.buffSize = aParams.periodSize * aParams.BuffSizeCnt
	aParams.buffBytes = aParams.buffSize * aParams.frameSize
	handle := unsafe.Pointer(C.getOslHandle(C.uint(aParams.BuffSizeCnt), C.uint(aParams.buffBytes)))
	if handle == nil {
		log.Print("Audio: openslES. E: malloc failed")
	}
	err := int(C.createEngine((*C.oslHandle)(handle)))
	if err > 0 {
		log.Printf("Audio: openslES can't create object engine. E: %v", openslGetError(err))
		return nil
	}
	return handle
}

func setAsyncWriteChan() chan int {
	//to avoid passing GO channel in C memory or keep channel as global var(shared library)
	//writeBuff is synchronous
	//create dummy channel
	return make(chan int, 10)
}

func setParams(handle unsafe.Pointer, aParams *AudioParams) bool {
	rate := C.uint(aParams.SampleRate)
	channels := C.uint(aParams.Channels)
	sample_size := C.uint(aParams.SampleSize)

	err := int(C.initOsl((*C.oslHandle)(handle)))
	if err > 0 {
		log.Printf("Audio: openslES setParams: initOsl. E: %v", openslGetError(err))
		return false
	}
	err = int(C.createPlayer((*C.oslHandle)(handle), rate, channels, sample_size))
	if err > 0 {
		log.Printf("Audio: openslES setParams: createPlayer. E: %v", openslGetError(err))
		return false
	}

	return true
}

//export availWriteCallback
func availWriteCallback(handle *C.oslHandle) {
	//TODO: find a way to notify player that write is available but don't store go channel in
	// C memory or using global var(shared lib)

}

func writeBuff(handle unsafe.Pointer, buff []byte, aParams *AudioParams) (bool, int, bool) {
	count := int(C.getQueuedBuffsCount((*C.oslHandle)(handle)))
	chunk := aParams.buffBytes / aParams.BuffSizeCnt
	chunk_max := 0
	start := 0
	sum_b := 0
	buff_len := len(buff)
	for i := count; i < aParams.BuffSizeCnt && sum_b < buff_len; i, start = i+1, start+chunk_max {
		chunk_max = chunk
		if buff_len-sum_b < chunk {
			chunk_max = buff_len - sum_b
		}
		sum_b += chunk_max
		sub := buff[start:sum_b]
		err := int(C.enqueueBuff((*C.oslHandle)(handle), unsafe.Pointer(&sub[0]), C.uint(len(sub))))
		if err > 0 {
			log.Printf("Audio: openslES enqueue. E: %v", openslGetError(err))
			return false, 0, false
		}
	}
	return true, sum_b, true
}

func closeDevice(handle unsafe.Pointer) {
	if handle != nil {
		C.closeDevice((*C.oslHandle)(handle))
	}
	handle = nil
}

func openslGetError(err int) string {
	switch err {
	case C.SL_RESULT_PRECONDITIONS_VIOLATED:
		return "SL_RESULT_PRECONDITIONS_VIOLATED"
	case C.SL_RESULT_PARAMETER_INVALID:
		return "SL_RESULT_PARAMETER_INVALID"
	case C.SL_RESULT_MEMORY_FAILURE:
		return "SL_RESULT_MEMORY_FAILURE"
	case C.SL_RESULT_RESOURCE_ERROR:
		return "SL_RESULT_RESOURCE_ERROR"
	case C.SL_RESULT_RESOURCE_LOST:
		return "SL_RESULT_RESOURCE_LOST"
	case C.SL_RESULT_IO_ERROR:
		return "SL_RESULT_IO_ERROR"
	case C.SL_RESULT_BUFFER_INSUFFICIENT:
		return "SL_RESULT_BUFFER_INSUFFICIENT"
	case C.SL_RESULT_CONTENT_CORRUPTED:
		return "SL_RESULT_CONTENT_CORRUPTED"
	case C.SL_RESULT_CONTENT_UNSUPPORTED:
		return "SL_RESULT_CONTENT_UNSUPPORTED"
	case C.SL_RESULT_CONTENT_NOT_FOUND:
		return "SL_RESULT_CONTENT_NOT_FOUND"
	case C.SL_RESULT_PERMISSION_DENIED:
		return "SL_RESULT_PERMISSION_DENIED"
	case C.SL_RESULT_FEATURE_UNSUPPORTED:
		return "SL_RESULT_FEATURE_UNSUPPORTED"
	case C.SL_RESULT_INTERNAL_ERROR:
		return "SL_RESULT_INTERNAL_ERROR"
	case C.SL_RESULT_UNKNOWN_ERROR:
		return "SL_RESULT_UNKNOWN_ERROR"
	case C.SL_RESULT_OPERATION_ABORTED:
		return "SL_RESULT_OPERATION_ABORTED"
	case C.SL_RESULT_CONTROL_LOST:
		return "SL_RESULT_CONTROL_LOST"
	default:
		return fmt.Sprintf("Unknown openslES error: %d", err)
	}
}
