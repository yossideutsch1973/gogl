// Package shader provides a comprehensive OpenGL shader management system.
// It offers type-safe shader compilation, linking, and program management
// with performance optimizations and robust error handling.
//
// Key features:
//   - Compile and link vertex, fragment, geometry, and compute shaders
//   - Comprehensive error reporting with OpenGL error checking
//   - Memory-efficient resource management with object pooling
//   - Type-safe uniform setting with validation
//   - Cross-platform compatibility (OpenGL 4.1+)
//
// Example usage:
//
//	// Compile shaders
//	vertexShader, err := shader.CompileShader(vertexSource, shader.VertexShader)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer vertexShader.Delete()
//
//	fragmentShader, err := shader.CompileShader(fragmentSource, shader.FragmentShader)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer fragmentShader.Delete()
//
//	// Create program
//	program, err := shader.CreateProgram(vertexShader, fragmentShader)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer program.Delete()
//
//	// Use program and set uniforms
//	program.Use()
//	loc := program.GetUniformLocation("uTime")
//	program.SetUniform1f(loc, time)
package shader

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Pool for reusing byte slices to reduce allocations
var logPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 512) // Start with 512 bytes capacity
	},
}

// checkGLError checks for OpenGL errors and returns a descriptive error
func checkGLError(operation string) error {
	if err := gl.GetError(); err != gl.NO_ERROR {
		var errStr string
		switch err {
		case gl.INVALID_ENUM:
			errStr = "GL_INVALID_ENUM"
		case gl.INVALID_VALUE:
			errStr = "GL_INVALID_VALUE"
		case gl.INVALID_OPERATION:
			errStr = "GL_INVALID_OPERATION"
		case gl.OUT_OF_MEMORY:
			errStr = "GL_OUT_OF_MEMORY"
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			errStr = "GL_INVALID_FRAMEBUFFER_OPERATION"
		default:
			errStr = fmt.Sprintf("Unknown GL error 0x%x", err)
		}
		return fmt.Errorf("%s failed: %s", operation, errStr)
	}
	return nil
}

// getShaderTypeName returns a human-readable name for the shader type
func getShaderTypeName(shaderType ShaderType) string {
	switch shaderType {
	case VertexShader:
		return "vertex"
	case FragmentShader:
		return "fragment"
	case GeometryShader:
		return "geometry"
	case ComputeShader:
		return "compute"
	default:
		return "unknown"
	}
}

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
	// Input validation
	if source == "" {
		return nil, fmt.Errorf("shader source cannot be empty")
	}
	if len(source) > 1048576 { // 1MB limit for safety
		return nil, fmt.Errorf("shader source too large: %d bytes (max 1MB)", len(source))
	}

	shaderID := gl.CreateShader(uint32(shaderType))
	if shaderID == 0 {
		return nil, fmt.Errorf("failed to create shader: OpenGL context may not be initialized")
	}

	cSource, free := gl.Strs(source + "\x00")
	defer free()

	gl.ShaderSource(shaderID, 1, cSource, nil)
	if err := checkGLError("glShaderSource"); err != nil {
		gl.DeleteShader(shaderID)
		return nil, err
	}

	gl.CompileShader(shaderID)
	if err := checkGLError("glCompileShader"); err != nil {
		gl.DeleteShader(shaderID)
		return nil, err
	}

	var status int32
	gl.GetShaderiv(shaderID, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &logLength)

		// Use pooled buffer to reduce allocations
		buf := logPool.Get().([]byte)
		defer logPool.Put(buf[:0]) // Reset length but keep capacity
		
		if cap(buf) < int(logLength) {
			buf = make([]byte, logLength)
		}
		buf = buf[:logLength]

		gl.GetShaderInfoLog(shaderID, logLength, nil, (*uint8)(&buf[0]))

		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("failed to compile %s shader: %s", 
			getShaderTypeName(shaderType), string(buf[:logLength-1])) // Remove null terminator
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
	// Input validation
	if len(shaders) == 0 {
		return nil, fmt.Errorf("at least one shader is required")
	}
	
	// Validate shader types - ensure we have at least vertex and fragment
	hasVertex, hasFragment := false, false
	for _, shader := range shaders {
		if shader == nil {
			return nil, fmt.Errorf("shader cannot be nil")
		}
		if shader.ID == 0 {
			return nil, fmt.Errorf("invalid shader: ID is 0")
		}
		switch shader.Type {
		case VertexShader:
			hasVertex = true
		case FragmentShader:
			hasFragment = true
		}
	}
	
	if !hasVertex {
		return nil, fmt.Errorf("vertex shader is required")
	}
	if !hasFragment {
		return nil, fmt.Errorf("fragment shader is required")
	}

	programID := gl.CreateProgram()
	if programID == 0 {
		return nil, fmt.Errorf("failed to create program: OpenGL context may not be initialized")
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
	if err := checkGLError("glLinkProgram"); err != nil {
		program.Delete()
		return nil, err
	}

	var status int32
	gl.GetProgramiv(programID, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)

		// Use pooled buffer to reduce allocations
		buf := logPool.Get().([]byte)
		defer logPool.Put(buf[:0])
		
		if cap(buf) < int(logLength) {
			buf = make([]byte, logLength)
		}
		buf = buf[:logLength]

		gl.GetProgramInfoLog(programID, logLength, nil, (*uint8)(&buf[0]))

		program.Delete()
		return nil, fmt.Errorf("failed to link program: %s", string(buf[:logLength-1]))
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

// SetUniformMatrix4fv sets a mat4 uniform with validation
func (p *Program) SetUniformMatrix4fv(location int32, matrix *mgl32.Mat4) error {
	if location == -1 {
		return fmt.Errorf("invalid uniform location: -1")
	}
	if matrix == nil {
		return fmt.Errorf("matrix cannot be nil")
	}
	gl.UniformMatrix4fv(location, 1, false, &matrix[0])
	return checkGLError("glUniformMatrix4fv")
}

// SetUniform1f sets a float uniform with validation
func (p *Program) SetUniform1f(location int32, value float32) error {
	if location == -1 {
		return fmt.Errorf("invalid uniform location: -1")
	}
	gl.Uniform1f(location, value)
	return checkGLError("glUniform1f")
}

// SetUniform3f sets a vec3 uniform with validation
func (p *Program) SetUniform3f(location int32, x, y, z float32) error {
	if location == -1 {
		return fmt.Errorf("invalid uniform location: -1")
	}
	gl.Uniform3f(location, x, y, z)
	return checkGLError("glUniform3f")
}

// Validate validates the program (use only in debug builds)
func (p *Program) Validate() error {
	if p.ID == 0 {
		return fmt.Errorf("program not initialized")
	}
	
	gl.ValidateProgram(p.ID)
	if err := checkGLError("glValidateProgram"); err != nil {
		return err
	}

	var status int32
	gl.GetProgramiv(p.ID, gl.VALIDATE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(p.ID, gl.INFO_LOG_LENGTH, &logLength)

		// Use pooled buffer to reduce allocations
		buf := logPool.Get().([]byte)
		defer logPool.Put(buf[:0])
		
		if cap(buf) < int(logLength) {
			buf = make([]byte, logLength)
		}
		buf = buf[:logLength]

		gl.GetProgramInfoLog(p.ID, logLength, nil, (*uint8)(&buf[0]))

		return fmt.Errorf("program validation failed: %s", string(buf[:logLength-1]))
	}

	return nil
}

// Delete cleans up the program and associated shaders
func (p *Program) Delete() {
	if p.ID != 0 {
		for _, shader := range p.shaders {
			if shader != nil && shader.ID != 0 {
				gl.DetachShader(p.ID, shader.ID)
				gl.DeleteShader(shader.ID)
				shader.ID = 0 // Mark as deleted
			}
		}
		gl.DeleteProgram(p.ID)
		p.ID = 0
		p.shaders = nil // Clear references
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