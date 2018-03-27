package background

import (
	"graphs/engine/gles2"
	"graphs/engine/glprog"
)

const vs = `
	#version 100
	 
	attribute vec3 a_vertex;

	uniform mat4 u_rotation_matrix;

	varying vec3 v_texture_coord;
	  
	void main(void) {
		gl_Position = u_rotation_matrix * vec4(a_vertex, 1.0);
		v_texture_coord = a_vertex;
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

func Program () *glprog.Prog {
	return &glprog.Prog{
		Vs:     vs,
		Fs:     fs,
		Mode:   gles2.TRIANGLES,
		Length: 0,
		Buffs: map[string]*glprog.Buff{
			"elements": {
				Target: gles2.ELEMENT_ARRAY_BUFFER,
				TypeGl: gles2.UNSIGNED_SHORT,
				Usage:  gles2.STATIC_DRAW,
			},
			"allFloats": {
				Target: gles2.ARRAY_BUFFER,
				TypeGl: gles2.FLOAT,
				Usage:  gles2.STATIC_DRAW,
				Attribs: map[string]*glprog.Attrib{
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
		Uniforms: map[string]*glprog.Uniform{
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
