package shader

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// ShaderType represents the type of shader
type ShaderType uint32

const (
	VertexShader   ShaderType = gl.VERTEX_SHADER
	FragmentShader ShaderType = gl.FRAGMENT_SHADER
	GeometryShader ShaderType = gl.GEOMETRY_SHADER
	ComputeShader  ShaderType = gl.COMPUTE_SHADER
)

// Shader represents a compiled OpenGL shader
type Shader struct {
	ID   uint32
	Type ShaderType
}

// Program represents a linked shader program
type Program struct {
	ID      uint32
	shaders []*Shader
}

// CompileShader compiles a shader from source code
func CompileShader(source string, shaderType ShaderType) (*Shader, error) {
	shaderID := gl.CreateShader(uint32(shaderType))
	if shaderID == 0 {
		return nil, fmt.Errorf("failed to create shader")
	}

	cSource, free := gl.Strs(source + "\x00")
	defer free()

	gl.ShaderSource(shaderID, 1, cSource, nil)
	gl.CompileShader(shaderID)

	var status int32
	gl.GetShaderiv(shaderID, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shaderID, logLength, nil, gl.Str(log))

		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("failed to compile shader: %v", log)
	}

	return &Shader{
		ID:   shaderID,
		Type: shaderType,
	}, nil
}

// CompileShaderFromFile compiles a shader from a file
func CompileShaderFromFile(filepath string, shaderType ShaderType) (*Shader, error) {
	source, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read shader file %s: %w", filepath, err)
	}

	return CompileShader(string(source), shaderType)
}

// CreateProgram creates a new shader program
func CreateProgram(shaders ...*Shader) (*Program, error) {
	programID := gl.CreateProgram()
	if programID == 0 {
		return nil, fmt.Errorf("failed to create program")
	}

	program := &Program{
		ID:      programID,
		shaders: make([]*Shader, len(shaders)),
	}

	// Attach all shaders
	for i, shader := range shaders {
		gl.AttachShader(programID, shader.ID)
		program.shaders[i] = shader
	}

	// Link the program
	gl.LinkProgram(programID)

	var status int32
	gl.GetProgramiv(programID, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(programID, logLength, nil, gl.Str(log))

		program.Delete()
		return nil, fmt.Errorf("failed to link program: %v", log)
	}

	return program, nil
}

// Use activates the shader program
func (p *Program) Use() {
	gl.UseProgram(p.ID)
}

// GetUniformLocation returns the location of a uniform variable
func (p *Program) GetUniformLocation(name string) int32 {
	return gl.GetUniformLocation(p.ID, gl.Str(name+"\x00"))
}

// SetUniformMatrix4fv sets a mat4 uniform
func (p *Program) SetUniformMatrix4fv(location int32, matrix *mgl32.Mat4) {
	gl.UniformMatrix4fv(location, 1, false, &matrix[0])
}

// SetUniform1f sets a float uniform
func (p *Program) SetUniform1f(location int32, value float32) {
	gl.Uniform1f(location, value)
}

// SetUniform3f sets a vec3 uniform
func (p *Program) SetUniform3f(location int32, x, y, z float32) {
	gl.Uniform3f(location, x, y, z)
}

// Validate validates the program (use only in debug builds)
func (p *Program) Validate() error {
	gl.ValidateProgram(p.ID)

	var status int32
	gl.GetProgramiv(p.ID, gl.VALIDATE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(p.ID, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(p.ID, logLength, nil, gl.Str(log))

		return fmt.Errorf("program validation failed: %v", log)
	}

	return nil
}

// Delete cleans up the program and associated shaders
func (p *Program) Delete() {
	if p.ID != 0 {
		for _, shader := range p.shaders {
			if shader.ID != 0 {
				gl.DetachShader(p.ID, shader.ID)
				gl.DeleteShader(shader.ID)
			}
		}
		gl.DeleteProgram(p.ID)
		p.ID = 0
	}
}

// Delete cleans up the shader
func (s *Shader) Delete() {
	if s.ID != 0 {
		gl.DeleteShader(s.ID)
		s.ID = 0
	}
}

// DispatchCompute dispatches compute work groups (only valid for compute shaders)
func (p *Program) DispatchCompute(numGroupsX, numGroupsY, numGroupsZ uint32) {
	p.Use()
	gl.DispatchCompute(numGroupsX, numGroupsY, numGroupsZ)
}

// MemoryBarrier ensures memory writes are visible to subsequent operations
func (p *Program) MemoryBarrier(barriers uint32) {
	gl.MemoryBarrier(barriers)
}

// GetWorkGroupSize returns the local work group size for compute shaders
func (p *Program) GetWorkGroupSize() (x, y, z int32) {
	var size [3]int32
	gl.GetProgramiv(p.ID, gl.COMPUTE_WORK_GROUP_SIZE, &size[0])
	return size[0], size[1], size[2]
}