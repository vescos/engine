package simple2d

import (
	"github.com/vescos/engine/gles2/gl"
	"github.com/vescos/engine/gles2/program"
)

const vs = `
	#version 100
		
	attribute vec2 a_vertex;
	attribute vec3 a_texture_coord; //last component is alpha coefficient

	varying vec3 v_texture_coord;
		  
	void main(void) {
		gl_Position =  vec4(a_vertex, 0.0, 1.0);
		v_texture_coord = a_texture_coord;
	}
`
const fs = `
	#version 100
	precision mediump float;

	uniform sampler2D u_texture;

	varying vec3 v_texture_coord;

	void main(void) {
		vec4 color = texture2D(u_texture, v_texture_coord.st);
		gl_FragColor = vec4(color.rgb, color.a * v_texture_coord.p);
	}
`

func Program() *program.Prog {
	return &program.Prog{
		Vs:     vs,
		Fs:     fs,
		Mode:   gl.TRIANGLES,
		Length: 0,
		Buffs: map[string]*program.Buff{
			"elements": {
				Target: gl.ELEMENT_ARRAY_BUFFER,
				TypeGl: gl.UNSIGNED_SHORT,
				Usage:  gl.STATIC_DRAW,
			},
			"allFloats": {
				Target: gl.ARRAY_BUFFER,
				TypeGl: gl.FLOAT,
				Usage:  gl.STATIC_DRAW,
				Attribs: map[string]*program.Attrib{
					"vertices": {
						Name:       "a_vertex",
						Size:       2,
						Normalized: false,
						Stride:     5 * 4,
						Offset:     0,
					},
					"textureCoord": {
						Name:       "a_texture_coord",
						Size:       3,
						Normalized: false,
						Stride:     5 * 4,
						Offset:     2 * 4,
					},
				},
			},
		},
		Uniforms: map[string]*program.Uniform{
			"texture": {
				Name: "u_texture",
				Fn:   "uniform1i",
			},
		},
	}

}
