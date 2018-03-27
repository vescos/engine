package glprog

import (
	"log"

	"graphs/engine/gles2"
)

//DONE: optimization: USE Interleaved Vertex Data
//https://developer.apple.com/library/ios/documentation/3DDrawing/Conceptual/OpenGLES_ProgrammingGuide/TechniquesforWorkingwithVertexData/TechniquesforWorkingwithVertexData.html
//https://en.wikibooks.org/wiki/OpenGL_Programming/Modern_OpenGL_Tutorial_03

type Uniform struct {
	Name     string
	Location gles2.Uniform
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
}

type Buff struct {
	Buffer  gles2.Buffer
	Target  gles2.Enum
	TypeGl  gles2.Enum
	Usage   gles2.Enum
	Attribs map[string]*Attrib
}

type Prog struct {
	P           gles2.Program
	Vs          string
	Fs          string
	Mode        gles2.Enum
	Length      int
	Buffs       map[string]*Buff
	Uniforms    map[string]*Uniform
	MstUnits    [32]int32
	UseMstUnits bool
}

func (p *Prog) SetBuffer(key string, buffer []byte) {
	if p.Buffs[key].Buffer.Value == 0 {
		p.Buffs[key].Buffer = gles2.CreateBuffer()
	}
	gles2.BindBuffer(p.Buffs[key].Target, p.Buffs[key].Buffer)
	gles2.BufferData(p.Buffs[key].Target, buffer, p.Buffs[key].Usage)
}

//create and link shader program
func (p *Prog) BuildProgram() {
	var (
		program gles2.Program
		vshader gles2.Shader
		fshader gles2.Shader
	)
	program = gles2.CreateProgram()
	//vertex shader
	vshader = gles2.CreateShader(gles2.VERTEX_SHADER)
	gles2.ShaderSource(vshader, p.Vs)
	gles2.CompileShader(vshader)
	if gles2.GetShaderi(vshader, gles2.COMPILE_STATUS) == gles2.FALSE {
		log.Printf("glprog: VS compilation failed: %v", gles2.GetShaderInfoLog(vshader))
	}
	//fragment shader
	fshader = gles2.CreateShader(gles2.FRAGMENT_SHADER)
	gles2.ShaderSource(fshader, p.Fs)
	gles2.CompileShader(fshader)
	if gles2.GetShaderi(fshader, gles2.COMPILE_STATUS) == gles2.FALSE {
		log.Printf("glprog: FS compilation failed: %v", gles2.GetShaderInfoLog(fshader))
	}
	//link program
	gles2.AttachShader(program, vshader)
	gles2.AttachShader(program, fshader)
	gles2.LinkProgram(program)
	if gles2.GetProgrami(program, gles2.LINK_STATUS) == gles2.FALSE {
		log.Printf("glprog: LinkProgram failed: %v", gles2.GetProgramInfoLog(program))
		gles2.DeleteProgram(program)
	}
	//mark shaders for deletion when program is unlinked
	gles2.DeleteShader(vshader)
	gles2.DeleteShader(fshader)

	p.P = program
	for i := range p.Uniforms {
		p.Uniforms[i].Location = gles2.GetUniformLocation(p.P, p.Uniforms[i].Name)
	}
}

func DrawObject(prog *Prog, ctMaxTexureU int32) {
	if prog.Length <= 0 {
		return
	}
	gles2.UseProgram(prog.P)
	t := prog.Buffs["allFloats"]
	gles2.BindBuffer(t.Target, t.Buffer)
	for attr := range t.Attribs {
		l := gles2.GetAttribLocation(prog.P, t.Attribs[attr].Name)
		gles2.VertexAttribPointer(
			l,
			t.Attribs[attr].Size,
			t.TypeGl,
			t.Attribs[attr].Normalized,
			t.Attribs[attr].Stride,
			t.Attribs[attr].Offset)
		gles2.EnableVertexAttribArray(l)
		defer gles2.DisableVertexAttribArray(l) //it is ok to call deffer in loop
	}

	for key := range prog.Uniforms {
		s := prog.Uniforms[key]
		if s.Fn == "uniform1i" {
			gles2.Uniform1i(s.Location, s.V0.(int))
		} else if s.Fn == "uniform1iv" {
			gles2.Uniform1iv(s.Location, s.V0.([]int32))
		} else if s.Fn == "uniform3f" {
			gles2.Uniform3f(s.Location, s.V0.(float32), s.V1.(float32), s.V2.(float32))
		} else if s.Fn == "uniform3fv" {
			gles2.Uniform3fv(s.Location, s.V0.([]float32))
		} else if s.Fn == "uniformMatrix4fv" {
			gles2.UniformMatrix4fv(s.Location, s.V0.([]float32))
		}
	}
	if prog.Mode == gles2.TRIANGLES {
		if prog.UseMstUnits {
			offset := 0
			count := 0
			for k, v := range prog.MstUnits {
				if v == 0 {
					continue
				}
				count = int(v)
				//FIXME: hardcoded to use unit 6 and 7 if limit is set to max 8
				if ctMaxTexureU-1 < int32(k) {
					gles2.Uniform1i(prog.Uniforms["texture"].Location, 6)
					gles2.Uniform1i(prog.Uniforms["texture_normal"].Location, 7)
				} else {
					gles2.Uniform1i(prog.Uniforms["texture"].Location, int(k))
					gles2.Uniform1i(prog.Uniforms["texture_normal"].Location, int(k+1))
				}
				gles2.BindBuffer(prog.Buffs["elements"].Target, prog.Buffs["elements"].Buffer)
				gles2.DrawElements(prog.Mode, count, prog.Buffs["elements"].TypeGl, offset*2)
				offset += count
			}
		} else {
			gles2.BindBuffer(prog.Buffs["elements"].Target, prog.Buffs["elements"].Buffer)
			gles2.DrawElements(prog.Mode, prog.Length, prog.Buffs["elements"].TypeGl, 0)
		}
	}
}

//not exactly related with glprog package
func PrintErrors() {
	var msg string
	for {
		error := gles2.GetError()
		if error == gles2.NO_ERROR {
			log.Print("PrintErrors: NO_ERROR: All error flags are reset")
			return
		}
		if error == gles2.INVALID_ENUM {
			msg = "INVALID_ENUM: Enum argument out of range."
		} else if error == gles2.INVALID_VALUE {
			msg = "INVALID_VALUE: Numeric argument out of range."
		} else if error == gles2.INVALID_OPERATION {
			msg = "INVALID_OPERATION: Operation illegal in current state."
		} else if error == gles2.INVALID_FRAMEBUFFER_OPERATION {
			msg = "INVALID_FRAMEBUFFER_OPERATION: Framebuffer is incomplete."
		} else if error == gles2.OUT_OF_MEMORY {
			msg = "OUT_OF_MEMORY: Not enough memory left to execute command."
		}
		log.Print("Error: ", msg)
	}
}
