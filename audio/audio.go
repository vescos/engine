// Veselin Kostov Jan, 2017
//
// Part of "Labyrinth lost gems" game engine android/linux version.
// Audio player(mixer), ALSA and OpenSL ES backends

// TODO: increase/decrease audio strenght by distance.
// TODO: more formats (only WAV|OGG 16bit 2channels interleaved LE is supported for now)
package audio

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

type Cmd int

const (
	// control player
	CmdPlayerStop Cmd = iota
	CmdPlayerMute
	CmdPlayerUnmute
	CmdPlayerPause
	CmdPlayerResume
	// control sources
	CmdSourceRemove
	// control streams
	CmdStreamPlay
	CmdStreamPause
	CmdStreamRestart
	CmdStreamRemove
)

type streamType int

const (
	typeNormal streamType = iota
	typeInterval
	typeRand
	typeGroup
)

type ctlSource struct {
	name string
	cmd  Cmd
}

type stream struct {
	srcName     string
	src         []string
	streamId    string
	streamType  streamType
	playing     bool
	readAt      int
	loop        bool
	paused      bool
	interval    uint
	min         uint
	max         uint
	startTime   time.Time
	deleted     bool
	lastGroupRd int
}

type addStream struct {
	srcName  string
	streamId string
	play     bool
	loop     bool
}

type addRandStream struct {
	srcName  string
	streamId string
	play     bool
	min      uint // in ms
	max      uint // in ms
}

type addIntervalStream struct {
	srcName  string
	streamId string
	play     bool
	interval uint // in ms
}

type addGroupStream struct {
	src      []string
	streamId string
	play     bool
	interval uint // in ms
}

type ctlStream struct {
	streamId string
	cmd      Cmd
	complete bool
}

// audio parameters, HARDCODED for now
type AccessMode int // access mode
const (
	ModeInterleaved AccessMode = iota
	ModeNonInterleaved
)

type AudioParams struct {
	//Format int
	//AccessMode int // interleaved, noninterleaved (NOTIMPLEMENTED)
	SampleRate  int // samples per second
	SampleSize  int // in bytes 16bit = 2 bytes
	Channels    int // mono = 1, stereo = 2, 5.1 = 6
	PeriodTime  int // in microsecons (us)
	BuffSizeCnt int // buffSize = periodSize * BuffSizeCnt

	periodSize int // in frames, frameRate * (PeriodTime / 1000000)
	buffSize   int // in frames, periodSize * BuffSizeCnt (or more?)
	buffBytes  int // in bytes, buffSize * frameSize
	frameRate  int // SampleRate / Channels
	frameSize  int // SampleSize * Channels
}

const (
	splitStr      string = "__!__"
	emptyStreamId string = "sgweedgtehbxc235SDFfsd@#Fs_"
	rbc           int    = 15876 * 5 // why 79385
)

func poller(player *Player) {
	log.Print(">>>>> audio: Start.")
	if !player.initialized {
		player.sources = make(map[string]*Source, 0)
		player.streams = make(map[string]*stream, 0)
		player.play = true
		player.initialized = true
		player.emptyStreamIdCount = 0
	}

	write_avail := true
	sigioc := setAsyncWriteChan()

	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		mix_buff  []byte
		ring_buff [rbc]byte
		rbs       int = 0 // ring buff start
	)

	// HARDCODED for now
	params := &AudioParams{
		//AccessMode: ModeInterleaved,
		SampleRate:  44100,
		SampleSize:  2,
		Channels:    2,
		PeriodTime:  30 * 1000, // 30ms //TODO: test latency vs performance
		BuffSizeCnt: 3,
	}
	handle := openDevice(params)

	if handle != nil {
		player.state = setParams(handle, params)
	}

	mix_buff = ring_buff[:0:params.buffBytes]

	// player loop
	for {
		select {
		// TODO: SLES backend - use sigioc chan to signal that write is available instead of using synchronous writes
		default: // don't block on select, instead sleep on loop end
		// comm with output buffer
		case <-sigioc:
			write_avail = true
		// communication with main thread
		case e := <-player.in:
			switch e.(type) {
			case Cmd:
				command := e.(Cmd)
				switch command {
				case CmdPlayerStop:
					closeDevice(handle)
					player.state = false
					log.Print("<<<<< audio: Stop.")
					return
				case CmdPlayerMute:
					player.mute = true
				case CmdPlayerUnmute:
					player.mute = false
				case CmdPlayerPause:
					player.play = false
				case CmdPlayerResume:
					player.play = true
				}
			case *Source:
				a := e.(*Source)
				if a.Name != "" {
					a.length = len(a.Buff)
					if a.length > 0 {
						//TODO: check that sorce is not in the map
						player.sources[a.Name] = a
					} else {
						log.Printf("Audio: Source.Buff len is 0. Skipping source. Name: %v", a.Name)
					}
				} else {
					log.Print("Audio: Source.Name is missing. Skipping source.")
				}
			case *ctlSource:
				ctl := e.(*ctlSource)
				if ctl.cmd == CmdSourceRemove {
					//delete streams for given source
					for key, stream := range player.streams {
						if stream.srcName == ctl.name {
							delete(player.streams, key) //safe to delete key in range loop
						}
					}
					delete(player.sources, ctl.name)
				}
			case *ctlStream:
				ctl := e.(*ctlStream)
				if _, ok := player.streams[ctl.streamId]; ok {
					if ctl.cmd == CmdStreamPause {
						if !ctl.complete {
							player.streams[ctl.streamId].playing = false
						}
						player.streams[ctl.streamId].paused = true
					} else if ctl.cmd == CmdStreamPlay {
						player.streams[ctl.streamId].playing = true
						player.streams[ctl.streamId].paused = false
					} else if ctl.cmd == CmdStreamRemove {
						if ctl.complete && player.streams[ctl.streamId].playing {
							player.streams[ctl.streamId].deleted = true
							//rename stream to allow reuse of streamId
							stream_id := emptyStreamId + strconv.Itoa(player.emptyStreamIdCount)
							player.emptyStreamIdCount += 1
							st := player.streams[ctl.streamId]
							delete(player.streams, ctl.streamId)
							player.streams[stream_id] = st
						} else {
							delete(player.streams, ctl.streamId)
						}
					} else if ctl.cmd == CmdStreamRestart {
						player.streams[ctl.streamId].playing = true
						player.streams[ctl.streamId].paused = false
						player.streams[ctl.streamId].readAt = 0
					}
				} else {
					log.Printf("Audio: CtlStream not in the map: streamId: %v", ctl.streamId)
				}
			case *addStream:
				ctl := e.(*addStream)
				_, ok := player.sources[ctl.srcName]
				if ok && ctl.srcName != "" {
					var stream_id string
					if ctl.streamId == "" {
						stream_id = emptyStreamId + strconv.Itoa(player.emptyStreamIdCount)
						player.emptyStreamIdCount += 1
					} else {
						stream_id = ctl.streamId
					}
					//TODO: check that stream_id is valid map key
					if _, ok := player.streams[stream_id]; ok {
						log.Printf("Audio: Add: stream exists - skipping: %v", stream_id)
					} else {
						player.streams[stream_id] = &stream{
							srcName:    ctl.srcName,
							streamId:   stream_id,
							playing:    ctl.play,
							readAt:     0,
							loop:       ctl.loop,
							streamType: typeNormal,
						}
					}
				} else {
					if !ok {
						log.Printf("Audio: Source is not in the map - Name: %v", ctl.srcName)
					} else if ctl.srcName == "" {
						log.Print("Audio: empty setStream.SrcName: skipping.")
					}
				}
			case *addRandStream:
				ctl := e.(*addRandStream)
				if _, ok := player.streams[ctl.streamId]; ok {
					log.Printf("Audio: Rand stream exists - skipping: %v", ctl.streamId)
				} else if _, ok := player.sources[ctl.srcName]; !ok {
					log.Printf("Audio: Rand: Source is not in the map - Name: %v", ctl.srcName)
				} else if ctl.max < ctl.min {
					log.Printf("Audio: max < min??? - skipping: %v", ctl.streamId)
				} else {
					player.streams[ctl.streamId] = &stream{
						srcName:    ctl.srcName,
						streamId:   ctl.streamId,
						playing:    false,
						paused:     !ctl.play,
						readAt:     0,
						streamType: typeRand,
						min:        ctl.min,
						max:        ctl.max,
					}
					if ctl.play {
						rd := random.Intn(int(ctl.max-ctl.min)) + int(ctl.min)
						st := time.Now().Add(time.Duration(rd) * time.Millisecond)
						player.streams[ctl.streamId].startTime = st
					}
				}
			case *addIntervalStream:
				ctl := e.(*addIntervalStream)
				if _, ok := player.streams[ctl.streamId]; ok {
					log.Printf("Audio: Interval: stream exists - skipping: %v", ctl.streamId)
				} else if _, ok := player.sources[ctl.srcName]; !ok {
					log.Printf("Audio: Interval: Source is not in the map - Name: %v", ctl.srcName)
				} else {
					player.streams[ctl.streamId] = &stream{
						srcName:    ctl.srcName,
						streamId:   ctl.streamId,
						playing:    true,
						paused:     !ctl.play,
						readAt:     0,
						streamType: typeInterval,
						interval:   ctl.interval,
					}
				}
			case *addGroupStream:
				ctl := e.(*addGroupStream)

				filtered_src := make([]string, 0, len(ctl.src))
				for _, v := range ctl.src {
					if _, ok := player.sources[v]; !ok {
						log.Printf("Audio: Group: Source is not in the map - Name: %v", v)
					} else {
						filtered_src = append(filtered_src, v)
					}
				}
				if _, ok := player.streams[ctl.streamId]; ok {
					log.Printf("Audio: Group: stream exists - skipping: %v", ctl.streamId)
				} else if len(filtered_src) < 1 {
					log.Printf("Audio: Group: empty list - skipping: %v", ctl.streamId)
				} else {

					player.streams[ctl.streamId] = &stream{
						src:        filtered_src,
						streamId:   ctl.streamId,
						playing:    false,
						paused:     !ctl.play,
						readAt:     0,
						streamType: typeGroup,
						interval:   ctl.interval,
					}
					if ctl.play {
						rd := random.Intn(int(len(filtered_src)))
						if _, ok := player.sources[filtered_src[rd]]; !ok {
							log.Printf("Audio: Group: Source is not in the map - Name: %v", filtered_src[rd])
						} else {
							player.streams[ctl.streamId].srcName = filtered_src[rd]
							player.streams[ctl.streamId].playing = true
							player.streams[ctl.streamId].lastGroupRd = rd
						}
					}
				}
			default:
				log.Printf("Audio: in_chann: Unknown type: %T", e)
			}
		}
		// Mixer
		// Mix multiple streams into single stream and write result to output buffer
		if player.state && player.play && write_avail {
			samples_to_read := (params.buffBytes - len(mix_buff)) / params.Channels
			if samples_to_read > 0 {
				// check and start rand and interval streams if it is time
				for _, stream := range player.streams {
					if stream.streamType == typeInterval || stream.streamType == typeRand || stream.streamType == typeGroup {
						if !stream.paused && !stream.playing && stream.startTime.Before(time.Now()) {
							stream.playing = true
						}
					}
				}
				for i := 0; i < samples_to_read && len(player.streams) > 0; i += params.Channels {
					// HARDCODED to 16 bit audio
					var f0, f1 uint16
					frame_has_val := false
					// TODO: optimize loops and mapaccess
					for key, stream := range player.streams {
						if stream.playing {
							// stream ended: restart or delete stream
							if player.sources[stream.srcName].length-stream.readAt < params.Channels {
								if stream.deleted {
									delete(player.streams, key) //safe to delete key in range
									continue
								} else if stream.loop {
									stream.readAt = 0
								} else if stream.streamType == typeInterval {
									stream.readAt = 0
									stream.playing = false
									st := time.Now().Add(time.Duration(stream.interval) * time.Millisecond)
									stream.startTime = st
								} else if stream.streamType == typeGroup {
									stream.readAt = 0
									stream.playing = false
									stream.srcName = ""
									rd := random.Intn(int(len(stream.src)))
									//try not to repeat
									if stream.lastGroupRd == rd {
										rd = random.Intn(int(len(stream.src)))
									}
									if stream.lastGroupRd == rd {
										rd = random.Intn(int(len(stream.src)))
									}
									if _, ok := player.sources[stream.src[rd]]; !ok {
										log.Printf("Audio: Group: Source is not in the map - Name: %v", stream.src[rd])
									} else {
										stream.srcName = stream.src[rd]
										stream.lastGroupRd = rd
										st := time.Now().Add(time.Duration(stream.interval) * time.Millisecond)
										stream.startTime = st
									}
								} else if stream.streamType == typeRand {
									stream.readAt = 0
									stream.playing = false
									rd := random.Intn(int(stream.max-stream.min)) + int(stream.min)
									st := time.Now().Add(time.Duration(rd) * time.Millisecond)
									stream.startTime = st
								} else {
									delete(player.streams, key) //safe to delete key in range
									continue
								}
							}
							// mix streams and normalize audio
							// TODO: optimize math
							if !player.mute {
								b0 := uint16(player.sources[stream.srcName].Buff[stream.readAt+0])
								f0 = (f0 + b0) - ((f0 * b0) / 65535)
								b1 := uint16(player.sources[stream.srcName].Buff[stream.readAt+1])
								f1 = (f1 + b1) - ((f1 * b1) / 65535)
								frame_has_val = true
							}
							stream.readAt += params.Channels
						}
					}
					// convert uint16 to byte
					// empty frames will be skipped from buffer!!!
					if frame_has_val {
						//HARDCODED to 2 channel 16 bit audio
						//for skey, a := range frame {}
						l1 := byte(f0)
						l2 := byte(f0 >> 8)
						r1 := byte(f1)
						r2 := byte(f1 >> 8)
						mix_buff = append(mix_buff, l1, l2, r1, r2)
					}
				}
			}

			// write to output buffer
			n := len(mix_buff)
			if n > 0 {
				var cnt int // how many bytes are written to output
				//write_avail = false
				player.state, cnt, write_avail = writeBuff(handle, mix_buff[:n], params)

				rbs += cnt
				if rbs+params.buffBytes > rbc {
					// if not enough space to the end of ring buffer move to the begining
					rbs = 0
					tmp_slice := mix_buff[cnt:n]
					mix_buff = ring_buff[rbs:0:params.buffBytes]
					mix_buff = append(mix_buff, tmp_slice...)
				} else {
					// advance in the ring buffer
					mix_buff = ring_buff[rbs : rbs+(n-cnt) : rbs+params.buffBytes]
				}

			}
		}
		time.Sleep(time.Millisecond * 5)
	}
}
