# GoGL Examples

This directory contains example applications demonstrating the capabilities of the GoGL OpenGL shader library.

## Examples Overview

### 0. Shader Test (`shader_test/`)
- **Purpose**: Validates all shaders in the library compile and link correctly
- **Features**:
  - Tests 8 vertex/fragment shader pairs
  - Tests 5 geometry shaders
  - Tests 4 post-processing shaders
  - Reports pass/fail for each shader
- **OpenGL Requirements**: 4.1+
- **Run**: `go run cmd/examples/shader_test/main.go`

### 1. Basic Example (`basic/`)
- **Purpose**: Demonstrates fundamental shader compilation and rendering
- **Features**: 
  - Vertex and fragment shader compilation
  - Uniform variable management  
  - Time-based animation
  - Basic triangle rendering with color interpolation
- **OpenGL Requirements**: 4.1+
- **Run**: `go run cmd/examples/basic/main.go`

### 2. Geometry Shader Example (`geometry/`)
- **Purpose**: Showcases geometry shader capabilities for expanding primitives
- **Features**:
  - Point-to-quad expansion (billboard particles)
  - Wireframe generation from triangles
  - Interactive demo switching (1/2 keys)
  - Point size control (+/- keys)
- **OpenGL Requirements**: 3.2+ (geometry shaders)
- **Run**: `go run cmd/examples/geometry/main.go`

### 3. Compute Shader Example (`compute/`)
- **Purpose**: Demonstrates GPU compute capabilities with automatic fallback
- **Features**:
  - GPU-based particle system (when compute shaders available)
  - CPU-based particle fallback (when compute shaders unavailable)
  - Mouse-controlled attractor
  - Physics simulation (gravity, collision detection)
  - Platform-aware feature detection
- **OpenGL Requirements**: 4.3+ for GPU compute, 4.1+ for CPU fallback
- **Run**: `go run cmd/examples/compute/main.go`

### 4. Pipeline Example (`pipeline/`)
- **Purpose**: Shows advanced rendering pipeline state management
- **Features**:
  - Multiple render passes
  - Blending modes
  - Depth testing configurations
  - Viewport management
- **OpenGL Requirements**: 4.1+
- **Run**: `go run cmd/examples/pipeline/main.go`

### 5. Platform Detection (`platform/`)
- **Purpose**: Analyzes system capabilities and provides development guidance
- **Features**:
  - Comprehensive OpenGL feature detection
  - Platform-specific recommendations
  - Library limitation awareness
  - Generated initialization code snippets
- **OpenGL Requirements**: Any
- **Run**: `go run cmd/examples/platform/main.go`

## Platform Compatibility

### macOS
- Maximum OpenGL version: 4.1 (due to Apple deprecation)
- Compute shaders: Not available
- Geometry shaders: ✅ Available
- Recommendation: Consider Metal for production applications

### Linux/Windows
- OpenGL version: Up to 4.6 (driver dependent)
- Compute shaders: Available if system supports OpenGL 4.3+
- Note: This Go library is limited to OpenGL 4.1 core features
- Geometry shaders: ✅ Available

## Library Limitations

**Important**: This project uses `github.com/go-gl/gl/v4.1-core/gl`, which limits available features to OpenGL 4.1 core profile, regardless of system capabilities.

**Impact**:
- ❌ Compute shaders (require OpenGL 4.3+)
- ❌ Shader storage buffers (require OpenGL 4.3+)
- ✅ Geometry shaders (available in OpenGL 3.2+)
- ✅ Tessellation shaders (available in OpenGL 4.0+)
- ✅ All other modern rendering features up to 4.1

## Getting Started

1. **Install dependencies**:
   ```bash
   # Ubuntu/Debian
   sudo apt install libgl1-mesa-dev libglu1-mesa-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev

   # macOS
   # Xcode command line tools include OpenGL

   # Windows  
   # OpenGL typically included with graphics drivers
   ```

2. **Check your system capabilities**:
   ```bash
   go run cmd/examples/platform/main.go
   ```

3. **Run the basic example**:
   ```bash
   go run cmd/examples/basic/main.go
   ```

4. **Explore advanced features**:
   ```bash
   go run cmd/examples/geometry/main.go
   go run cmd/examples/compute/main.go
   ```

## Controls

### Basic Example
- **ESC**: Exit

### Geometry Example  
- **1**: Point-to-quad demo
- **2**: Wireframe demo
- **+/-**: Adjust point size
- **ESC**: Exit

### Compute Example
- **Mouse**: Move particle attractor
- **ESC**: Exit

## Troubleshooting

### "Failed to initialize GLFW" or OpenGL context creation errors
- Install required system libraries (see Getting Started)
- Update graphics drivers
- Run platform detection to check capabilities

### Shader compilation errors
- Check OpenGL version compatibility
- Ensure your system supports the required OpenGL version
- Review shader source for version-specific syntax

### Performance issues
- Use platform detection to check GPU vendor and capabilities
- Consider conservative settings for Intel integrated graphics
- Enable VSync for smooth animation

## Development Notes

- All examples include proper resource cleanup
- Platform detection helps choose appropriate rendering techniques
- Fallback mechanisms ensure examples work across different systems
- Code demonstrates professional OpenGL practices and error handling