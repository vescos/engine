// gl textures
// gl functions must be called only from main(goroutine that bound context)

package texture

import (
	"image"

	"github.com/vescos/engine/gles2/gl"
)

const (
	ETC1ImgFormat = gl.Enum(0x8D64) //ETC1_RGB8_OES 0x8D64
)

type ImageTextures struct {
	Images  []*image.RGBA
	Format  gl.Enum `json:"Format"`
	FileExt string     `json:"FileExt"`
	GenMips bool       `json:"GenMips"`
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
	Sources       []string     `json:"sources"`
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

func (t *Texture) SetImage(i int, img *image.RGBA) {
	t.Images.Images[i] = img
}

func (t *Texture) Set(tex gl.Texture) {
	t.Texture = tex
}


func Build2DTexture(t *Texture, ctMaxTexureU int) gl.Texture {
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

func BuildCubeMapTexture(t *Texture, ctMaxTexureU int) gl.Texture {
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
