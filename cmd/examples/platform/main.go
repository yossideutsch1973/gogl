package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yossideutsch1973/gogl/internal/platform"
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

	// Configure GLFW for maximum compatibility
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False) // Hidden window for detection only

	// Create window
	window, err := glfw.CreateWindow(100, 100, "Platform Detection", nil, nil)
	if err != nil {
		log.Fatal("Failed to create window:", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	// Create platform detector
	detector := platform.New()

	// Detect system capabilities
	info, err := detector.Detect()
	if err != nil {
		log.Fatal("Failed to detect platform:", err)
	}

	// Print detailed information
	detector.PrintInfo()

	// Get recommended settings
	settings := detector.GetRecommendedSettings()
	fmt.Println("=== Recommended Settings ===")
	for key, value := range settings {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Demo feature compatibility checking
	fmt.Println("\n=== Feature Compatibility Demo ===")

	features := []struct {
		name      string
		supported bool
		required  string
	}{
		{"Basic Rendering", true, "OpenGL 2.0+"},
		{"Vertex Array Objects", info.Capabilities.SupportsVAO, "OpenGL 3.0+"},
		{"Uniform Buffer Objects", info.Capabilities.SupportsUniformBuffers, "OpenGL 3.1+"},
		{"Geometry Shaders", info.Capabilities.SupportsGeometryShaders, "OpenGL 3.2+"},
		{"Tessellation Shaders", info.Capabilities.SupportsTessellation, "OpenGL 4.0+"},
		{"Compute Shaders", info.Capabilities.SupportsComputeShaders, "OpenGL 4.3+"},
		{"Shader Storage Buffers", info.Capabilities.SupportsShaderStorageBuffers, "OpenGL 4.3+"},
	}

	for _, feature := range features {
		status := "❌ Not Supported"
		if feature.supported {
			status = "✅ Supported"
		}
		fmt.Printf("%-25s %s (%s)\n", feature.name+":", status, feature.required)
	}

	// Provide specific recommendations
	fmt.Println("\n=== Development Recommendations ===")

	if info.Platform == platform.PlatformMacOS {
		fmt.Println("🍎 macOS Development:")
		fmt.Println("  • Target OpenGL 4.1 maximum")
		fmt.Println("  • Avoid compute shaders")
		fmt.Println("  • Consider Metal for production")
		fmt.Println("  • Test on both Intel and Apple Silicon")
	}

	if info.Vendor == platform.VendorIntel {
		fmt.Println("💻 Intel Graphics:")
		fmt.Println("  • Use conservative memory allocation")
		fmt.Println("  • Avoid very large textures")
		fmt.Println("  • Test performance carefully")
	}

	if info.OpenGLVersion.Compare(platform.OpenGLVersion{Major: 4, Minor: 0}) < 0 {
		fmt.Println("⚠️  Older OpenGL Version:")
		fmt.Println("  • Some modern features unavailable")
		fmt.Println("  • Consider fallback rendering paths")
		fmt.Println("  • Update graphics drivers if possible")
	}

	if info.Capabilities.SupportsComputeShaders {
		fmt.Println("🚀 Compute Shaders Available:")
		fmt.Printf("  • Max work group size: %dx%dx%d\n",
			info.Capabilities.MaxWorkGroupSize[0],
			info.Capabilities.MaxWorkGroupSize[1],
			info.Capabilities.MaxWorkGroupSize[2])
		fmt.Printf("  • Max invocations: %d\n", info.Capabilities.MaxWorkGroupInvocations)
	} else {
		fmt.Println("⚡ Compute Shaders Not Available:")
		fmt.Println("  • Use vertex/fragment shaders for GPU computation")
		fmt.Println("  • Consider CPU-based alternatives")
	}

	fmt.Println("\n=== Example Code Generation ===")
	fmt.Println("Based on your system, here's recommended initialization code:")
	fmt.Println()

	if info.Platform == platform.PlatformMacOS {
		fmt.Println("// macOS-optimized initialization")
		fmt.Println("glfw.WindowHint(glfw.ContextVersionMajor, 4)")
		fmt.Println("glfw.WindowHint(glfw.ContextVersionMinor, 1)")
		fmt.Println("glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)")
	} else {
		fmt.Println("// General OpenGL initialization")
		fmt.Println("glfw.WindowHint(glfw.ContextVersionMajor, 4)")
		fmt.Println("glfw.WindowHint(glfw.ContextVersionMinor, 6)")
	}

	if info.Capabilities.SupportsVAO {
		fmt.Println("// VAO usage recommended")
		fmt.Println("vao, _ := resource.NewVertexArray()")
	}

	if !info.Capabilities.SupportsComputeShaders {
		fmt.Println("// Compute shaders not available - use alternatives")
		fmt.Println("// Consider vertex shader transform feedback or CPU computation")
	}
}
