package platform

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// GPUVendor represents the GPU vendor
type GPUVendor int

const (
	VendorUnknown GPUVendor = iota
	VendorNVIDIA
	VendorAMD
	VendorIntel
	VendorApple
)

func (v GPUVendor) String() string {
	switch v {
	case VendorNVIDIA:
		return "NVIDIA"
	case VendorAMD:
		return "AMD"
	case VendorIntel:
		return "Intel"
	case VendorApple:
		return "Apple"
	default:
		return "Unknown"
	}
}

// Platform represents the current platform
type Platform int

const (
	PlatformUnknown Platform = iota
	PlatformWindows
	PlatformLinux
	PlatformMacOS
)

func (p Platform) String() string {
	switch p {
	case PlatformWindows:
		return "Windows"
	case PlatformLinux:
		return "Linux"
	case PlatformMacOS:
		return "macOS"
	default:
		return "Unknown"
	}
}

// OpenGLVersion represents an OpenGL version
type OpenGLVersion struct {
	Major int
	Minor int
}

func (v OpenGLVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// Compare compares this version with another
func (v OpenGLVersion) Compare(other OpenGLVersion) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	return v.Minor - other.Minor
}

// IsAtLeast checks if this version is at least the specified version
func (v OpenGLVersion) IsAtLeast(major, minor int) bool {
	return v.Compare(OpenGLVersion{major, minor}) >= 0
}

// Capabilities represents platform-specific OpenGL capabilities
type Capabilities struct {
	MaxTextureSize          int32
	MaxTextureUnits         int32
	MaxVertexAttributes     int32
	MaxUniformBufferBindings int32
	MaxWorkGroupSize        [3]int32
	MaxWorkGroupInvocations int32
	
	// Feature support
	SupportsGeometryShaders    bool
	SupportsComputeShaders     bool
	SupportsTessellation       bool
	SupportsTextureArrays      bool
	SupportsUniformBuffers     bool
	SupportsShaderStorageBuffers bool
	SupportsInstancedRendering bool
	SupportsVAO               bool
	SupportsDebugCallback     bool
}

// SystemInfo contains complete system and OpenGL information
type SystemInfo struct {
	Platform        Platform
	OpenGLVersion   OpenGLVersion
	GLSLVersion     OpenGLVersion
	Vendor          GPUVendor
	VendorString    string
	RendererString  string
	Capabilities    Capabilities
	
	// Platform-specific notes
	Notes []string
}

// Detector handles platform detection and capability queries
type Detector struct {
	info *SystemInfo
}

// New creates a new platform detector
func New() *Detector {
	return &Detector{}
}

// Detect analyzes the current platform and OpenGL context
func (d *Detector) Detect() (*SystemInfo, error) {
	if d.info != nil {
		return d.info, nil
	}

	info := &SystemInfo{}

	// Detect platform
	info.Platform = d.detectPlatform()

	// Get OpenGL version
	var err error
	info.OpenGLVersion, err = d.detectOpenGLVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to detect OpenGL version: %w", err)
	}

	// Get GLSL version
	info.GLSLVersion, err = d.detectGLSLVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to detect GLSL version: %w", err)
	}

	// Get vendor and renderer
	info.VendorString = gl.GoStr(gl.GetString(gl.VENDOR))
	info.RendererString = gl.GoStr(gl.GetString(gl.RENDERER))
	info.Vendor = d.detectVendor(info.VendorString, info.RendererString)

	// Query capabilities
	info.Capabilities = d.queryCapabilities(info.OpenGLVersion)

	// Add platform-specific notes
	info.Notes = d.generateNotes(info)

	d.info = info
	return info, nil
}

func (d *Detector) detectPlatform() Platform {
	switch runtime.GOOS {
	case "windows":
		return PlatformWindows
	case "linux":
		return PlatformLinux
	case "darwin":
		return PlatformMacOS
	default:
		return PlatformUnknown
	}
}

func (d *Detector) detectOpenGLVersion() (OpenGLVersion, error) {
	versionStr := gl.GoStr(gl.GetString(gl.VERSION))
	
	// Parse version string (e.g., "4.1 Metal - 89.4" or "4.6.0")
	parts := strings.Fields(versionStr)
	if len(parts) == 0 {
		return OpenGLVersion{}, fmt.Errorf("empty version string")
	}

	versionPart := parts[0]
	dotIndex := strings.Index(versionPart, ".")
	if dotIndex == -1 {
		return OpenGLVersion{}, fmt.Errorf("invalid version format: %s", versionPart)
	}

	majorStr := versionPart[:dotIndex]
	minorStr := versionPart[dotIndex+1:]

	// Handle cases like "4.6.0" where we want just "4.6"
	if dotIndex2 := strings.Index(minorStr, "."); dotIndex2 != -1 {
		minorStr = minorStr[:dotIndex2]
	}

	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return OpenGLVersion{}, fmt.Errorf("invalid major version: %s", majorStr)
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return OpenGLVersion{}, fmt.Errorf("invalid minor version: %s", minorStr)
	}

	return OpenGLVersion{major, minor}, nil
}

func (d *Detector) detectGLSLVersion() (OpenGLVersion, error) {
	versionStr := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	
	// Parse GLSL version (e.g., "4.10" or "4.60")
	parts := strings.Fields(versionStr)
	if len(parts) == 0 {
		return OpenGLVersion{}, fmt.Errorf("empty GLSL version string")
	}

	versionPart := parts[0]
	dotIndex := strings.Index(versionPart, ".")
	if dotIndex == -1 {
		return OpenGLVersion{}, fmt.Errorf("invalid GLSL version format: %s", versionPart)
	}

	majorStr := versionPart[:dotIndex]
	minorStr := versionPart[dotIndex+1:]

	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return OpenGLVersion{}, fmt.Errorf("invalid GLSL major version: %s", majorStr)
	}

	// GLSL minor version is usually two digits (e.g., "10", "60")
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return OpenGLVersion{}, fmt.Errorf("invalid GLSL minor version: %s", minorStr)
	}

	return OpenGLVersion{major, minor}, nil
}

func (d *Detector) detectVendor(vendorStr, rendererStr string) GPUVendor {
	vendorLower := strings.ToLower(vendorStr)
	rendererLower := strings.ToLower(rendererStr)

	if strings.Contains(vendorLower, "nvidia") || strings.Contains(rendererLower, "nvidia") {
		return VendorNVIDIA
	}
	if strings.Contains(vendorLower, "amd") || strings.Contains(rendererLower, "amd") || 
	   strings.Contains(vendorLower, "ati") || strings.Contains(rendererLower, "radeon") {
		return VendorAMD
	}
	if strings.Contains(vendorLower, "intel") || strings.Contains(rendererLower, "intel") {
		return VendorIntel
	}
	if strings.Contains(vendorLower, "apple") || strings.Contains(rendererLower, "apple") {
		return VendorApple
	}

	return VendorUnknown
}

func (d *Detector) queryCapabilities(version OpenGLVersion) Capabilities {
	caps := Capabilities{}

	// Query basic limits
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &caps.MaxTextureSize)
	gl.GetIntegerv(gl.MAX_TEXTURE_IMAGE_UNITS, &caps.MaxTextureUnits)
	gl.GetIntegerv(gl.MAX_VERTEX_ATTRIBS, &caps.MaxVertexAttributes)

	// Feature support based on OpenGL version AND Go library limitations
	// NOTE: This go-gl library is compiled for OpenGL 4.1 core, so we're limited
	// to 4.1 features regardless of what the system reports
	effectiveVersion := version
	if version.Compare(OpenGLVersion{4, 1}) > 0 {
		// System supports higher than 4.1, but Go library is limited to 4.1
		effectiveVersion = OpenGLVersion{4, 1}
	}

	caps.SupportsVAO = effectiveVersion.IsAtLeast(3, 0)
	caps.SupportsTextureArrays = effectiveVersion.IsAtLeast(3, 0)
	caps.SupportsUniformBuffers = effectiveVersion.IsAtLeast(3, 1)
	caps.SupportsInstancedRendering = effectiveVersion.IsAtLeast(3, 1)
	caps.SupportsGeometryShaders = effectiveVersion.IsAtLeast(3, 2)
	caps.SupportsTessellation = effectiveVersion.IsAtLeast(4, 0)
	
	// These require OpenGL 4.3+ which is not available in go-gl v4.1-core
	caps.SupportsComputeShaders = false // Always false due to library limitation
	caps.SupportsShaderStorageBuffers = false // Always false due to library limitation  
	caps.SupportsDebugCallback = effectiveVersion.IsAtLeast(4, 3) // This might work in 4.1

	// Query additional limits if supported
	if caps.SupportsUniformBuffers {
		gl.GetIntegerv(gl.MAX_UNIFORM_BUFFER_BINDINGS, &caps.MaxUniformBufferBindings)
	}

	if caps.SupportsComputeShaders {
		gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 0, &caps.MaxWorkGroupSize[0])
		gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 1, &caps.MaxWorkGroupSize[1])
		gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 2, &caps.MaxWorkGroupSize[2])
		gl.GetIntegerv(gl.MAX_COMPUTE_WORK_GROUP_INVOCATIONS, &caps.MaxWorkGroupInvocations)
	}

	return caps
}

func (d *Detector) generateNotes(info *SystemInfo) []string {
	var notes []string

	// Go library limitation note
	if info.OpenGLVersion.Compare(OpenGLVersion{4, 1}) > 0 {
		notes = append(notes, "NOTE: System supports "+info.OpenGLVersion.String()+" but Go library limited to OpenGL 4.1 core")
		notes = append(notes, "Compute shaders and other 4.3+ features are not available through this library")
	}

	// macOS-specific notes
	if info.Platform == PlatformMacOS {
		notes = append(notes, "OpenGL is deprecated on macOS since 2018")
		notes = append(notes, "Maximum supported version is OpenGL 4.1")
		
		if info.Vendor == VendorApple {
			notes = append(notes, "Using Apple Silicon GPU with Metal backend")
		}
		
		notes = append(notes, "Consider using Metal for production applications on macOS")
	}

	// Version-specific warnings
	if info.OpenGLVersion.Compare(OpenGLVersion{3, 3}) < 0 {
		notes = append(notes, "WARNING: OpenGL version is quite old, consider updating drivers")
	}

	if info.OpenGLVersion.Compare(OpenGLVersion{4, 0}) < 0 {
		notes = append(notes, "Some modern features may not be available")
	}

	// Vendor-specific notes
	switch info.Vendor {
	case VendorIntel:
		notes = append(notes, "Intel integrated graphics may have limited performance")
	case VendorNVIDIA:
		notes = append(notes, "NVIDIA GPU detected - excellent OpenGL support expected")
	case VendorAMD:
		notes = append(notes, "AMD GPU detected - good OpenGL support expected")
	}

	return notes
}

// GetRecommendedSettings returns recommended settings based on the detected platform
func (d *Detector) GetRecommendedSettings() map[string]interface{} {
	if d.info == nil {
		return nil
	}

	settings := make(map[string]interface{})

	// General recommendations
	settings["useVAO"] = d.info.Capabilities.SupportsVAO
	settings["useUniformBuffers"] = d.info.Capabilities.SupportsUniformBuffers
	settings["useInstancedRendering"] = d.info.Capabilities.SupportsInstancedRendering

	// Platform-specific recommendations
	switch d.info.Platform {
	case PlatformMacOS:
		settings["targetOpenGLVersion"] = "4.1"
		settings["useForwardCompatible"] = true
		settings["avoidComputeShaders"] = !d.info.Capabilities.SupportsComputeShaders
		settings["preferMetal"] = true
	case PlatformWindows, PlatformLinux:
		settings["targetOpenGLVersion"] = "4.6"
		settings["useForwardCompatible"] = false
		settings["useComputeShaders"] = d.info.Capabilities.SupportsComputeShaders
	}

	// Vendor-specific recommendations
	switch d.info.Vendor {
	case VendorIntel:
		settings["conservativeMemoryUsage"] = true
		settings["avoidLargeTextures"] = true
	case VendorNVIDIA, VendorAMD:
		settings["aggressiveOptimizations"] = true
		settings["largeTexturesOK"] = true
	}

	return settings
}

// PrintInfo prints detailed system information
func (d *Detector) PrintInfo() {
	if d.info == nil {
		fmt.Println("No system information available. Call Detect() first.")
		return
	}

	info := d.info

	fmt.Println("=== GoGL System Information ===")
	fmt.Printf("Platform: %s\n", info.Platform)
	fmt.Printf("OpenGL Version: %s\n", info.OpenGLVersion)
	fmt.Printf("GLSL Version: %s\n", info.GLSLVersion)
	fmt.Printf("Vendor: %s (%s)\n", info.Vendor, info.VendorString)
	fmt.Printf("Renderer: %s\n", info.RendererString)
	
	fmt.Println("\n=== Capabilities ===")
	fmt.Printf("Max Texture Size: %d\n", info.Capabilities.MaxTextureSize)
	fmt.Printf("Max Texture Units: %d\n", info.Capabilities.MaxTextureUnits)
	fmt.Printf("Max Vertex Attributes: %d\n", info.Capabilities.MaxVertexAttributes)
	
	fmt.Println("\n=== Feature Support ===")
	fmt.Printf("Vertex Array Objects: %v\n", info.Capabilities.SupportsVAO)
	fmt.Printf("Geometry Shaders: %v\n", info.Capabilities.SupportsGeometryShaders)
	fmt.Printf("Compute Shaders: %v\n", info.Capabilities.SupportsComputeShaders)
	fmt.Printf("Tessellation: %v\n", info.Capabilities.SupportsTessellation)
	fmt.Printf("Uniform Buffers: %v\n", info.Capabilities.SupportsUniformBuffers)
	fmt.Printf("Shader Storage Buffers: %v\n", info.Capabilities.SupportsShaderStorageBuffers)
	fmt.Printf("Instanced Rendering: %v\n", info.Capabilities.SupportsInstancedRendering)

	if len(info.Notes) > 0 {
		fmt.Println("\n=== Platform Notes ===")
		for _, note := range info.Notes {
			fmt.Printf("• %s\n", note)
		}
	}

	fmt.Println()
}