/*
 * Read JSON data from stdin, serialize this data as gob and save in a file
 * Usage from bash
 * node node_exporter.js gname export| gobencoder /full/filename.blah export
 * gname - thetower|vertical|cubes
 * export - meshes|textures|gems
 */

package main

import (
	"encoding/gob"
	"encoding/json"
	"image"
	"io"
	"log"
	"os"

	"graphs/ext/gl"
	"graphs/tobyte"
)

//TODO: move all type definitions in separate package because of duplication
type Floor struct {
	C []string  `json:"c"`
	T []float32 `json:"t"`
	Y float32   `json:"y"`
}

type SectionIn struct {
	B   []float32         `json:"b"`
	Cnt [32]int32         `json:"cnt"`
	F   map[string]*Floor `json:"f"`
	I   []uint16          `json:"i"`
	S   [3]float32        `json:"s"`
	SK  string            `json:"sk"`
	AI  []uint16          `json:"ai"`
	AB  []float32         `json:"ab"`
}

type Section struct {
	B   []byte            `json:"b"`
	Cnt [32]int32         `json:"cnt"`
	F   map[string]*Floor `json:"f"`
	I   []byte            `json:"i"`
	S   [3]float32        `json:"s"`
	SK  string            `json:"sk"`
	AI  []byte            `json:"ai"`
	AB  []byte            `json:"ab"`
}

type ImageTextures struct {
	Images  []*image.RGBA
	Format  gl.Enum `json:"Format"`
	FileExt string  `json:"FileExt"`
	GenMips bool    `json:"GenMips"`
}

type MipmapETC1 struct {
	W      int
	H      int
	Level  int
	Mipmap []byte
	Target gl.Enum
}

type CompressedETC1Textures struct {
	Available    bool `json:"Available"`
	Mipmaps      []*MipmapETC1
	MipmapLevels int `json:"MipmapLevels"`
}

type TextureGL struct {
	Target        gl.Enum   `json:"target"`
	Sources       []string  `json:"sources"`
	Texturemap    []gl.Enum `json:"texturemap"`
	Unit          gl.Enum   `json:"unit"`
	Texture       gl.Texture
	WrapS         int `json:"wrap_s"`
	WrapT         int `json:"wrap_t"`
	MagFilter     int `json:"MagFilter"`
	MinFilter     int `json:"MinFilter"`
	Disabled      bool
	Images        *ImageTextures          `json:"Img"`
	ComprETC1     *CompressedETC1Textures `json:"Compr"`
	UseCompressed bool
	AtlassMap     map[string][8]float32 `json:"atlassMap"`
}

func encodeSections(dec *json.Decoder, enc *gob.Encoder) {
	for {
		var s SectionIn
		var sout Section
		err := dec.Decode(&s)
		if err != nil {
			if err != io.EOF {
				log.Print("encodeSections: Can't parse json: ", err)
			}
			break
		}
		//convert buffers to byte
		sout.B = tobyte.Float32Le(s.B)
		sout.I = tobyte.Uint16Le(s.I)
		sout.AB = tobyte.Float32Le(s.AB)
		sout.AI = tobyte.Uint16Le(s.AI)
		sout.Cnt = s.Cnt
		sout.F = s.F
		sout.S = s.S
		sout.SK = s.SK
		enc.Encode(sout)
	}
}

func encodeTextures(dec *json.Decoder, enc *gob.Encoder) {
	for {
		var t map[string]*TextureGL
		err := dec.Decode(&t)
		if err != nil {
			if err != io.EOF {
				log.Print("encodeTextures: Can't parse json: ", err)
			} else {
				log.Print("Textures exported.")
			}
			break
		}
		enc.Encode(t)
	}
}

type gemProps struct {
	Visible_from []string  `json:"visible_from"`
	Xyz          []float32 `json:"xyz"`
	XyzO         []float32
	Collected    bool `json:"collected"`
}

func encodeGems(dec *json.Decoder, enc *gob.Encoder) {
	for {
		var t map[string]*gemProps
		err := dec.Decode(&t)
		if err != nil {
			if err != io.EOF {
				log.Print("encodeGems: Can't parse json: ", err)
			} else {
				log.Print("Gems exported.")
			}
			break
		}
		enc.Encode(t)
	}
}

func main() {
	//check that stdin is connected to pipe
	stat, _ := os.Stdin.Stat()
	var usage = `Usage: node node_exporter.js gname [export]| gobencoder /full/filename.blah [export]
        gname - thetower|vertical|cubes
        export - meshes|textures|gems
    `
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		log.Print("stdin is from a terminal!!! " + usage)
		return
	}
	if len(os.Args) <= 1 {
		log.Print("/full/filename argument is required and writable")
		return
	}
	fn := os.Args[1]
	export := "meshes"
	if len(os.Args) > 2 {
		export = os.Args[2]
	}
	dec := json.NewDecoder(os.Stdin)
	fd, err := os.Create(fn)
	if err != nil {
		log.Printf("Can't create file: %v, Error: %v", fn, err)
		return
	}
	defer fd.Close()
	enc := gob.NewEncoder(fd)

	if export == "textures" {
		encodeTextures(dec, enc)
	} else if export == "gems" {
		encodeGems(dec, enc)
	} else {
		encodeSections(dec, enc)
	}
}
