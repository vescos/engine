package program

import (
	"log"

	"github.com/vescos/engine/gles2/gl"
)

//DONE: optimization: USE Interleaved Vertex Data
//https://developer.apple.com/library/ios/documentation/3DDrawing/Conceptual/OpenGLES_ProgrammingGuide/TechniquesforWorkingwithVertexData/TechniquesforWorkingwithVertexData.html
//https://en.wikibooks.org/wiki/OpenGL_Programming/Modern_OpenGL_Tutorial_03

type Uniform struct {
	Name     string
	Location gl.Uniform
	Fn       string
	V0       interface{}
	V1       interface{}
	V2       interface{}
	V3       interface{}
}

type Attrib struct {
	Name       string
	Size       int
	Normalized bool
	Stride     int
	Offset     int
	MatSize int // if attribute is matrix 
}

type Buff struct {
	Buffer  gl.Buffer
	Target  gl.Enum
	TypeGl  gl.Enum
	Usage   gl.Enum
	Attribs map[string]*Attrib
}

type Prog struct {
	P           gl.Program
	Vs          string
	Fs          string
	Mode        gl.Enum
	Length      int
	Buffs       map[string]*Buff
	Uniforms    map[string]*Uniform
	UseMstUnits bool
}

func (p *Prog) SetBuffer(key string, buffer []byte) {
	if p.Buffs[key].Buffer.Value == 0 {
		p.Buffs[key].Buffer = gl.CreateBuffer()
	}
	gl.BindBuffer(p.Buffs[key].Target, p.Buffs[key].Buffer)
	gl.BufferData(p.Buffs[key].Target, buffer, p.Buffs[key].Usage)
}

//create and link shader program
func (p *Prog) BuildProgram() {
	var (
		program gl.Program
		vshader gl.Shader
		fshader gl.Shader
	)
	program = gl.CreateProgram()
	//vertex shader
	vshader = gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(vshader, p.Vs)
	gl.CompileShader(vshader)
	if gl.GetShaderi(vshader, gl.COMPILE_STATUS) == gl.FALSE {
		log.Printf("glprog: VS compilation failed: %v", gl.GetShaderInfoLog(vshader))
	}
	//fragment shader
	fshader = gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(fshader, p.Fs)
	gl.CompileShader(fshader)
	if gl.GetShaderi(fshader, gl.COMPILE_STATUS) == gl.FALSE {
		log.Printf("glprog: FS compilation failed: %v", gl.GetShaderInfoLog(fshader))
	}
	//link program
	gl.AttachShader(program, vshader)
	gl.AttachShader(program, fshader)
	gl.LinkProgram(program)
	if gl.GetProgrami(program, gl.LINK_STATUS) == gl.FALSE {
		log.Printf("glprog: LinkProgram failed: %v", gl.GetProgramInfoLog(program))
		gl.DeleteProgram(program)
	}
	//mark shaders for deletion when program is unlinked
	gl.DeleteShader(vshader)
	gl.DeleteShader(fshader)

	p.P = program
	for i := range p.Uniforms {
		p.Uniforms[i].Location = gl.GetUniformLocation(p.P, p.Uniforms[i].Name)
	}
}

func DrawObject(prog *Prog, offset int, count int) {
	if prog.Length <= 0 {
		return
	}
	gl.UseProgram(prog.P)
	t := prog.Buffs["allFloats"]
	gl.BindBuffer(t.Target, t.Buffer)
	for attr := range t.Attribs {
		l := gl.GetAttribLocation(prog.P, t.Attribs[attr].Name)
		if t.Attribs[attr].MatSize > 1 {
			for j := 0; j < t.Attribs[attr].MatSize; j += 1 {
				loc := gl.Attrib{l.Value + uint(j)}
				gl.VertexAttribPointer(
					loc,
					t.Attribs[attr].Size / t.Attribs[attr].MatSize,
					t.TypeGl,
					t.Attribs[attr].Normalized,
					t.Attribs[attr].Stride,
					t.Attribs[attr].Offset + (j * 4 * t.Attribs[attr].MatSize),
				)
				gl.EnableVertexAttribArray(loc)
				defer gl.DisableVertexAttribArray(loc) 
			}
		} else {
			gl.VertexAttribPointer(
				l,
				t.Attribs[attr].Size,
				t.TypeGl,
				t.Attribs[attr].Normalized,
				t.Attribs[attr].Stride,
				t.Attribs[attr].Offset,
			)
			gl.EnableVertexAttribArray(l)
			defer gl.DisableVertexAttribArray(l) //it is ok to call deffer in loop
		}
	}

	for key := range prog.Uniforms {
		s := prog.Uniforms[key]
		if s.Fn == "uniform1i" {
			gl.Uniform1i(s.Location, s.V0.(int))
		} else if s.Fn == "uniform1iv" {
			gl.Uniform1iv(s.Location, s.V0.([]int32))
		} else if s.Fn == "uniform3f" {
			gl.Uniform3f(s.Location, s.V0.(float32), s.V1.(float32), s.V2.(float32))
		} else if s.Fn == "uniform3fv" {
			gl.Uniform3fv(s.Location, s.V0.([]float32))
		} else if s.Fn == "uniformMatrix4fv" {
			gl.UniformMatrix4fv(s.Location, s.V0.([]float32))
		}
	}
	if prog.Mode == gl.TRIANGLES {
		gl.BindBuffer(prog.Buffs["elements"].Target, prog.Buffs["elements"].Buffer)
		gl.DrawElements(prog.Mode, count, prog.Buffs["elements"].TypeGl, offset)
	}
}

//not exactly related with glprog package
func PrintErrors() {
	var msg string
	for {
		error := gl.GetError()
		if error == gl.NO_ERROR {
			log.Print("PrintErrors: NO_ERROR: All error flags are reset")
			return
		}
		if error == gl.INVALID_ENUM {
			msg = "INVALID_ENUM: Enum argument out of range."
		} else if error == gl.INVALID_VALUE {
			msg = "INVALID_VALUE: Numeric argument out of range."
		} else if error == gl.INVALID_OPERATION {
			msg = "INVALID_OPERATION: Operation illegal in current state."
		} else if error == gl.INVALID_FRAMEBUFFER_OPERATION {
			msg = "INVALID_FRAMEBUFFER_OPERATION: Framebuffer is incomplete."
		} else if error == gl.OUT_OF_MEMORY {
			msg = "OUT_OF_MEMORY: Not enough memory left to execute command."
		}
		log.Print("Error: ", msg)
	}
}
