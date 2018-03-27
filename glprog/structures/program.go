package structures

import (
	"graphs/engine/gles2"
	"graphs/engine/glprog"
)

const vs = `
	#version 100
		
	attribute vec3 a_vertex;
	attribute vec2 a_texture_coord;
	attribute vec3 a_normal_vector;
	attribute vec3 a_tangent;
	attribute vec3 a_bitangent;

	uniform mat4 u_merged_matrix;
	uniform vec3 u_diffuse_vector;
		
	varying vec2 v_texture_coord;
	//light vector in tangent space
	varying vec3 v_l;
		  
	void main(void) {
		gl_Position = u_merged_matrix * vec4(a_vertex, 1.0);
		v_texture_coord = a_texture_coord;
		mat3 t_space = mat3(a_tangent, a_bitangent, a_normal_vector);
		v_l = normalize(u_diffuse_vector * t_space);
	}
`
const fs = `
	#version 100

	precision mediump float;

	uniform vec3 u_ambient;
	uniform vec3 u_diffuse;//light's diffuse color
	uniform sampler2D u_texture;
	uniform sampler2D u_texture_normal;

	varying vec2 v_texture_coord;
	//light vector in tangent space
	varying vec3 v_l;

	//diffuse = max(l.n, 0) * diffuse_light * diffuse_material
	//l - v_l - light vector in tangent space
	//n - per pixel normal vector in tangent space
	void main(void) {
		//the material's diffuse color
		vec3  diffuse_material = texture2D(u_texture, v_texture_coord.st).rgb;
		vec4 nc = texture2D(u_texture_normal, v_texture_coord.st); 
		//per pixel normal vector in tangent space
		vec3 n = nc.rgb * 2.0 - 1.0;
		//FIXME: Is implementation of ambient correct?
		vec3 diffuse = (u_ambient + max(dot(v_l, n), 0.0) * u_diffuse) * diffuse_material;
		gl_FragColor = vec4(diffuse, 1.0);
	}
`

func Program () *glprog.Prog {
	return &glprog.Prog {
		Vs:          vs,
		Fs:          fs,
		Mode:        gles2.TRIANGLES,
		Length:      0,
		UseMstUnits: true,
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
						Stride:     14 * 4,
						Offset:     0,
					},
					"textureCoord": {
						Name:       "a_texture_coord",
						Size:       2,
						Normalized: false,
						Stride:     14 * 4,
						Offset:     3 * 4,
					},
					"normals": {
						Name:       "a_normal_vector",
						Size:       3,
						Normalized: false,
						Stride:     14 * 4,
						Offset:     5 * 4,
					},
					"tangents": {
						Name:       "a_tangent",
						Size:       3,
						Normalized: false,
						Stride:     14 * 4,
						Offset:     8 * 4,
					},
					"bitangents": {
						Name:       "a_bitangent",
						Size:       3,
						Normalized: false,
						Stride:     14 * 4,
						Offset:     11 * 4,
					},
				},
			},
		},
		Uniforms: map[string]*glprog.Uniform{
			"mergedMatrix": {
				Name: "u_merged_matrix",
				Fn:   "uniformMatrix4fv",
			},
			//hardcoded in gles2DrawObject
			"texture": {
				Name: "u_texture",
				Fn:   "uniform1i",
			},
			"texture_normal": {
				Name: "u_texture_normal",
				Fn:   "uniform1i",
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
