package shadow

import (
	"github.com/vescos/engine/gles2/gl"
	"github.com/vescos/engine/gles2/program"
)

const vs = `
	#version 100
		
	attribute vec3 a_vertex;

	uniform mat4 u_shadow_matrix;
		
	void main(void) {
		gl_Position = u_shadow_matrix * vec4(a_vertex, 1.0);
	}
`
const fs = `
	#version 100

	void main(void) {}
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
				},
			},
		},
		Uniforms: map[string]*program.Uniform{
			"shadowMatrix": {
				Name: "u_shadow_matrix",
				Fn:   "uniformMatrix4fv",
			},
		},
	}

}
