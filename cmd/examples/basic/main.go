package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/yossideutsch/gogl/pkg/shader"
)

const (
	windowWidth  = 800
	windowHeight = 600
	windowTitle  = "GoGL Basic Shader Example"
)

var (
	// Triangle vertices: position (x,y,z) + color (r,g,b)
	vertices = []float32{
		// Top vertex (red)
		0.0, 0.5, 0.0, 1.0, 0.0, 0.0,
		// Bottom left (green)
		-0.5, -0.5, 0.0, 0.0, 1.0, 0.0,
		// Bottom right (blue)
		0.5, -0.5, 0.0, 0.0, 0.0, 1.0,
	}
)

func main() {
	// Lock the main thread for OpenGL calls
	runtime.LockOSThread()

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatal("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Configure OpenGL version and profile
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True) // Required on macOS

	// Create window
	window, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
	if err != nil {
		log.Fatal("Failed to create window:", err)
	}
	defer window.Destroy()

	// Make context current
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	// Print OpenGL version info
	fmt.Printf("OpenGL version: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))
	fmt.Printf("GLSL version: %s\n", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	fmt.Printf("Vendor: %s\n", gl.GoStr(gl.GetString(gl.VENDOR)))
	fmt.Printf("Renderer: %s\n", gl.GoStr(gl.GetString(gl.RENDERER)))

	// Create and compile shaders
	vertexShader, err := shader.CompileShaderFromFile("shaders/vertex/basic.vert", shader.VertexShader)
	if err != nil {
		log.Fatal("Failed to compile vertex shader:", err)
	}
	defer vertexShader.Delete()

	fragmentShader, err := shader.CompileShaderFromFile("shaders/fragment/basic.frag", shader.FragmentShader)
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

	// Validate program (debug only)
	if err := program.Validate(); err != nil {
		log.Println("Program validation warning:", err)
	}

	// Get uniform locations
	mvpLocation := program.GetUniformLocation("uModelViewProjection")
	timeLocation := program.GetUniformLocation("uTime")

	// Create VAO and VBO
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	defer gl.DeleteVertexArrays(1, &vao)
	defer gl.DeleteBuffers(1, &vbo)

	// Bind and configure VAO
	gl.BindVertexArray(vao)

	// Upload vertex data
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Configure vertex attributes
	// Position attribute (location 0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, nil)
	gl.EnableVertexAttribArray(0)

	// Color attribute (location 1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Unbind VAO
	gl.BindVertexArray(0)

	// Configure OpenGL state
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Enable(gl.DEPTH_TEST)

	// Setup projection matrix
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/float32(windowHeight), 0.1, 100.0)
	view := mgl32.LookAtV(mgl32.Vec3{0, 0, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	model := mgl32.Ident4()

	// Enable V-Sync
	glfw.SwapInterval(1)

	// Main render loop
	startTime := glfw.GetTime()
	for !window.ShouldClose() {
		// Poll events
		glfw.PollEvents()

		// Check for escape key
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
		}

		// Clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Use shader program
		program.Use()

		// Update time uniform
		currentTime := float32(glfw.GetTime() - startTime)
		if err := program.SetUniform1f(timeLocation, currentTime); err != nil {
			log.Printf("Warning: Failed to set time uniform: %v", err)
		}

		// Create rotating model matrix
		rotationAngle := currentTime * 0.5
		model = mgl32.HomogRotate3DY(rotationAngle)

		// Calculate MVP matrix
		mvp := projection.Mul4(view).Mul4(model)
		if err := program.SetUniformMatrix4fv(mvpLocation, &mvp); err != nil {
			log.Printf("Warning: Failed to set MVP uniform: %v", err)
		}

		// Draw triangle
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		// Swap buffers
		window.SwapBuffers()
	}

	fmt.Println("Shader evaluation completed successfully!")
}