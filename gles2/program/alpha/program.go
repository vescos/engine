package alpha

import (
	"github.com/vescos/engine/gles2/gl"
	"github.com/vescos/engine/gles2/program"
)

const vs = `
	#version 100
	
	attribute vec3 a_vertex;
	attribute vec3 a_normal_vector;
	attribute vec2 a_texture_coord;
	
	uniform mat4 u_merged_matrix;
	uniform vec3 u_ambient;
	uniform vec3 u_diffuse;
	uniform vec3 u_diffuse_vector;

	varying vec2 v_texture_coord;
	varying vec3 v_light_value;
	  
	void main(void) {
		gl_Position =  u_merged_matrix * vec4(a_vertex, 1.0);
		v_texture_coord = a_texture_coord;
		float directional = max(dot(a_normal_vector, u_diffuse_vector), 0.0);
		v_light_value = u_ambient + (u_diffuse * directional);
	}
`

const fs = `
	#version 100
	precision mediump float;
	
	uniform sampler2D u_texture;
	
	varying vec2 v_texture_coord;
	varying vec3 v_light_value;
	
	void main(void) {
		vec4 color = texture2D(u_texture, v_texture_coord.st);
		if (color.a < 0.5) {
			discard;
		}
		gl_FragColor = vec4(color.rgb * v_light_value, color.a);
	}
`

func Program() *program.Prog {
	return &program.Prog{
		Vs:          vs,
		Fs:          fs,
		Mode:        gl.TRIANGLES,
		Length:      0,
		UseMstUnits: false,
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
						Stride:     8 * 4,
						Offset:     0,
					},
					"textureCoord": {
						Name:       "a_texture_coord",
						Size:       2,
						Normalized: false,
						Stride:     8 * 4,
						Offset:     3 * 4,
					},
					"normals": {
						Name:       "a_normal_vector",
						Size:       3,
						Normalized: false,
						Stride:     8 * 4,
						Offset:     5 * 4,
					},
				},
			},
		},
		Uniforms: map[string]*program.Uniform{
			"texture": {
				Name: "u_texture",
				Fn:   "uniform1i",
			},
			"mergedMatrix": {
				Name: "u_merged_matrix",
				Fn:   "uniformMatrix4fv",
			},
			"ambient": {
				Name: "u_ambient",
				Fn:   "uniform3f",
			},
			"directional": {
				Name: "u_diffuse",
				Fn:   "uniform3f",
			},
			"directionalVector": {
				Name: "u_diffuse_vector",
				Fn:   "uniform3f",
			},
		},
	}

}
