package shader_test

import (
	"os"
	"testing"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yossideutsch/gogl/pkg/shader"
)

var testWindow *glfw.Window

// TestMain sets up OpenGL context once for all tests
func TestMain(m *testing.M) {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		panic("Failed to initialize GLFW: " + err.Error())
	}
	defer glfw.Terminate()

	// Configure OpenGL context
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False) // Keep window hidden

	// Create window
	var err error
	testWindow, err = glfw.CreateWindow(100, 100, "Test", nil, nil)
	if err != nil {
		panic("Failed to create test window: " + err.Error())
	}
	defer testWindow.Destroy()

	// Make context current
	testWindow.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		panic("Failed to initialize OpenGL: " + err.Error())
	}

	// Run tests
	os.Exit(m.Run())
}

func TestCompileVertexShader(t *testing.T) {
	source := `#version 410 core
layout(location = 0) in vec3 aPosition;
void main() {
    gl_Position = vec4(aPosition, 1.0);
}`

	compiledShader, err := shader.CompileShader(source, shader.VertexShader)
	if err != nil {
		t.Fatal("Failed to compile vertex shader:", err)
	}
	defer compiledShader.Delete()

	if compiledShader.ID == 0 {
		t.Error("Shader ID should not be 0")
	}

	if compiledShader.Type != shader.VertexShader {
		t.Error("Shader type should be VertexShader")
	}
}

func TestCompileFragmentShader(t *testing.T) {
	source := `#version 410 core
out vec4 fragColor;
void main() {
    fragColor = vec4(1.0, 0.0, 0.0, 1.0);
}`

	compiledShader, err := shader.CompileShader(source, shader.FragmentShader)
	if err != nil {
		t.Fatal("Failed to compile fragment shader:", err)
	}
	defer compiledShader.Delete()

	if compiledShader.ID == 0 {
		t.Error("Shader ID should not be 0")
	}

	if compiledShader.Type != shader.FragmentShader {
		t.Error("Shader type should be FragmentShader")
	}
}

func TestCompileInvalidShader(t *testing.T) {
	source := `#version 410 core
invalid syntax here
`

	_, err := shader.CompileShader(source, shader.VertexShader)
	if err == nil {
		t.Error("Expected compilation error for invalid shader")
	}
}

func TestCreateProgram(t *testing.T) {
	vertexSource := `#version 410 core
layout(location = 0) in vec3 aPosition;
void main() {
    gl_Position = vec4(aPosition, 1.0);
}`

	fragmentSource := `#version 410 core
out vec4 fragColor;
void main() {
    fragColor = vec4(1.0, 0.0, 0.0, 1.0);
}`

	vertexShader, err := shader.CompileShader(vertexSource, shader.VertexShader)
	if err != nil {
		t.Fatal("Failed to compile vertex shader:", err)
	}
	defer vertexShader.Delete()

	fragmentShader, err := shader.CompileShader(fragmentSource, shader.FragmentShader)
	if err != nil {
		t.Fatal("Failed to compile fragment shader:", err)
	}
	defer fragmentShader.Delete()

	program, err := shader.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		t.Fatal("Failed to create program:", err)
	}
	defer program.Delete()

	if program.ID == 0 {
		t.Error("Program ID should not be 0")
	}

	// Test program validation
	if err := program.Validate(); err != nil {
		// Note: Program validation may fail without a VAO bound, which is expected
		t.Log("Program validation warning:", err)
	}
}

func TestGetUniformLocation(t *testing.T) {
	vertexSource := `#version 410 core
layout(location = 0) in vec3 aPosition;
uniform mat4 uModelViewProjection;
void main() {
    gl_Position = uModelViewProjection * vec4(aPosition, 1.0);
}`

	fragmentSource := `#version 410 core
uniform float uTime;
out vec4 fragColor;
void main() {
    fragColor = vec4(sin(uTime), 0.0, 0.0, 1.0);
}`

	vertexShader, err := shader.CompileShader(vertexSource, shader.VertexShader)
	if err != nil {
		t.Fatal("Failed to compile vertex shader:", err)
	}
	defer vertexShader.Delete()

	fragmentShader, err := shader.CompileShader(fragmentSource, shader.FragmentShader)
	if err != nil {
		t.Fatal("Failed to compile fragment shader:", err)
	}
	defer fragmentShader.Delete()

	program, err := shader.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		t.Fatal("Failed to create program:", err)
	}
	defer program.Delete()

	mvpLocation := program.GetUniformLocation("uModelViewProjection")
	if mvpLocation == -1 {
		t.Error("Failed to get uniform location for uModelViewProjection")
	}

	timeLocation := program.GetUniformLocation("uTime")
	if timeLocation == -1 {
		t.Error("Failed to get uniform location for uTime")
	}

	invalidLocation := program.GetUniformLocation("uNonExistent")
	if invalidLocation != -1 {
		t.Error("Should return -1 for non-existent uniform")
	}
}