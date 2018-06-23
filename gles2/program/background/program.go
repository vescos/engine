package background

import (
	"github.com/vescos/engine/gles2/gl"
	"github.com/vescos/engine/gles2/program"
)

const vs = `
	#version 100
	 
	attribute vec3 a_vertex;

	uniform mat4 u_rotation_matrix;

	varying vec3 v_texture_coord;
	  
	void main(void) {
		v_texture_coord = a_vertex;
		vec4 pos = u_rotation_matrix * vec4(a_vertex, 1.0);
		gl_Position = pos.xyww;
	}
`
const fs = `
	#version 100

	precision mediump float;

	uniform samplerCube u_texture;

	varying vec3 v_texture_coord;

	void main(void) {
		gl_FragColor = textureCube(u_texture, v_texture_coord.stp);
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
						Size:       3,
						Normalized: false,
						Stride:     0,
						Offset:     0,
					},
				},
			},
		},
		Uniforms: map[string]*program.Uniform{
			"rotationMatrix": {
				Name: "u_rotation_matrix",
				Fn:   "uniformMatrix4fv",
			},
			"texture": {
				Name: "u_texture",
				Fn:   "uniform1i",
			},
		},
	}
}
