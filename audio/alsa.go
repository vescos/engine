//ALSA(linux) backend

// +build linux,!android

package audio

//#cgo LDFLAGS: -lasound
//#include "alsa.h"
import "C"
import (
	"log"
	//"os"
	//"os/signal"
	//"syscall"
	"unsafe"
)

func openDevice(aParams *AudioParams) unsafe.Pointer {
	//FIXME: alocate handle in C memory
	var handle *C.snd_pcm_t
	name := C.CString("hw:0,0") //device name "hw:0,0" "default" is used by pulse audio
	defer C.free(unsafe.Pointer(name))
	//Last parameter in open is for mode async/nonblock. What is zero?
	err := int(C.snd_pcm_open(&handle, name, C.SND_PCM_STREAM_PLAYBACK, 0))
	if err < 0 {
		handle = nil
		log.Printf("Audio: can't open audio device: %v, E: %v", C.GoString(name), C.GoString(C.snd_strerror(C.int(err))))
	}
	return unsafe.Pointer(handle)
}

//func setAsyncWriteChan() chan os.Signal {
//sigioc := make(chan os.Signal, 10)
//signal.Notify(sigioc, syscall.SIGIO)
//return sigioc
//}

func setParams(handle unsafe.Pointer, aParams *AudioParams) bool {
	state := setHWParams(handle, aParams)
	if state {
		state = setSWParams(handle, aParams)
	}
	//if state {
	//return setAvailWriteCallback(handle)
	//}
	return state
}

func setHWParams(h unsafe.Pointer, aParams *AudioParams) bool {
	handle := (*C.snd_pcm_t)(h)
	rate := C.uint(aParams.SampleRate)
	channels := C.uint(aParams.Channels)
	period_time := C.uint(aParams.PeriodTime)
	buffer_time := C.uint(aParams.PeriodTime * aParams.BuffSizeCnt)
	var params *C.snd_pcm_hw_params_t
	var period_size, buffer_size C.snd_pcm_uframes_t

	defer C.snd_pcm_hw_params_free(params)
	err := int(C.snd_pcm_hw_params_malloc(&params))
	if err < 0 {
		log.Printf("Audio: can't malloc params. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_any(handle, params))
	if err < 0 {
		log.Printf("Audio: can't fill params. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_set_access(handle, params, C.SND_PCM_ACCESS_RW_INTERLEAVED))
	if err < 0 {
		log.Printf("Audio: can't set access param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_set_format(handle, params, C.SND_PCM_FORMAT_S16_LE))
	if err < 0 {
		log.Printf("Audio: can't set format param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_set_rate_near(handle, params, &rate, nil))
	if err < 0 {
		log.Printf("Audio: can't set rate_near param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_set_channels(handle, params, channels))
	if err < 0 {
		log.Printf("Audio: can't set channels param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_set_period_time_near(handle, params, &period_time, nil))
	if err < 0 {
		log.Printf("Audio: can't set period size param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_hw_params_set_buffer_time_near(handle, params, &buffer_time, nil))
	if err < 0 {
		log.Printf("Audio: can't set buffer size param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}

	//get hw rate, period_size and buffer_size
	C.snd_pcm_hw_params_get_rate(params, &rate, nil)
	C.snd_pcm_hw_params_get_period_size(params, &period_size, nil)
	C.snd_pcm_hw_params_get_buffer_size(params, &buffer_size)

	aParams.frameRate = int(rate) / aParams.Channels
	aParams.frameSize = int(aParams.SampleSize * aParams.Channels)
	aParams.periodSize = int(period_size)
	aParams.buffSize = int(buffer_size)
	aParams.buffBytes = aParams.buffSize * aParams.frameSize

	err = int(C.snd_pcm_hw_params(handle, params))
	if err < 0 {
		log.Printf("Audio: can't apply HW params. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	//log.Printf("State: %v", C.GoString(C.snd_pcm_state_name(C.snd_pcm_state(handle)))  )
	err = int(C.snd_pcm_prepare(handle))
	if err < 0 {
		log.Printf("Audio: can't prepare device. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}

	return true
}

func setSWParams(h unsafe.Pointer, aParams *AudioParams) bool {
	handle := (*C.snd_pcm_t)(h)
	var params *C.snd_pcm_sw_params_t

	defer C.snd_pcm_sw_params_free(params)
	err := int(C.snd_pcm_sw_params_malloc(&params))
	if err < 0 {
		log.Printf("Audio: can't malloc SW params. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_sw_params_current(handle, params))
	if err < 0 {
		log.Printf("Audio: can't get current SW params. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_sw_params_set_start_threshold(handle, params, C.snd_pcm_uframes_t(aParams.buffSize)))
	if err < 0 {
		log.Printf("Audio: can't set start treshold SW param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}
	err = int(C.snd_pcm_sw_params_set_avail_min(handle, params, C.snd_pcm_uframes_t(aParams.periodSize)))
	if err < 0 {
		log.Printf("Audio: can't set avail min SW param. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}

	err = int(C.snd_pcm_sw_params(handle, params))
	if err < 0 {
		log.Printf("Audio: can't apply SW params. E: %v", C.GoString(C.snd_strerror(C.int(err))))
		return false
	}

	return true
}

//export availWriteCallback
//func availWriteCallback(ahandler *C.snd_async_handler_t) {
//this function is never called
//"Notify disables the default behavior for a given set of asynchronous signals
//and instead delivers them over one or more registered channels"
//}

//func setAvailWriteCallback(handle *C.snd_pcm_t) bool {
//err := int(C.cSetAvailWriteCallback(handle))
//if err < 0 {
//return false
//}
//return true
//}

func writeBuff(h unsafe.Pointer, buff []byte, aParams *AudioParams) (bool, int, bool) {
	handle := (*C.snd_pcm_t)(h)
	frames := C.snd_pcm_uframes_t(len(buff) / int(aParams.frameSize))

	err := int(C.snd_pcm_avail_update(handle))
	if err > 0 {
		err = int(C.snd_pcm_writei(handle, unsafe.Pointer(&buff[0]), C.snd_pcm_uframes_t(err)))
	}

	if err < 0 {
		if err == -int(C.EPIPE) {
			err = int(C.snd_pcm_prepare(handle))
			if err < 0 {
				log.Printf("Audio: Can't recover underrun, prepare failed. E: %v", C.GoString(C.snd_strerror(C.int(err))))
				return false, 0, false
			}
			err = int(C.snd_pcm_writei(handle, unsafe.Pointer(&buff[0]), frames))
			if err < 0 {
				log.Printf("Audio: Can't recover, write failed. E: %v", C.GoString(C.snd_strerror(C.int(err))))
				return false, 0, false
			}
		} else {
			log.Printf("Audio: Can't write to device. E: %v", C.GoString(C.snd_strerror(C.int(err))))
			return false, 0, false
		}
	}
	return true, err * aParams.Channels * aParams.SampleSize, true
}

func closeDevice(h unsafe.Pointer) {
	handle := (*C.snd_pcm_t)(h)
	if handle != nil {
		C.snd_pcm_close(handle)
	}
}
