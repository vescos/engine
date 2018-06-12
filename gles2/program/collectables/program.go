package collectables

import (
	"github.com/vescos/engine/gles2/gl"
	"github.com/vescos/engine/gles2/program"
)

const vs = `
	#version 100
	
	attribute vec3 a_vertex;
	attribute vec3 a_normal_vector; 
	attribute vec3 a_color;
	attribute vec3 a_translate;

	uniform mat4 u_merged_matrix;
	uniform mat4 u_animate_matrix;
	uniform vec3 u_ambient;
	uniform vec3 u_directional;
	uniform vec3 u_directional_vector;

	varying vec3 v_color;

	void main(void) {
		mat4 translate_m = mat4(
			vec4(1.0, 0.0, 0.0, 0.0),
			vec4(0.0, 1.0, 0.0, 0.0),
			vec4(0.0, 0.0, 1.0, 0.0),
			vec4(a_translate.x, a_translate.y, a_translate.z, 1.0)
		);
		vec3 rotated_normal = (u_animate_matrix * vec4(a_normal_vector, 1.0)).xyz;
		gl_Position = u_merged_matrix * translate_m * u_animate_matrix * vec4(a_vertex, 1.0);
		float directional = max(dot(rotated_normal, u_directional_vector), 0.0);
		v_color = a_color * (u_ambient + (u_directional * directional));
	}
`
const fs = `
	#version 100
	precision mediump float;

	varying vec3 v_color;

	void main(void) {
		gl_FragColor = vec4(v_color, 1.0);
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
				Usage:  gl.STATIC_DRAW,
				Target: gl.ARRAY_BUFFER,
				TypeGl: gl.FLOAT,
				Attribs: map[string]*program.Attrib{
					"vertices": {
						Name:       "a_vertex",
						Size:       3,
						Normalized: false,
						Stride:     12 * 4,
						Offset:     0,
					},
					"normals": {
						Name:       "a_normal_vector",
						Size:       3,
						Normalized: false,
						Stride:     12 * 4,
						Offset:     3 * 4,
					},
					"color": {
						Name:       "a_color",
						Size:       3,
						Normalized: false,
						Stride:     12 * 4,
						Offset:     6 * 4,
					},
					"translate": {
						Name:       "a_translate",
						Size:       3,
						Normalized: false,
						Stride:     12 * 4,
						Offset:     9 * 4,
					},
				},
			},
		},
		Uniforms: map[string]*program.Uniform{
			"mergedMatrix": {
				Name: "u_merged_matrix",
				Fn:   "uniformMatrix4fv",
			},
			"animateMatrix": {
				Name: "u_animate_matrix",
				Fn:   "uniformMatrix4fv",
			},
			"ambient": {
				Name: "u_ambient",
				Fn:   "uniform3f",
			},
			"directional": {
				Name: "u_directional",
				Fn:   "uniform3f",
			},
			"directionalVector": {
				Name: "u_directional_vector",
				Fn:   "uniform3f",
			},
		},
	}
}
