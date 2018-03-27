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
