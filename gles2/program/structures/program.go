package structures

import (
	"github.com/vescos/engine/gles2/gl"
	"github.com/vescos/engine/gles2/program"
)

const vs = `
	#version 100
		
	attribute vec3 a_vertex;
	attribute vec2 a_texture_coord;
	attribute mat3 a_tspace_matrix;

	uniform mat4 u_merged_matrix;
	uniform mat4 u_shadowmap_matrix;
	uniform vec3 u_diffuse_vector;
		
	varying vec2 v_texture_coord;
	// light vector in tangent space
	varying vec3 v_light_tspace;
	varying vec4 v_shadow_coord;
		  
	void main(void) {
		v_texture_coord = a_texture_coord;
		v_light_tspace = normalize(u_diffuse_vector * a_tspace_matrix);
		v_shadow_coord = u_shadowmap_matrix * vec4(a_vertex, 1.0);
		gl_Position = u_merged_matrix * vec4(a_vertex, 1.0);
	}
`
const fs = `
	#version 100

	precision mediump float;

	uniform vec3 u_ambient;
	uniform vec3 u_diffuse; // light's diffuse color
	uniform sampler2D u_texture;
	uniform sampler2D u_texture_normal;
	uniform sampler2D u_shadow_map;

	varying vec2 v_texture_coord;
	varying vec3 v_light_tspace;
	varying vec4 v_shadow_coord;

	//diffuse = max(l.n, 0) * diffuse_light * diffuse_material
	//l - v_l - light vector in tangent space
	//n - per pixel normal vector in tangent space
	void main(void) {
		// shadow
		float distance = texture2DProj(u_shadow_map, v_shadow_coord).r;
		
		float shadow = 1.0;
	 	if (v_shadow_coord.w > 0.0) {
	 		shadow = distance < (v_shadow_coord.z / v_shadow_coord.w) ? 0.3 : 1.0;
		}
		// the material's diffuse color
		vec3  diffuse_material = texture2D(u_texture, v_texture_coord.st).rgb;
		vec4 nc = texture2D(u_texture_normal, v_texture_coord.st); 
		// per pixel normal vector in tangent space
		vec3 n = (nc.rgb * 2.0) - 1.0;
		vec3 color = (u_ambient + shadow * max(dot(v_light_tspace, n), 0.0) * u_diffuse) * diffuse_material;
		gl_FragColor = vec4(color, 1.0);
	}
`

func Program() *program.Prog {
	return &program.Prog{
		Vs:          vs,
		Fs:          fs,
		Mode:        gl.TRIANGLES,
		Length:      0,
		UseMstUnits: true,
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
					"tspaceMatrix": {
						Name:       "a_tspace_matrix",
						Size:       9,
						Normalized: false,
						Stride:     14 * 4,
						Offset:     5 * 4,
						MatSize:	3,
					},
				},
			},
		},
		Uniforms: map[string]*program.Uniform{
			"mergedMatrix": {
				Name: "u_merged_matrix",
				Fn:   "uniformMatrix4fv",
			},
			"shadowMatrix": {
				Name: "u_shadowmap_matrix",
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
			"shadow_map": {
				Name: "u_shadow_map",
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
