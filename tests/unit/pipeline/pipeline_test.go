package pipeline_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yossideutsch1973/gogl/pkg/pipeline"
	"github.com/yossideutsch1973/gogl/pkg/shader"
)

var testWindow *glfw.Window

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

func TestPipelineCreation(t *testing.T) {
	p := pipeline.New()
	if p == nil {
		t.Fatal("Failed to create pipeline")
	}

	state := p.GetState()
	if state == nil {
		t.Fatal("Pipeline state should not be nil")
	}

	if !state.DepthEnabled {
		t.Error("Depth test should be enabled by default")
	}

	if !state.CullEnabled {
		t.Error("Face culling should be enabled by default")
	}

	if state.BlendEnabled {
		t.Error("Blending should be disabled by default")
	}
}

func TestStateBuilder(t *testing.T) {
	builder := pipeline.NewBuilder()

	state := builder.
		WithBlending(true, pipeline.BlendSrcAlpha, pipeline.BlendOneMinusSrcAlpha).
		WithDepthTest(true, true, pipeline.DepthLess).
		WithCulling(true, pipeline.CullBack).
		WithViewport(0, 0, 1920, 1080).
		WithWireframe(false).
		WithPrimitive(pipeline.Triangles).
		Build()

	if state == nil {
		t.Fatal("Builder should return non-nil state")
	}

	if !state.BlendEnabled {
		t.Error("Blending should be enabled")
	}

	if state.BlendSrc != pipeline.BlendSrcAlpha {
		t.Error("Incorrect blend source function")
	}

	if state.ViewportWidth != 1920 || state.ViewportHeight != 1080 {
		t.Error("Incorrect viewport dimensions")
	}

	if state.Primitive != pipeline.Triangles {
		t.Error("Incorrect primitive type")
	}
}

func TestStateValidation(t *testing.T) {
	validState := pipeline.DefaultState()
	if err := validState.Validate(); err != nil {
		t.Error("Default state should be valid:", err)
	}

	invalidViewport := pipeline.DefaultState()
	invalidViewport.ViewportWidth = 0
	if err := invalidViewport.Validate(); err == nil {
		t.Error("State with zero viewport width should be invalid")
	}

	invalidBlend := pipeline.DefaultState()
	invalidBlend.BlendEnabled = true
	invalidBlend.BlendSrc = pipeline.BlendZero
	invalidBlend.BlendDst = pipeline.BlendZero
	if err := invalidBlend.Validate(); err == nil {
		t.Error("State with both blend functions as ZERO should be invalid")
	}
}

func TestPipelineStateStack(t *testing.T) {
	glSetup(t)
	p := pipeline.New()

	initialState := p.GetState()
	initialDepth := initialState.DepthEnabled

	p.PushState()

	p.SetDepthTest(false, false, pipeline.DepthAlways)

	if p.GetState().DepthEnabled {
		t.Error("Depth test should be disabled")
	}

	if err := p.PopState(); err != nil {
		t.Fatal("Failed to pop state:", err)
	}

	if p.GetState().DepthEnabled != initialDepth {
		t.Error("State not properly restored after pop")
	}

	if err := p.PopState(); err == nil {
		t.Error("Popping from empty stack should return error")
	}
}

func TestPipelineStateApplication(t *testing.T) {
	glSetup(t)
	p := pipeline.New()

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

	state := pipeline.NewBuilder().
		WithProgram(program).
		WithBlending(true, pipeline.BlendSrcAlpha, pipeline.BlendOneMinusSrcAlpha).
		WithDepthTest(true, true, pipeline.DepthLess).
		Build()

	if err := p.SetState(state); err != nil {
		t.Fatal("Failed to set pipeline state:", err)
	}

	if p.GetState().Program != program {
		t.Error("Program not properly set")
	}
}

func TestPipelineIndividualSetters(t *testing.T) {
	glSetup(t)
	p := pipeline.New()

	p.SetBlending(true, pipeline.BlendOne, pipeline.BlendOne)
	if !p.GetState().BlendEnabled {
		t.Error("Blending should be enabled")
	}
	if p.GetState().BlendSrc != pipeline.BlendOne {
		t.Error("Blend source not set correctly")
	}

	p.SetDepthTest(false, false, pipeline.DepthAlways)
	if p.GetState().DepthEnabled {
		t.Error("Depth test should be disabled")
	}
	if p.GetState().DepthFunc != pipeline.DepthAlways {
		t.Error("Depth function not set correctly")
	}

	p.SetCulling(false, pipeline.CullNone)
	if p.GetState().CullEnabled {
		t.Error("Culling should be disabled")
	}

	p.SetViewport(10, 20, 640, 480)
	if p.GetState().ViewportX != 10 || p.GetState().ViewportY != 20 {
		t.Error("Viewport position not set correctly")
	}
	if p.GetState().ViewportWidth != 640 || p.GetState().ViewportHeight != 480 {
		t.Error("Viewport dimensions not set correctly")
	}

	p.SetWireframe(true)
	if !p.GetState().WireframeMode {
		t.Error("Wireframe mode should be enabled")
	}
}

func TestPipelineClearOperations(t *testing.T) {
	glSetup(t)
	p := pipeline.New()

	p.SetClearColor(0.1, 0.2, 0.3, 1.0)

	p.Clear(true, true, false)
	p.Clear(false, true, true)
	p.Clear(true, false, true)
}
