package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/yossideutsch/gogl/pkg/pipeline"
	"github.com/yossideutsch/gogl/pkg/resource"
	"github.com/yossideutsch/gogl/pkg/shader"
)

const (
	windowWidth  = 800
	windowHeight = 600
)

var (
	// Cube vertices with positions and colors
	cubeVertices = []float32{
		// Front face (red)
		-0.5, -0.5,  0.5,  1.0, 0.0, 0.0,
		 0.5, -0.5,  0.5,  1.0, 0.0, 0.0,
		 0.5,  0.5,  0.5,  1.0, 0.0, 0.0,
		-0.5,  0.5,  0.5,  1.0, 0.0, 0.0,

		// Back face (green)
		-0.5, -0.5, -0.5,  0.0, 1.0, 0.0,
		 0.5, -0.5, -0.5,  0.0, 1.0, 0.0,
		 0.5,  0.5, -0.5,  0.0, 1.0, 0.0,
		-0.5,  0.5, -0.5,  0.0, 1.0, 0.0,

		// Top face (blue)
		-0.5,  0.5,  0.5,  0.0, 0.0, 1.0,
		 0.5,  0.5,  0.5,  0.0, 0.0, 1.0,
		 0.5,  0.5, -0.5,  0.0, 0.0, 1.0,
		-0.5,  0.5, -0.5,  0.0, 0.0, 1.0,

		// Bottom face (yellow)
		-0.5, -0.5,  0.5,  1.0, 1.0, 0.0,
		 0.5, -0.5,  0.5,  1.0, 1.0, 0.0,
		 0.5, -0.5, -0.5,  1.0, 1.0, 0.0,
		-0.5, -0.5, -0.5,  1.0, 1.0, 0.0,

		// Right face (magenta)
		 0.5, -0.5,  0.5,  1.0, 0.0, 1.0,
		 0.5, -0.5, -0.5,  1.0, 0.0, 1.0,
		 0.5,  0.5, -0.5,  1.0, 0.0, 1.0,
		 0.5,  0.5,  0.5,  1.0, 0.0, 1.0,

		// Left face (cyan)
		-0.5, -0.5,  0.5,  0.0, 1.0, 1.0,
		-0.5, -0.5, -0.5,  0.0, 1.0, 1.0,
		-0.5,  0.5, -0.5,  0.0, 1.0, 1.0,
		-0.5,  0.5,  0.5,  0.0, 1.0, 1.0,
	}

	// Cube indices
	cubeIndices = []uint32{
		// Front face
		0, 1, 2, 2, 3, 0,
		// Back face
		4, 5, 6, 6, 7, 4,
		// Top face
		8, 9, 10, 10, 11, 8,
		// Bottom face
		12, 13, 14, 14, 15, 12,
		// Right face
		16, 17, 18, 18, 19, 16,
		// Left face
		20, 21, 22, 22, 23, 20,
	}

	vertexShaderSource = `#version 410 core
layout(location = 0) in vec3 aPosition;
layout(location = 1) in vec3 aColor;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProjection;

out vec3 vColor;

void main() {
    vColor = aColor;
    gl_Position = uProjection * uView * uModel * vec4(aPosition, 1.0);
}`

	fragmentShaderSource = `#version 410 core
in vec3 vColor;
out vec4 fragColor;

void main() {
    fragColor = vec4(vColor, 1.0);
}`
)

func init() {
	// This is needed to arrange for main() to run on the main thread
	runtime.LockOSThread()
}

func main() {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatal("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Configure GLFW
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create window
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Pipeline Example - Rotating Cube", nil, nil)
	if err != nil {
		log.Fatal("Failed to create window:", err)
	}

	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	// Print OpenGL info
	fmt.Printf("OpenGL version: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))
	fmt.Printf("GLSL version: %s\n", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	fmt.Printf("Vendor: %s\n", gl.GoStr(gl.GetString(gl.VENDOR)))
	fmt.Printf("Renderer: %s\n", gl.GoStr(gl.GetString(gl.RENDERER)))

	// Compile shaders
	vertexShader, err := shader.CompileShader(vertexShaderSource, shader.VertexShader)
	if err != nil {
		log.Fatal("Failed to compile vertex shader:", err)
	}
	defer vertexShader.Delete()

	fragmentShader, err := shader.CompileShader(fragmentShaderSource, shader.FragmentShader)
	if err != nil {
		log.Fatal("Failed to compile fragment shader:", err)
	}
	defer fragmentShader.Delete()

	// Create shader program
	program, err := shader.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatal("Failed to create shader program:", err)
	}
	defer program.Delete()

	// Create mesh using the resource system
	layout := resource.NewVertexLayout().
		AddFloat(0, 3). // Position
		AddFloat(1, 3)  // Color

	mesh, err := resource.NewMesh(cubeVertices, cubeIndices, layout)
	if err != nil {
		log.Fatal("Failed to create mesh:", err)
	}
	defer mesh.Delete()

	// Create pipeline
	renderPipeline := pipeline.New()

	// Configure pipeline state
	pipelineState := pipeline.NewBuilder().
		WithProgram(program).
		WithDepthTest(true, true, pipeline.DepthLess).
		WithCulling(true, pipeline.CullBack).
		WithViewport(0, 0, windowWidth, windowHeight).
		Build()

	// Get uniform locations
	modelLoc := program.GetUniformLocation("uModel")
	viewLoc := program.GetUniformLocation("uView")
	projLoc := program.GetUniformLocation("uProjection")

	// Create transformation matrices
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/float32(windowHeight), 0.1, 100.0)
	view := mgl32.LookAtV(mgl32.Vec3{2, 2, 2}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	// Enable vsync
	glfw.SwapInterval(1)

	// Wireframe mode toggle
	wireframe := false

	// Set key callback for wireframe toggle
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyW && action == glfw.Press {
			wireframe = !wireframe
			renderPipeline.SetWireframe(wireframe)
			if wireframe {
				fmt.Println("Wireframe mode: ON")
			} else {
				fmt.Println("Wireframe mode: OFF")
			}
		}
		if key == glfw.KeyEscape && action == glfw.Press {
			w.SetShouldClose(true)
		}
	})

	fmt.Println("Press 'W' to toggle wireframe mode")
	fmt.Println("Press 'ESC' to exit")

	// Main loop
	for !window.ShouldClose() {
		// Poll events
		glfw.PollEvents()

		// Clear screen
		renderPipeline.SetClearColor(0.1, 0.1, 0.1, 1.0)
		renderPipeline.Clear(true, true, false)

		// Apply pipeline state
		if err := renderPipeline.SetState(pipelineState); err != nil {
			log.Fatal("Failed to set pipeline state:", err)
		}

		// Calculate rotation
		time := float32(glfw.GetTime())
		model := mgl32.HomogRotate3DY(time).Mul4(mgl32.HomogRotate3DX(time * 0.5))

		// Set uniforms
		program.SetUniformMatrix4fv(modelLoc, &model)
		program.SetUniformMatrix4fv(viewLoc, &view)
		program.SetUniformMatrix4fv(projLoc, &projection)

		// Draw mesh
		mesh.Draw(gl.TRIANGLES)

		// Swap buffers
		window.SwapBuffers()
	}
}