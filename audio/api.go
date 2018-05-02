// Veselin Kostov April, 2017

// TODO: streaming instead of load entire audio file
// TODO: pass AudioParams at NewPlayer instead of hardcoded
// TODO: audio strenght per stream and/or general
package audio

import (
	"log"
	"math/rand"
	"sync"
	"unsafe"

	"graphs/engine/assets"
	"graphs/engine/audio/decoders/oggvorbis"
	"graphs/engine/audio/decoders/wav"
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

//Start player
func NewPlayer(p *Player) *Player {
	if p == nil {
		p = &Player{}
		p.in = make(chan interface{}, 100)
	}
	go poller(p)
	p.mutex.Lock()
	p.runing = true
	p.mutex.Unlock()
	return p
}

// Load wav file and add as source
// Loading is happening in main thread, audio will block if loading is there
func (p *Player) LoadWav(fname, streamName string, fhandle assets.OpenAsset) {
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
func (p *Player) LoadOgg(fname, streamName string, chandle assets.OpenAsset) {
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

func (p *Player) IsRuning() bool {
	return p.runing
}
