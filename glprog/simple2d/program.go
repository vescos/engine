package simple2d

import (
	"graphs/engine/gles2"
	"graphs/engine/glprog"
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

func Program() *glprog.Prog {
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
		Uniforms: map[string]*glprog.Uniform{
			"texture": {
				Name: "u_texture",
				Fn:   "uniform1i",
			},
		},
	}

}
