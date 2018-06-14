// Veselin Kostov April, 2017

// mixer supports oggvorbis or wav, 16 bit, stereo, 44100 or 48000 audio, interleaved, LE.

// TODO: add streaming for audio files
// TODO: increase/decrease audio strenght by distance(per stream).
// TODO: increase/decrease audio strenght(for player).
package audio

import (
	"log"
	"math/rand"
	"sync"
	"unsafe"

	"github.com/vescos/engine/assets"
	"github.com/vescos/engine/audio/decoders/oggvorbis"
	"github.com/vescos/engine/audio/decoders/wav"
)

type Channels int

const (
	Stereo Channels = 2
)

type SampleSize int

const (
	SampleSize16 SampleSize = 2
)

type SampleRate int

const (
	SampleRate44100 SampleRate = 44100
	SampleRate48000 SampleRate = 48000
)

type AccessMode int // access mode
const (
	Interleaved AccessMode = iota
	//NonInterleaved
)

type Source struct {
	Name   string
	Buff   []int16
	length int
}

type Player struct {
	in                 chan interface{}
	initialized        bool
	state              bool
	play               bool
	mute               bool
	emptyStreamIdCount int
	sources            map[string]*Source
	streams            map[string]*stream
	handle             unsafe.Pointer
	rand               *rand.Rand
	// can be readed/writen from different goroutines
	// use sync.Mutex
	runing bool
	mutex  sync.Mutex
}

type AudioParams struct {
	//Format int
	AccessMode  AccessMode // interleaved, noninterleaved (NOTIMPLEMENTED)
	SampleRate  SampleRate // samples per second in Hz
	SampleSize  SampleSize // in bytes 16bit = 2 bytes - 20bit and other non multiple of 8 are not supported
	Channels    Channels   // Mono, Stereo
	PeriodTime  int        // in microsecons (us)
	BuffSizeCnt int        // buffSize = periodSize * BuffSizeCnt

	periodSize int // in frames, sampleRate * (PeriodTime / 1000000)
	buffSize   int // in frames, periodSize * BuffSizeCnt (or more?)
	buffBytes  int // in bytes, buffSize * frameSize
	frameSize  int // SampleSize * Channels
}

//Start player
func NewPlayer(p *Player, params AudioParams) *Player {
	if p == nil {
		p = &Player{}
	} else {
		if len(p.in) > 0 {
			log.Printf("Audio: NewPlayer: discarding %v commands.", len(p.in))
		}
	}
	p.in = make(chan interface{}, 100)
	go poller(p, &params)
	p.mutex.Lock()
	p.runing = true
	p.mutex.Unlock()
	return p
}

// Load wav file and add as source
// Loading is happening in main thread, audio will block if loading is there
func (p *Player) LoadWav(fname, streamName string, fhandle assets.FileManager) {
	file, err := fhandle.OpenAsset(fname)
	if err != nil {
		log.Printf("Audio: loadStream: can't open file error: %v, filename: %v", err, fname)
	} else {
		defer file.Close()
		dec := &wav.State{}
		buff := dec.Decode(file)
		p.AddSource(streamName, buff)
	}
}

// Load oggvorbis file and add as source
// require cfile handle
func (p *Player) LoadOgg(fname, streamName string, chandle assets.FileManager) {
	file, err := chandle.OpenAsset(fname)
	if err != nil {
		log.Printf("Audio: loadStream: can't open file error: %v, filename: %v", err, fname)
	} else {
		defer file.Close()
		dec := &oggvorbis.State{}
		buff := dec.Decode(file)
		p.AddSource(streamName, buff)
	}

}

func (p *Player) AddSource(name string, buff []int16) {
	p.in <- &Source{Name: name, Buff: buff}
}

func (p *Player) AddSourceStruct(s *Source) {
	p.in <- s
}

// Remove source and all streams that use this source
// FIXME: remove source that belong to group stream will crash engine
func (p *Player) RemoveSource(name string) {
	p.in <- &ctlSource{name: name, cmd: cmdSourceRemove}
}

// Play source once - no control
func (p *Player) Play(srcName string) {
	p.in <- &addStream{srcName: srcName, play: true}
}

// Add and Play stream - control: loop, Pause, Resume Remove
func (p *Player) AddStream(srcName string, streamId string, play bool, loop bool) {
	p.in <- &addStream{srcName: srcName, streamId: streamId, play: play, loop: loop}
}

func (p *Player) AddRandStream(srcName string, streamId string, play bool, min uint, max uint) {
	p.in <- &addRandStream{srcName: srcName, streamId: streamId, play: play, min: min, max: max}
}

func (p *Player) AddIntervalStream(srcName string, streamId string, play bool, interval uint) {
	p.in <- &addIntervalStream{srcName: srcName, streamId: streamId, play: play, interval: interval}
}

func (p *Player) AddGroupStream(src []string, streamId string, play bool, interval uint) {
	p.in <- &addGroupStream{src: src, streamId: streamId, play: play, interval: interval}
}

//Pause stream
//if complete is true, current playng sound will be completed
func (p *Player) PauseStream(streamId string, complete bool) {
	p.in <- &ctlStream{streamId: streamId, cmd: cmdStreamPause, complete: complete}
}

func (p *Player) PlayStream(streamId string) {
	p.in <- &ctlStream{streamId: streamId, cmd: cmdStreamPlay}
}

func (p *Player) RestartStream(streamId string) {
	p.in <- &ctlStream{streamId: streamId, cmd: cmdStreamRestart}
}

//Remove stream
//if complete is true, current playng sound will be completed
func (p *Player) RemoveStream(streamId string, complete bool) {
	p.in <- &ctlStream{streamId: streamId, cmd: cmdStreamRemove, complete: complete}
}

func (p *Player) Mute() {
	p.in <- cmdPlayerMute
}

func (p *Player) Unmute() {
	p.in <- cmdPlayerUnmute
}

func (p *Player) Pause() {
	p.in <- cmdPlayerPause
}

func (p *Player) Resume() {
	p.in <- cmdPlayerResume
}

func (p *Player) Stop() {
	p.mutex.Lock()
	p.runing = false
	p.mutex.Unlock()
	p.in <- cmdPlayerStop
}

// Stop() can happen between call to isRunning() and call to api.
// As channel is buffered such calls will not block (almost till channel is not full)
// but all commands sent between Stop() and NewPlayer(stoped_player) will be discarded.
func (p *Player) IsRuning() bool {
	return p.runing
}
