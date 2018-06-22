// gl textures
// TODO: this is incomplete(refactoring needed)

// gl functions must be called only from main(goroutine that bound context) and after context is bound!!!

package texture

import (
	"bytes"
	"encoding/gob"
	"image"
	"image/draw"
	//_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"strconv"

	"github.com/vescos/engine/assets"
	"github.com/vescos/engine/gles2/gl"
)

const (
	ETC1ImgFormat = gl.Enum(0x8D64) //ETC1_RGB8_OES 0x8D64
)

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

type Texture struct {
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

func Build2D(t *Texture, ctMaxTexureU int) gl.Texture {
	var texture gl.Texture
	if int(t.Unit) > ctMaxTexureU-1 {
		return texture
	}
	gl.ActiveTexture(gl.TEXTURE0 + t.Unit)
	texture = gl.CreateTexture()
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.WrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.WrapT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.MagFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.MinFilter)
	if t.UseCompressed {
		for _, v := range t.ComprETC1.Mipmaps {
			if v != nil {
				gl.CompressedTexImage2D(gl.TEXTURE_2D, v.Level, ETC1ImgFormat, v.W, v.H, 0, v.Mipmap)
			}
		}
	} else {
		b := t.Images.Images[0].Bounds()
		gl.TexImage2D(gl.TEXTURE_2D, 0, b.Dx(), b.Dy(), t.Images.Format, gl.UNSIGNED_BYTE, t.Images.Images[0].Pix)
		if t.Images.GenMips {
			gl.GenerateMipmap(gl.TEXTURE_2D)
		}
	}
	return texture
}

func BuildCubeMap(t *Texture, ctMaxTexureU int) gl.Texture {
	var texture gl.Texture
	if int(t.Unit) > ctMaxTexureU-1 {
		return texture
	}
	gl.ActiveTexture(gl.TEXTURE0 + t.Unit)
	texture = gl.CreateTexture()
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, texture)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, t.MagFilter)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, t.MinFilter)
	//_MIPMAP_ filters don't wokr when mipmaps are not available!!!
	if t.UseCompressed {
		for _, v := range t.ComprETC1.Mipmaps {
			if v != nil {
				gl.CompressedTexImage2D(v.Target, v.Level, ETC1ImgFormat, v.W, v.H, 0, v.Mipmap)
			}
		}
	} else {
		for i := range t.Images.Images {
			b := t.Images.Images[i].Bounds()
			gl.TexImage2D(t.Texturemap[i], 0, b.Dx(), b.Dy(), t.Images.Format, gl.UNSIGNED_BYTE, t.Images.Images[i].Pix)
		}
		if t.Images.GenMips {
			gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP)
		}
	}
	return texture
}

func Png(t *Texture, am assets.FileManager) {
	for i := range t.Sources {
		fname := t.Sources[i] + "." + t.Images.FileExt
		file, err := am.OpenAsset(fname)
		if err == nil {
			defer file.Close()
			img, _, err := image.Decode(file)
			if err == nil {
				b := img.Bounds()
				newImg := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
				draw.Draw(newImg, newImg.Bounds(), img, b.Min, draw.Src)
				t.Images.Images[i] = newImg
				img = nil
			} else {
				log.Print("texturesLoadFiles: ", err, ": ", fname)
			}
		} else {
			log.Print("texturesLoadFiles", err, ": ", fname)
		}
	}
}

func Etc1(t *Texture, am assets.FileManager, startAt int) {
	index := 0
	for i := range t.Sources {
		for j := startAt; j < t.ComprETC1.MipmapLevels; j += 1 {
			fname := t.Sources[i] + "_mip_" + strconv.Itoa(j) + ".pkm"
			file, err := am.OpenAsset(fname)
			if err == nil {
				defer file.Close()
				var buff bytes.Buffer
				_, err := buff.ReadFrom(file)
				if err != nil && err != io.EOF {
					log.Print("texturesLoadCompressedFile", err, ": ", fname)
					return
				}

				img := buff.Bytes()
				wb := img[12:14]
				hb := img[14:16]
				var tmap gl.Enum = gl.TEXTURE_2D
				if t.Target == gl.TEXTURE_CUBE_MAP {
					tmap = t.Texturemap[i]
				}
				t.ComprETC1.Mipmaps[index] = &MipmapETC1{
					Level:  j - startAt,
					Target: tmap,
					W:      int(wb[0])*256 + int(wb[1]),
					H:      int(hb[0])*256 + int(hb[1]),
					Mipmap: img[16:],
				}
			} else {
				log.Print("texturesLoadCompressedFile", err, ": ", fname)
			}
			index += 1
		}
	}
}

func DescrFromGob(fname string, am assets.FileManager) (map[string]*Texture) {
	var t map[string]*Texture
	file, err := am.OpenAsset(fname)
	if err == nil {
		defer file.Close()
		dec := gob.NewDecoder(file)
		err := dec.Decode(&t)
		if err != nil {
			if err != io.EOF {
				log.Print("texturesSet: ", err)
			}
		}
	} else {
		log.Print("texturesSet: ", err)
	}
	//TODO: make mipmaps and images
	for key := range t {
		txtr := t[key]
		n := 1
		if txtr.Target == gl.TEXTURE_CUBE_MAP {
			n = 6
		}
		txtr.Images.Images = make([]*image.RGBA, n)
		if txtr.ComprETC1.Available {
			txtr.ComprETC1.Mipmaps = make([]*MipmapETC1, n*txtr.ComprETC1.MipmapLevels)
		}
		t[key] = txtr
	}
	return t
}

// Load all texures in set to GPU
func LoadAll(txs map[string]*Texture, am assets.FileManager, maxTextureUnits int, startAt int) {
	for _, t := range txs {
		Load(t, am, maxTextureUnits, startAt)
	}
}

// Load single texture to GPU
func Load(t *Texture, am assets.FileManager, maxTextureUnits int, startAt int) {
	if t.Disabled {
		return
	}
	if t.ComprETC1 != nil && t.ComprETC1.Available {
		t.UseCompressed = true
		Etc1(t, am, startAt)
	} else {
		t.UseCompressed = false
		Png(t, am)
	}
	if t.Target == gl.TEXTURE_2D {
		t.Texture = Build2D(t, maxTextureUnits)
	} else if t.Target == gl.TEXTURE_CUBE_MAP {
		t.Texture = BuildCubeMap(t, maxTextureUnits)
	} else {
		log.Print("Target not Supported: ", t.Target)
	}
}
