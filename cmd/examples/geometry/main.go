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
	// Point positions for point-to-quad demo
	points = []float32{
		-0.5, -0.5, 0.0, 1.0, 0.0, 0.0, // Red point
		 0.5, -0.5, 0.0, 0.0, 1.0, 0.0, // Green point
		 0.0,  0.5, 0.0, 0.0, 0.0, 1.0, // Blue point
	}

	vertexShaderSource = `#version 410 core
layout(location = 0) in vec3 aPosition;
layout(location = 1) in vec3 aColor;

uniform mat4 uModel;

out vec3 vColor;

void main() {
    vColor = aColor;
    gl_Position = uModel * vec4(aPosition, 1.0);
    gl_PointSize = 10.0;
}`

	geometryShaderSource = `#version 410 core

layout(points) in;
layout(triangle_strip, max_vertices = 4) out;

uniform mat4 uProjection;
uniform mat4 uView;
uniform float uPointSize;

in vec3 vColor[];
out vec3 fColor;
out vec2 fTexCoord;

void main() {
    vec4 position = gl_in[0].gl_Position;
    fColor = vColor[0];
    
    float size = uPointSize * 0.5;
    
    // Bottom-left vertex
    gl_Position = uProjection * uView * (position + vec4(-size, -size, 0.0, 0.0));
    fTexCoord = vec2(0.0, 0.0);
    EmitVertex();
    
    // Bottom-right vertex
    gl_Position = uProjection * uView * (position + vec4(size, -size, 0.0, 0.0));
    fTexCoord = vec2(1.0, 0.0);
    EmitVertex();
    
    // Top-left vertex
    gl_Position = uProjection * uView * (position + vec4(-size, size, 0.0, 0.0));
    fTexCoord = vec2(0.0, 1.0);
    EmitVertex();
    
    // Top-right vertex
    gl_Position = uProjection * uView * (position + vec4(size, size, 0.0, 0.0));
    fTexCoord = vec2(1.0, 1.0);
    EmitVertex();
    
    EndPrimitive();
}`

	fragmentShaderSource = `#version 410 core
in vec3 fColor;
in vec2 fTexCoord;
out vec4 fragColor;

void main() {
    // Create circular points
    vec2 center = vec2(0.5, 0.5);
    float distance = length(fTexCoord - center);
    
    if (distance > 0.5) {
        discard;
    }
    
    // Add some shading based on distance from center
    float intensity = 1.0 - (distance * 2.0);
    fragColor = vec4(fColor * intensity, 1.0);
}`

	// Wireframe shaders for comparison
	wireframeGeometrySource = `#version 410 core

layout(triangles) in;
layout(line_strip, max_vertices = 4) out;

in vec3 vColor[];
out vec3 fColor;

void main() {
    // Emit the three edges of the triangle
    for(int i = 0; i < 3; i++) {
        gl_Position = gl_in[i].gl_Position;
        fColor = vColor[i];
        EmitVertex();
    }
    
    // Complete the triangle by connecting back to the first vertex
    gl_Position = gl_in[0].gl_Position;
    fColor = vColor[0];
    EmitVertex();
    
    EndPrimitive();
}`

	wireframeFragmentSource = `#version 410 core
in vec3 fColor;
out vec4 fragColor;

void main() {
    fragColor = vec4(fColor, 1.0);
}`
)

func init() {
	runtime.LockOSThread()
}

type Demo struct {
	// Point-to-quad demo
	pointProgram *shader.Program
	pointMesh    *resource.Mesh

	// Wireframe demo
	wireframeProgram *shader.Program
	triangleMesh     *resource.Mesh

	// Pipeline
	renderPipeline *pipeline.Pipeline

	// Matrices
	projection mgl32.Mat4
	view       mgl32.Mat4

	// Demo state
	currentDemo    int
	pointSize      float32
	animationSpeed float32
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
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Geometry Shader Demo", nil, nil)
	if err != nil {
		log.Fatal("Failed to create window:", err)
	}
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	fmt.Printf("OpenGL version: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))

	// Create demo
	demo, err := NewDemo()
	if err != nil {
		log.Fatal("Failed to create demo:", err)
	}
	defer demo.Cleanup()

	// Set up input handling
	window.SetKeyCallback(demo.keyCallback)

	// Enable vsync
	glfw.SwapInterval(1)

	fmt.Println("Controls:")
	fmt.Println("1/2 - Switch between point-to-quad and wireframe demos")
	fmt.Println("+/- - Increase/decrease point size")
	fmt.Println("ESC - Exit")

	// Main loop
	for !window.ShouldClose() {
		glfw.PollEvents()

		demo.Update(float32(glfw.GetTime()))
		demo.Render()

		window.SwapBuffers()
	}
}

func NewDemo() (*Demo, error) {
	demo := &Demo{
		currentDemo:    0,
		pointSize:      0.2,
		animationSpeed: 1.0,
	}

	// Create pipeline
	demo.renderPipeline = pipeline.New()

	// Set up matrices
	demo.projection = mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/float32(windowHeight), 0.1, 100.0)
	demo.view = mgl32.LookAtV(mgl32.Vec3{0, 0, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	// Create point-to-quad shader program
	if err := demo.createPointProgram(); err != nil {
		return nil, err
	}

	// Create wireframe shader program
	if err := demo.createWireframeProgram(); err != nil {
		return nil, err
	}

	// Create meshes
	if err := demo.createMeshes(); err != nil {
		return nil, err
	}

	return demo, nil
}

func (d *Demo) createPointProgram() error {
	vertexShader, err := shader.CompileShader(vertexShaderSource, shader.VertexShader)
	if err != nil {
		return fmt.Errorf("vertex shader: %w", err)
	}
	defer vertexShader.Delete()

	geometryShader, err := shader.CompileShader(geometryShaderSource, shader.GeometryShader)
	if err != nil {
		return fmt.Errorf("geometry shader: %w", err)
	}
	defer geometryShader.Delete()

	fragmentShader, err := shader.CompileShader(fragmentShaderSource, shader.FragmentShader)
	if err != nil {
		return fmt.Errorf("fragment shader: %w", err)
	}
	defer fragmentShader.Delete()

	d.pointProgram, err = shader.CreateProgram(vertexShader, geometryShader, fragmentShader)
	if err != nil {
		return fmt.Errorf("program: %w", err)
	}

	return nil
}

func (d *Demo) createWireframeProgram() error {
	vertexShader, err := shader.CompileShader(vertexShaderSource, shader.VertexShader)
	if err != nil {
		return fmt.Errorf("vertex shader: %w", err)
	}
	defer vertexShader.Delete()

	geometryShader, err := shader.CompileShader(wireframeGeometrySource, shader.GeometryShader)
	if err != nil {
		return fmt.Errorf("geometry shader: %w", err)
	}
	defer geometryShader.Delete()

	fragmentShader, err := shader.CompileShader(wireframeFragmentSource, shader.FragmentShader)
	if err != nil {
		return fmt.Errorf("fragment shader: %w", err)
	}
	defer fragmentShader.Delete()

	d.wireframeProgram, err = shader.CreateProgram(vertexShader, geometryShader, fragmentShader)
	if err != nil {
		return fmt.Errorf("program: %w", err)
	}

	return nil
}

func (d *Demo) createMeshes() error {
	// Create point mesh
	layout := resource.NewVertexLayout().
		AddFloat(0, 3). // Position
		AddFloat(1, 3)  // Color

	var err error
	d.pointMesh, err = resource.NewMesh(points, nil, layout)
	if err != nil {
		return fmt.Errorf("point mesh: %w", err)
	}

	// Create triangle mesh for wireframe demo
	triangleVertices := []float32{
		// Triangle 1
		-0.5, -0.3, 0.0, 1.0, 0.0, 0.0,
		 0.5, -0.3, 0.0, 0.0, 1.0, 0.0,
		 0.0,  0.3, 0.0, 0.0, 0.0, 1.0,
	}

	d.triangleMesh, err = resource.NewMesh(triangleVertices, nil, layout)
	if err != nil {
		return fmt.Errorf("triangle mesh: %w", err)
	}

	return nil
}

func (d *Demo) Update(time float32) {
	// Animation is handled in render for this demo
}

func (d *Demo) Render() {
	// Clear screen
	d.renderPipeline.SetClearColor(0.1, 0.1, 0.1, 1.0)
	d.renderPipeline.Clear(true, true, false)

	// Set up pipeline state
	state := pipeline.NewBuilder().
		WithDepthTest(true, true, pipeline.DepthLess).
		WithViewport(0, 0, windowWidth, windowHeight).
		Build()

	d.renderPipeline.SetState(state)

	// Render current demo
	switch d.currentDemo {
	case 0:
		d.renderPointDemo()
	case 1:
		d.renderWireframeDemo()
	}
}

func (d *Demo) renderPointDemo() {
	d.renderPipeline.SetProgram(d.pointProgram)

	// Set uniforms
	modelLoc := d.pointProgram.GetUniformLocation("uModel")
	viewLoc := d.pointProgram.GetUniformLocation("uView")
	projLoc := d.pointProgram.GetUniformLocation("uProjection")
	pointSizeLoc := d.pointProgram.GetUniformLocation("uPointSize")

	// Animate rotation
	time := float32(glfw.GetTime()) * d.animationSpeed
	model := mgl32.HomogRotate3DZ(time)

	d.pointProgram.SetUniformMatrix4fv(modelLoc, &model)
	d.pointProgram.SetUniformMatrix4fv(viewLoc, &d.view)
	d.pointProgram.SetUniformMatrix4fv(projLoc, &d.projection)
	d.pointProgram.SetUniform1f(pointSizeLoc, d.pointSize)

	// Draw points (geometry shader will expand to quads)
	d.pointMesh.VAO.Draw(gl.POINTS, 3, 0)
}

func (d *Demo) renderWireframeDemo() {
	d.renderPipeline.SetProgram(d.wireframeProgram)

	// Set uniforms
	modelLoc := d.wireframeProgram.GetUniformLocation("uModel")
	viewLoc := d.wireframeProgram.GetUniformLocation("uView")
	projLoc := d.wireframeProgram.GetUniformLocation("uProjection")

	// Animate rotation
	time := float32(glfw.GetTime()) * d.animationSpeed
	model := mgl32.HomogRotate3DY(time)

	d.wireframeProgram.SetUniformMatrix4fv(modelLoc, &model)
	d.wireframeProgram.SetUniformMatrix4fv(viewLoc, &d.view)
	d.wireframeProgram.SetUniformMatrix4fv(projLoc, &d.projection)

	// Draw triangles (geometry shader will convert to wireframe)
	d.triangleMesh.Draw(gl.TRIANGLES)
}

func (d *Demo) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		switch key {
		case glfw.Key1:
			d.currentDemo = 0
			fmt.Println("Point-to-quad demo")
		case glfw.Key2:
			d.currentDemo = 1
			fmt.Println("Wireframe demo")
		case glfw.KeyEqual, glfw.KeyKPAdd: // + key
			d.pointSize += 0.05
			if d.pointSize > 1.0 {
				d.pointSize = 1.0
			}
			fmt.Printf("Point size: %.2f\n", d.pointSize)
		case glfw.KeyMinus, glfw.KeyKPSubtract: // - key
			d.pointSize -= 0.05
			if d.pointSize < 0.05 {
				d.pointSize = 0.05
			}
			fmt.Printf("Point size: %.2f\n", d.pointSize)
		case glfw.KeyEscape:
			w.SetShouldClose(true)
		}
	}
}

func (d *Demo) Cleanup() {
	if d.pointProgram != nil {
		d.pointProgram.Delete()
	}
	if d.wireframeProgram != nil {
		d.wireframeProgram.Delete()
	}
	if d.pointMesh != nil {
		d.pointMesh.Delete()
	}
	if d.triangleMesh != nil {
		d.triangleMesh.Delete()
	}
}