package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/yossideutsch/gogl/internal/platform"
	"github.com/yossideutsch/gogl/pkg/pipeline"
	"github.com/yossideutsch/gogl/pkg/resource"
	"github.com/yossideutsch/gogl/pkg/shader"
)

const (
	windowWidth  = 800
	windowHeight = 600
	numParticles = 1024
)

// Particle structure matching the compute shader
type Particle struct {
	Position [2]float32
	Velocity [2]float32
	Color    [4]float32
	Life     float32
	Size     float32
	_        [2]float32 // Padding to align to 16 bytes
}

var (
	computeShaderSource = `#version 430
// NOTE: Compute shaders require OpenGL 4.3+ / GLSL 4.30+
// This will only work if the Go OpenGL binding supports 4.3+

layout(local_size_x = 16, local_size_y = 1, local_size_z = 1) in;

// Particle data structure
struct Particle {
    vec2 position;
    vec2 velocity;
    vec4 color;
    float life;
    float size;
    vec2 padding;
};

// Shader storage buffer objects
layout(std430, binding = 0) restrict buffer ParticleBuffer {
    Particle particles[];
};

// Uniforms
uniform float uDeltaTime;
uniform float uGravity;
uniform vec2 uAttractor;
uniform float uAttractorStrength;
uniform vec2 uViewportSize;
uniform float uTime;

// Random number generation
uint hash(uint x) {
    x += (x << 10u);
    x ^= (x >> 6u);
    x += (x << 3u);
    x ^= (x >> 11u);
    x += (x << 15u);
    return x;
}

float random(uint seed) {
    return float(hash(seed)) / 4294967296.0;
}

void main() {
    uint index = gl_GlobalInvocationID.x;
    
    if (index >= particles.length()) {
        return;
    }
    
    Particle particle = particles[index];
    
    // Update life
    particle.life -= uDeltaTime;
    
    // If particle is dead, respawn it
    if (particle.life <= 0.0) {
        uint seed = index + uint(uTime * 1000.0);
        particle.position = vec2(
            uViewportSize.x * 0.5 + (random(seed) - 0.5) * 100.0,
            uViewportSize.y + random(seed + 1u) * 100.0
        );
        particle.velocity = vec2(
            (random(seed + 2u) - 0.5) * 200.0,
            -random(seed + 3u) * 150.0 - 50.0
        );
        particle.color = vec4(
            0.5 + random(seed + 4u) * 0.5,
            0.3 + random(seed + 5u) * 0.4,
            0.8 + random(seed + 6u) * 0.2,
            1.0
        );
        particle.life = 3.0 + random(seed + 7u) * 2.0;
        particle.size = 2.0 + random(seed + 8u) * 4.0;
    } else {
        // Apply gravity
        particle.velocity.y -= uGravity * uDeltaTime;
        
        // Apply attractor force
        vec2 toAttractor = uAttractor - particle.position;
        float distance = length(toAttractor);
        if (distance > 10.0) {
            vec2 force = normalize(toAttractor) * uAttractorStrength / (distance * 0.01);
            particle.velocity += force * uDeltaTime;
        }
        
        // Update position
        particle.position += particle.velocity * uDeltaTime;
        
        // Bounce off walls
        if (particle.position.x < 0.0 || particle.position.x > uViewportSize.x) {
            particle.velocity.x *= -0.8;
            particle.position.x = clamp(particle.position.x, 0.0, uViewportSize.x);
        }
        if (particle.position.y < 0.0) {
            particle.velocity.y *= -0.8;
            particle.position.y = 0.0;
        }
        
        // Fade out over time
        float lifeRatio = particle.life / 5.0;
        particle.color.a = lifeRatio;
    }
    
    particles[index] = particle;
}`

	renderVertexSource = `#version 410 core
layout(location = 0) in vec2 aPosition;
layout(location = 1) in vec2 aVelocity;
layout(location = 2) in vec4 aColor;
layout(location = 3) in float aLife;
layout(location = 4) in float aSize;

uniform mat4 uProjection;
uniform vec2 uViewportSize;

out vec4 vColor;
out float vSize;

void main() {
    vColor = aColor;
    vSize = aSize;
    
    // Convert from pixel coordinates to normalized coordinates
    vec2 normalizedPos = aPosition / uViewportSize * 2.0 - 1.0;
    normalizedPos.y = -normalizedPos.y; // Flip Y axis
    
    gl_Position = vec4(normalizedPos, 0.0, 1.0);
    gl_PointSize = aSize;
}`

	renderFragmentSource = `#version 410 core
in vec4 vColor;
in float vSize;
out vec4 fragColor;

void main() {
    // Create circular particles
    vec2 center = vec2(0.5, 0.5);
    float distance = length(gl_PointCoord - center);
    
    if (distance > 0.5) {
        discard;
    }
    
    // Add some glow effect
    float intensity = 1.0 - (distance * 2.0);
    intensity = pow(intensity, 2.0);
    
    fragColor = vec4(vColor.rgb * intensity, vColor.a * intensity);
}`
)

func init() {
	runtime.LockOSThread()
}

type ComputeDemo struct {
	computeProgram *shader.Program
	renderProgram  *shader.Program
	
	particleSSBO *resource.ShaderStorageBuffer
	particleVAO  *resource.VertexArray
	particleVBO  *resource.VertexBuffer
	
	renderPipeline *pipeline.Pipeline
	
	particles []Particle
	
	// Simulation parameters
	gravity           float32
	attractorStrength float32
	attractorPos      mgl32.Vec2
	
	// Mouse state
	mousePos mgl32.Vec2
}

func main() {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatal("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Configure GLFW - try for OpenGL 4.3+ first for compute shaders
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create window
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Compute Shader Particle System", nil, nil)
	if err != nil {
		// Fallback to 4.1 and show warning
		log.Println("Failed to create OpenGL 4.3 context, falling back to 4.1...")
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		
		window, err = glfw.CreateWindow(windowWidth, windowHeight, "Compute Shader Demo (CPU Fallback)", nil, nil)
		if err != nil {
			log.Fatal("Failed to create window:", err)
		}
	}
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	fmt.Printf("OpenGL version: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))

	// Detect platform capabilities
	detector := platform.New()
	sysInfo, err := detector.Detect()
	if err != nil {
		log.Fatal("Failed to detect platform:", err)
	}

	// Check if compute shaders are supported
	if !sysInfo.Capabilities.SupportsComputeShaders {
		fmt.Println("\n⚠️  WARNING: Compute shaders not supported on this platform!")
		fmt.Printf("Platform: %s, OpenGL: %s\n", sysInfo.Platform, sysInfo.OpenGLVersion)
		fmt.Println("Required: OpenGL 4.3+ for compute shaders")
		for _, note := range sysInfo.Notes {
			fmt.Printf("• %s\n", note)
		}
		fmt.Println("\nRunning CPU-based particle simulation instead...")
		
		// Run CPU fallback demo
		runCPUParticleDemo(window)
		return
	}

	// Create GPU compute demo
	demo, err := NewComputeDemo()
	if err != nil {
		fmt.Printf("⚠️  Failed to create GPU demo, falling back to CPU: %v\n", err)
		runCPUParticleDemo(window)
		return
	}
	defer demo.Cleanup()

	// Set up input handling
	window.SetCursorPosCallback(demo.mouseCallback)
	window.SetKeyCallback(demo.keyCallback)

	// Enable vsync
	glfw.SwapInterval(1)

	fmt.Println("Controls:")
	fmt.Println("Mouse - Move attractor")
	fmt.Println("ESC - Exit")

	lastTime := glfw.GetTime()

	// Main loop
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		glfw.PollEvents()

		demo.Update(deltaTime, float32(currentTime))
		demo.Render()

		window.SwapBuffers()
	}
}

func NewComputeDemo() (*ComputeDemo, error) {
	demo := &ComputeDemo{
		gravity:           200.0,
		attractorStrength: 50000.0,
		attractorPos:      mgl32.Vec2{windowWidth / 2, windowHeight / 2},
		mousePos:          mgl32.Vec2{windowWidth / 2, windowHeight / 2},
	}

	// Create pipeline
	demo.renderPipeline = pipeline.New()

	// Initialize particles
	demo.initParticles()

	// Create compute shader
	if err := demo.createComputeShader(); err != nil {
		return nil, fmt.Errorf("compute shader: %w", err)
	}

	// Create render shader
	if err := demo.createRenderShader(); err != nil {
		return nil, fmt.Errorf("render shader: %w", err)
	}

	// Create buffers
	if err := demo.createBuffers(); err != nil {
		return nil, fmt.Errorf("buffers: %w", err)
	}

	return demo, nil
}

func (d *ComputeDemo) initParticles() {
	d.particles = make([]Particle, numParticles)
	
	for i := range d.particles {
		d.particles[i] = Particle{
			Position: [2]float32{float32(windowWidth / 2), float32(windowHeight + 100)},
			Velocity: [2]float32{0, -100},
			Color:    [4]float32{1, 1, 1, 1},
			Life:     0, // Start dead so they get respawned
			Size:     3.0,
		}
	}
}

func (d *ComputeDemo) createComputeShader() error {
	computeShader, err := shader.CompileShader(computeShaderSource, shader.ComputeShader)
	if err != nil {
		return err
	}
	defer computeShader.Delete()

	d.computeProgram, err = shader.CreateProgram(computeShader)
	return err
}

func (d *ComputeDemo) createRenderShader() error {
	vertexShader, err := shader.CompileShader(renderVertexSource, shader.VertexShader)
	if err != nil {
		return err
	}
	defer vertexShader.Delete()

	fragmentShader, err := shader.CompileShader(renderFragmentSource, shader.FragmentShader)
	if err != nil {
		return err
	}
	defer fragmentShader.Delete()

	d.renderProgram, err = shader.CreateProgram(vertexShader, fragmentShader)
	return err
}

func (d *ComputeDemo) createBuffers() error {
	// Create SSBO for particle data
	particleSize := int(unsafe.Sizeof(Particle{}))
	totalSize := particleSize * numParticles

	var err error
	d.particleSSBO, err = resource.NewShaderStorageBuffer(totalSize, resource.DynamicDraw)
	if err != nil {
		return err
	}

	// Upload initial particle data
	d.particleSSBO.Bind()
	d.particleSSBO.UpdateData(0, unsafe.Pointer(&d.particles[0]), totalSize)
	d.particleSSBO.BindBase(0) // Binding point 0

	// Create VAO for rendering
	d.particleVAO, err = resource.NewVertexArray()
	if err != nil {
		return err
	}

	// Use the SSBO as a VBO for rendering (this is a bit of a hack but works)
	d.particleVBO = &resource.VertexBuffer{Buffer: d.particleSSBO.Buffer}

	d.particleVAO.SetVertexBuffer(d.particleVBO)

	// Set up vertex attributes to match the particle structure
	stride := int32(unsafe.Sizeof(Particle{}))
	d.particleVAO.AddFloatAttribute(0, 2, stride, 0)                                      // Position
	d.particleVAO.AddFloatAttribute(1, 2, stride, uintptr(unsafe.Offsetof(d.particles[0].Velocity))) // Velocity
	d.particleVAO.AddFloatAttribute(2, 4, stride, uintptr(unsafe.Offsetof(d.particles[0].Color)))    // Color
	d.particleVAO.AddFloatAttribute(3, 1, stride, uintptr(unsafe.Offsetof(d.particles[0].Life)))     // Life
	d.particleVAO.AddFloatAttribute(4, 1, stride, uintptr(unsafe.Offsetof(d.particles[0].Size)))     // Size

	return nil
}

func (d *ComputeDemo) Update(deltaTime float32, time float32) {
	// Update attractor position to follow mouse
	d.attractorPos = d.mousePos

	// Run compute shader
	d.computeProgram.Use()

	// Set uniforms
	d.computeProgram.SetUniform1f(d.computeProgram.GetUniformLocation("uDeltaTime"), deltaTime)
	d.computeProgram.SetUniform1f(d.computeProgram.GetUniformLocation("uGravity"), d.gravity)
	d.computeProgram.SetUniform1f(d.computeProgram.GetUniformLocation("uAttractorStrength"), d.attractorStrength)
	d.computeProgram.SetUniform1f(d.computeProgram.GetUniformLocation("uTime"), time)

	attractorLoc := d.computeProgram.GetUniformLocation("uAttractor")
	d.computeProgram.SetUniform3f(attractorLoc, d.attractorPos.X(), d.attractorPos.Y(), 0)

	viewportLoc := d.computeProgram.GetUniformLocation("uViewportSize")
	d.computeProgram.SetUniform3f(viewportLoc, windowWidth, windowHeight, 0)

	// Dispatch compute shader
	workGroups := (numParticles + 15) / 16 // Round up division
	d.computeProgram.DispatchCompute(uint32(workGroups), 1, 1)

	// Memory barrier to ensure compute shader writes are visible to vertex shader
	d.computeProgram.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT | gl.VERTEX_ATTRIB_ARRAY_BARRIER_BIT)
}

func (d *ComputeDemo) Render() {
	// Clear screen
	d.renderPipeline.SetClearColor(0.1, 0.1, 0.15, 1.0)
	d.renderPipeline.Clear(true, true, false)

	// Set up render state
	state := pipeline.NewBuilder().
		WithBlending(true, pipeline.BlendSrcAlpha, pipeline.BlendOneMinusSrcAlpha).
		WithDepthTest(false, false, pipeline.DepthLess).
		WithProgram(d.renderProgram).
		WithViewport(0, 0, windowWidth, windowHeight).
		Build()

	d.renderPipeline.SetState(state)

	// Set render uniforms
	projLoc := d.renderProgram.GetUniformLocation("uProjection")
	projection := mgl32.Ortho(0, windowWidth, windowHeight, 0, -1, 1)
	d.renderProgram.SetUniformMatrix4fv(projLoc, &projection)

	viewportLoc := d.renderProgram.GetUniformLocation("uViewportSize")
	d.renderProgram.SetUniform3f(viewportLoc, windowWidth, windowHeight, 0)

	// Render particles
	gl.Enable(gl.PROGRAM_POINT_SIZE)
	d.particleVAO.Draw(gl.POINTS, numParticles, 0)
}

func (d *ComputeDemo) mouseCallback(w *glfw.Window, xpos, ypos float64) {
	d.mousePos[0] = float32(xpos)
	d.mousePos[1] = float32(ypos)
}

func (d *ComputeDemo) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}

func (d *ComputeDemo) Cleanup() {
	if d.computeProgram != nil {
		d.computeProgram.Delete()
	}
	if d.renderProgram != nil {
		d.renderProgram.Delete()
	}
	if d.particleSSBO != nil {
		d.particleSSBO.Delete()
	}
	if d.particleVAO != nil {
		d.particleVAO.Delete()
	}
}

// CPU-based particle demo for platforms without compute shader support
func runCPUParticleDemo(window *glfw.Window) {
	fmt.Println("Running CPU-based particle simulation...")
	
	// Set up basic rendering state
	gl.ClearColor(0.1, 0.1, 0.15, 1.0)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	// Simple vertex shader for CPU particles
	vertexSource := `#version 410 core
layout(location = 0) in vec2 aPosition;
layout(location = 1) in vec4 aColor;
layout(location = 2) in float aSize;

out vec4 vColor;

void main() {
    vColor = aColor;
    gl_Position = vec4(aPosition / vec2(400.0, 300.0), 0.0, 1.0);
    gl_PointSize = aSize;
}`

	fragmentSource := `#version 410 core
in vec4 vColor;
out vec4 fragColor;

void main() {
    vec2 center = vec2(0.5, 0.5);
    float distance = length(gl_PointCoord - center);
    
    if (distance > 0.5) {
        discard;
    }
    
    float intensity = 1.0 - (distance * 2.0);
    fragColor = vec4(vColor.rgb * intensity, vColor.a * intensity);
}`

	// Compile shaders
	vertexShader, err := shader.CompileShader(vertexSource, shader.VertexShader)
	if err != nil {
		log.Fatal("Failed to compile vertex shader:", err)
	}
	defer vertexShader.Delete()

	fragmentShader, err := shader.CompileShader(fragmentSource, shader.FragmentShader)
	if err != nil {
		log.Fatal("Failed to compile fragment shader:", err)
	}
	defer fragmentShader.Delete()

	program, err := shader.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatal("Failed to create program:", err)
	}
	defer program.Delete()

	// Create simple particle system
	type CPUParticle struct {
		X, Y     float32
		VX, VY   float32
		R, G, B, A float32
		Life     float32
		Size     float32
	}

	particles := make([]CPUParticle, 100)
	for i := range particles {
		particles[i] = CPUParticle{
			X: 400, Y: 600, VX: 0, VY: -100,
			R: 1, G: 1, B: 1, A: 1,
			Life: 0, Size: 3,
		}
	}

	// Create buffers
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	defer gl.DeleteVertexArrays(1, &vao)
	defer gl.DeleteBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	// Position attribute
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 7*4, nil)
	gl.EnableVertexAttribArray(0)

	// Color attribute
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 7*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	// Size attribute
	gl.VertexAttribPointer(2, 1, gl.FLOAT, false, 7*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			w.SetShouldClose(true)
		}
	})

	glfw.SwapInterval(1)
	lastTime := glfw.GetTime()

	// Main loop
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		glfw.PollEvents()

		// Update particles on CPU
		for i := range particles {
			p := &particles[i]
			
			p.Life -= deltaTime
			if p.Life <= 0 {
				// Respawn
				p.X = 400 + (rand.Float32()-0.5)*100
				p.Y = 600
				p.VX = (rand.Float32()-0.5)*200
				p.VY = -rand.Float32()*150 - 50
				p.Life = 3 + rand.Float32()*2
				p.R = 0.5 + rand.Float32()*0.5
				p.G = 0.3 + rand.Float32()*0.4
				p.B = 0.8 + rand.Float32()*0.2
				p.A = 1.0
			} else {
				// Update physics
				p.VY -= 200 * deltaTime // gravity
				p.X += p.VX * deltaTime
				p.Y += p.VY * deltaTime
				
				// Bounce off walls
				if p.X < 0 || p.X > 800 {
					p.VX *= -0.8
					if p.X < 0 { p.X = 0 }
					if p.X > 800 { p.X = 800 }
				}
				if p.Y < 0 {
					p.VY *= -0.8
					p.Y = 0
				}
				
				// Fade
				p.A = p.Life / 5.0
			}
		}

		// Render
		gl.Clear(gl.COLOR_BUFFER_BIT)
		program.Use()

		// Upload particle data
		vertexData := make([]float32, len(particles)*7)
		for i, p := range particles {
			vertexData[i*7+0] = p.X
			vertexData[i*7+1] = p.Y
			vertexData[i*7+2] = p.R
			vertexData[i*7+3] = p.G
			vertexData[i*7+4] = p.B
			vertexData[i*7+5] = p.A
			vertexData[i*7+6] = p.Size
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertexData)*4, gl.Ptr(vertexData), gl.DYNAMIC_DRAW)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.POINTS, 0, int32(len(particles)))

		window.SwapBuffers()
	}
}