//Veselin Kostov April, 2017

//TODO: not thread safe
//Start, Stop are not thread safe
//all other functions are thread safe but only if they happen between Start and Stop
package audio

import (
	"log"

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
}

//Start player
func (p *Player) Start() {
	p.in = make(chan interface{}, 100)
	go poller(p)
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
	p.in <- &ctlSource{name: name, cmd: CmdSourceRemove}
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
	p.in <- &ctlStream{streamId: streamId, cmd: CmdStreamPause, complete: complete}
}

func (p *Player) PlayStream(streamId string) {
	p.in <- &ctlStream{streamId: streamId, cmd: CmdStreamPlay}
}

func (p *Player) RestartStream(streamId string) {
	p.in <- &ctlStream{streamId: streamId, cmd: CmdStreamRestart}
}

//Remove stream
//if complete is true, current playng sound will be completed
func (p *Player) RemoveStream(streamId string, complete bool) {
	p.in <- &ctlStream{streamId: streamId, cmd: CmdStreamRemove, complete: complete}
}

func (p *Player) Mute() {
	p.in <- CmdPlayerMute
}

func (p *Player) Unmute() {
	p.in <- CmdPlayerUnmute
}

func (p *Player) Pause() {
	p.in <- CmdPlayerPause
}

func (p *Player) Resume() {
	p.in <- CmdPlayerResume
}

func (p *Player) Stop() { //TODO: reset
	p.in <- CmdPlayerStop
}
