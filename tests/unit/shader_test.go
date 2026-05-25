package shader_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yossideutsch1973/gogl/pkg/shader"
)

var testWindow *glfw.Window

const (
	testVertexShaderSource = `#version 410 core
layout(location = 0) in vec3 aPosition;
void main() {
    gl_Position = vec4(aPosition, 1.0);
}`
)

func TestMain(m *testing.M) {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		panic("Failed to initialize GLFW: " + err.Error())
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False)

	var err error
	testWindow, err = glfw.CreateWindow(100, 100, "Test", nil, nil)
	if err != nil {
		panic("Failed to create test window: " + err.Error())
	}
	defer testWindow.Destroy()

	testWindow.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic("Failed to initialize OpenGL: " + err.Error())
	}

	glfw.DetachCurrentContext()
	runtime.UnlockOSThread()

	os.Exit(m.Run())
}

func glSetup(t *testing.T) {
	t.Helper()
	runtime.LockOSThread()
	testWindow.MakeContextCurrent()
	t.Cleanup(func() {
		glfw.DetachCurrentContext()
		runtime.UnlockOSThread()
	})
}

func TestCompileVertexShader(t *testing.T) {
	glSetup(t)
	compiledShader, err := shader.CompileShader(testVertexShaderSource, shader.VertexShader)
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
	glSetup(t)
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
	glSetup(t)
	source := `#version 410 core
invalid syntax here
`

	_, err := shader.CompileShader(source, shader.VertexShader)
	if err == nil {
		t.Error("Expected compilation error for invalid shader")
	}
}

func TestCreateProgram(t *testing.T) {
	glSetup(t)
	fragmentSource := `#version 410 core
out vec4 fragColor;
void main() {
    fragColor = vec4(1.0, 0.0, 0.0, 1.0);
}`

	vertexShader, err := shader.CompileShader(testVertexShaderSource, shader.VertexShader)
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

	if err := program.Validate(); err != nil {
		t.Log("Program validation warning:", err)
	}
}

func TestGetUniformLocation(t *testing.T) {
	glSetup(t)
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

func TestInputValidation(t *testing.T) {
	glSetup(t)
	_, err := shader.CompileShader("", shader.VertexShader)
	if err == nil {
		t.Error("Expected error for empty shader source")
	}

	largeSource := strings.Repeat("//comment\n", 100000)
	_, err = shader.CompileShader(largeSource, shader.VertexShader)
	if err == nil {
		t.Log("Large shader compilation attempted (may fail due to syntax)")
	}
}

func TestProgramValidation(t *testing.T) {
	glSetup(t)
	_, err := shader.CreateProgram()
	if err == nil {
		t.Error("Expected error when creating program with no shaders")
	}

	_, err = shader.CreateProgram(nil)
	if err == nil {
		t.Error("Expected error when creating program with nil shader")
	}

	vertexShader, err := shader.CompileShader(testVertexShaderSource, shader.VertexShader)
	if err != nil {
		t.Fatal("Failed to compile vertex shader:", err)
	}
	defer vertexShader.Delete()

	_, err = shader.CreateProgram(vertexShader)
	if err == nil {
		t.Error("Expected error when creating program with only vertex shader")
	}
}

func TestUniformValidation(t *testing.T) {
	glSetup(t)
	vertexSource := `#version 410 core
layout(location = 0) in vec3 aPosition;
uniform mat4 uMatrix;
void main() {
    gl_Position = uMatrix * vec4(aPosition, 1.0);
}`

	fragmentSource := `#version 410 core
uniform float uFloat;
uniform vec3 uVec3;
out vec4 fragColor;
void main() {
    fragColor = vec4(uVec3 * uFloat, 1.0);
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

	err = program.SetUniform1f(-1, 1.0)
	if err == nil {
		t.Error("Expected error for invalid uniform location in SetUniform1f")
	}

	err = program.SetUniform3f(-1, 1.0, 2.0, 3.0)
	if err == nil {
		t.Error("Expected error for invalid uniform location in SetUniform3f")
	}

	err = program.SetUniformMatrix4fv(0, nil)
	if err == nil {
		t.Error("Expected error for nil matrix in SetUniformMatrix4fv")
	}
}
