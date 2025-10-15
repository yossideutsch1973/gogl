package main

import (
"fmt"
"log"
"os"
"path/filepath"
"runtime"

"github.com/go-gl/gl/v4.1-core/gl"
"github.com/go-gl/glfw/v3.3/glfw"
"github.com/yossideutsch/gogl/pkg/shader"
)

func init() {
runtime.LockOSThread()
}

func main() {
// Initialize GLFW
if err := glfw.Init(); err != nil {
log.Fatal("Failed to initialize GLFW:", err)
}
defer glfw.Terminate()

// Create window hints for OpenGL 4.1
glfw.WindowHint(glfw.ContextVersionMajor, 4)
glfw.WindowHint(glfw.ContextVersionMinor, 1)
glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
glfw.WindowHint(glfw.Visible, glfw.False) // Hidden window for testing

// Create window
window, err := glfw.CreateWindow(800, 600, "Shader Test", nil, nil)
if err != nil {
log.Fatal("Failed to create window:", err)
}
defer window.Destroy()

window.MakeContextCurrent()

// Initialize OpenGL
if err := gl.Init(); err != nil {
log.Fatal("Failed to initialize OpenGL:", err)
}

fmt.Println("Testing GoGL Shader Library")
fmt.Printf("OpenGL Version: %s\n\n", gl.GoStr(gl.GetString(gl.VERSION)))

// Test shader pairs
shaderPairs := []struct {
name     string
vertex   string
fragment string
}{
{"Basic", "shaders/vertex/basic.vert", "shaders/fragment/basic.frag"},
{"Flat Color", "shaders/vertex/flat_color.vert", "shaders/fragment/flat_color.frag"},
{"Phong Lighting", "shaders/vertex/phong.vert", "shaders/fragment/phong.frag"},
{"Textured", "shaders/vertex/textured.vert", "shaders/fragment/textured.frag"},
{"Skybox", "shaders/vertex/skybox.vert", "shaders/fragment/skybox.frag"},
{"Screen Quad + Blur", "shaders/vertex/screen_quad.vert", "shaders/fragment/blur.frag"},
{"Screen Quad + Grayscale", "shaders/vertex/screen_quad.vert", "shaders/fragment/grayscale.frag"},
{"Standard", "shaders/vertex/standard.vert", "shaders/fragment/simple_texture.frag"},
}

passCount := 0
failCount := 0

for _, pair := range shaderPairs {
fmt.Printf("Testing: %s\n", pair.name)

// Test vertex shader
vertexShader, err := shader.CompileShaderFromFile(pair.vertex, shader.VertexShader)
if err != nil {
fmt.Printf("  ❌ Vertex shader failed: %v\n", err)
failCount++
continue
}
defer vertexShader.Delete()

// Test fragment shader
fragmentShader, err := shader.CompileShaderFromFile(pair.fragment, shader.FragmentShader)
if err != nil {
fmt.Printf("  ❌ Fragment shader failed: %v\n", err)
failCount++
continue
}
defer fragmentShader.Delete()

// Test program linking
program, err := shader.CreateProgram(vertexShader, fragmentShader)
if err != nil {
fmt.Printf("  ❌ Program linking failed: %v\n", err)
failCount++
continue
}
program.Delete()

fmt.Printf("  ✅ Success\n")
passCount++
}

// Test geometry shaders
fmt.Println("\nTesting Geometry Shaders:")
geometryTests := []struct {
name     string
geometry string
}{
{"Point to Quad", "shaders/geometry/point_to_quad.glsl"},
{"Wireframe", "shaders/geometry/wireframe.glsl"},
{"Normal Visualization", "shaders/geometry/normal_visualization.glsl"},
{"Normal Lines", "shaders/geometry/normal_lines.glsl"},
{"Explode", "shaders/geometry/explode.glsl"},
}

for _, test := range geometryTests {
fmt.Printf("Testing: %s\n", test.name)

// Need a basic vertex shader
vertexShader, err := shader.CompileShaderFromFile("shaders/vertex/basic.vert", shader.VertexShader)
if err != nil {
fmt.Printf("  ❌ Vertex shader failed: %v\n", err)
failCount++
continue
}
defer vertexShader.Delete()

// Test geometry shader
geometryShader, err := shader.CompileShaderFromFile(test.geometry, shader.GeometryShader)
if err != nil {
fmt.Printf("  ❌ Geometry shader failed: %v\n", err)
failCount++
continue
}
defer geometryShader.Delete()

// Need a basic fragment shader
fragmentShader, err := shader.CompileShaderFromFile("shaders/fragment/flat_color.frag", shader.FragmentShader)
if err != nil {
fmt.Printf("  ❌ Fragment shader failed: %v\n", err)
failCount++
continue
}
defer fragmentShader.Delete()

// Test program linking
program, err := shader.CreateProgram(vertexShader, geometryShader, fragmentShader)
if err != nil {
fmt.Printf("  ❌ Program linking failed: %v\n", err)
failCount++
continue
}
program.Delete()

fmt.Printf("  ✅ Success\n")
passCount++
}

// Test additional post-processing shaders
fmt.Println("\nTesting Post-Processing Shaders:")
postProcessTests := []string{
"edge_detection.frag",
"invert.frag",
"brightness_contrast.frag",
"gamma_correction.frag",
}

for _, fragName := range postProcessTests {
fmt.Printf("Testing: %s\n", fragName)

vertexShader, err := shader.CompileShaderFromFile("shaders/vertex/screen_quad.vert", shader.VertexShader)
if err != nil {
fmt.Printf("  ❌ Vertex shader failed: %v\n", err)
failCount++
continue
}
defer vertexShader.Delete()

fragmentShader, err := shader.CompileShaderFromFile(filepath.Join("shaders/fragment", fragName), shader.FragmentShader)
if err != nil {
fmt.Printf("  ❌ Fragment shader failed: %v\n", err)
failCount++
continue
}
defer fragmentShader.Delete()

program, err := shader.CreateProgram(vertexShader, fragmentShader)
if err != nil {
fmt.Printf("  ❌ Program linking failed: %v\n", err)
failCount++
continue
}
program.Delete()

fmt.Printf("  ✅ Success\n")
passCount++
}

// Summary
fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
fmt.Printf("Test Summary: %d passed, %d failed\n", passCount, failCount)

if failCount > 0 {
os.Exit(1)
}
}
