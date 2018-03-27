//Veselin Kostov Jan 2017

//TODO: Hardcoded to 2 chanells, litle Endian, WAVE PCM audio
package wav

import (
	"bytes"
	"io"
	"log"

	"graphs/engine/assets"
	"graphs/engine/tobyte"
)

type State struct{}

//?? signed or unsigned fileds?
type WavHeader struct {
	riff          string
	endian        string
	format        string
	audioFormat   uint16
	pcm           bool //true if audioFormat == 1
	numChannels   uint16
	sampleRate    uint32
	bitsPerSample uint16
	dataSize      uint32
}

func (s *State) Decode(file assets.Asset) []int16 {
	var buff bytes.Buffer
	hs := 44
	header := make([]byte, hs, hs)
	wh := WavHeader{}
	cnt, err := file.Read(header)
	if cnt < hs || err != nil {
		log.Printf("Decoder:wav: can't read header error: %v, filename: %v", err, file.Name())
		return nil
	}
	wh.riff = string(header[0:4])
	if wh.riff == "RIFF" {
		wh.endian = "litleEndian"
	} else if wh.riff == "RIFX" {
		wh.endian = "bigEndian"
		//TODO: implement bigEndian
		log.Printf("Decoder:wav: RIFF bigEndian not supported. filename: %v, riff: %v", file.Name(), wh.riff)
		return nil
	} else {
		log.Printf("Decoder:wav: unknown RIFF value. filename: %v, RIFF: %v", file.Name(), wh.riff)
		return nil
	}
	wh.format = string(header[8:12])
	if wh.format != "WAVE" {
		log.Printf("Decoder:wav: not WAVE format. filename: %v, format: %v", file.Name(), wh.format)
		return nil
	}
	wh.audioFormat = uint16(header[20]) | uint16(header[21])<<8
	if wh.audioFormat != 1 {
		log.Printf("Decoder:wav: PCM != 1, PCM only supported. filename: %v, PCM: %v", file.Name(), wh.audioFormat)
		return nil
	}
	wh.pcm = true
	wh.numChannels = uint16(header[22]) | uint16(header[23])<<8
	if wh.numChannels != 2 {
		log.Printf("Decoder:wav: 2 channels only supported. filename: %v, channels: %v", file.Name(), wh.audioFormat)
		return nil
	}
	wh.sampleRate = uint32(header[24]) | uint32(header[25])<<8 | uint32(header[26])<<16 | uint32(header[27])<<24
	wh.bitsPerSample = uint16(header[34]) | uint16(header[35])<<8
	wh.dataSize = uint32(header[40]) | uint32(header[41])<<8 | uint32(header[42])<<16 | uint32(header[43])<<24

	_, err = buff.ReadFrom(file)
	if err != nil && err != io.EOF {
		log.Printf("Decoder:wav: read error: %v, filename: %v", err, file.Name())
		return nil
	}

	b := buff.Bytes()[:wh.dataSize]
	//log.Printf("%v", wh)
	return tobyte.ByteLeToInt16(b)
}
