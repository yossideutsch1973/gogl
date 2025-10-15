# GitHub Copilot Instructions for GoGL

## Project Overview
GoGL is a professional OpenGL shader library for Go, focusing on cross-platform graphics development with emphasis on clean APIs, performance optimization, and comprehensive error handling.

## Technology Stack
- **Language**: Go (following Go conventions strictly)
- **Graphics**: OpenGL 4.1+ (4.1 for macOS compatibility, 4.6 for Linux/Windows)
- **Dependencies**: go-gl/gl, go-gl/glfw, go-gl/mathgl
- **Target Platforms**: macOS (OpenGL 4.1), Linux/Windows (OpenGL 4.6)

## Architecture & Code Organization

### Package Structure
- `pkg/shader/` - Core shader compilation, linking, and program management
- `pkg/pipeline/` - Rendering pipeline and state management
- `pkg/resource/` - Buffer, texture, and VAO lifecycle management
- `internal/platform/` - Platform-specific code and capability detection
- `cmd/examples/` - Demonstration applications
- `shaders/` - GLSL shader source files
- `tests/unit/` - Unit tests with OpenGL context setup

### Key Design Principles
1. **Performance First**: Minimize allocations in render loops, use object pooling
2. **Error Handling**: Comprehensive validation with context-rich errors
3. **Cross-Platform**: OpenGL 4.1 baseline for maximum compatibility
4. **Clean APIs**: Follow Go idioms with intuitive interfaces
5. **Resource Safety**: Proper cleanup patterns and lifecycle management

## Coding Standards

### Go Conventions
- Follow standard Go conventions (gofmt, golint)
- Use interfaces for abstractions
- Return errors, don't panic (except for unrecoverable situations)
- Include godoc comments for all exported functions and types
- Write table-driven tests

### Error Handling
```go
// Always return errors with context
if err != nil {
    return fmt.Errorf("failed to compile shader: %w", err)
}

// Validate inputs early
if source == "" {
    return nil, errors.New("shader source cannot be empty")
}
```

### Performance Patterns
```go
// Avoid allocations in render loops
// ❌ Bad
for frame := range renderLoop {
    data := make([]float32, size)  // Allocates every frame
}

// ✅ Good
data := make([]float32, size)
for frame := range renderLoop {
    // Reuse buffer
}
```

### OpenGL Resource Management
```go
// Always implement cleanup
type Resource struct {
    id uint32
}

func (r *Resource) Delete() {
    if r.id != 0 {
        gl.DeleteBuffers(1, &r.id)
        r.id = 0
    }
}
```

## Platform-Specific Considerations

### macOS
- **OpenGL Version**: Limited to 4.1 (deprecated since 2018)
- **Context Creation**: Must set `glfw.OpenGLForwardCompatible, glfw.True`
- **Compute Shaders**: Not available (requires OpenGL 4.3+)
- **Metal Backend**: OpenGL runs through Metal compatibility layer
- **Testing**: Verify on both Intel and Apple Silicon

### Linux/Windows  
- **OpenGL Version**: Target 4.6 core profile
- **Full Features**: Compute shaders, geometry shaders, tessellation available
- **GPU Vendors**: Test across NVIDIA, AMD, Intel

## Development Workflow

### Building & Testing
```bash
# Run basic example (verify foundation works)
go run cmd/examples/basic/main.go

# Run all tests (requires X11 for CI, uses xvfb)
go test ./...

# Run specific package tests
go test ./tests/unit/
go test ./pkg/shader/

# Clean dependencies
go mod tidy
```

### Adding New Features
1. Start with interface design in appropriate package
2. Implement with comprehensive error handling
3. Add unit tests with OpenGL context setup
4. Update examples if adding new capabilities
5. Document platform compatibility if relevant

### Shader Development
- Place GLSL sources in `shaders/` directory organized by type
- Use `#version 410 core` for macOS compatibility
- Use `#version 460 core` for advanced features (document limitations)
- Validate shader compilation with detailed error messages
- Check OpenGL errors after each significant operation

## Testing Requirements

### Unit Tests
- Use `tests/unit/` directory structure
- Initialize OpenGL context in test setup (see existing tests)
- Clean up resources in test teardown
- Test both success and error paths
- Include edge cases (empty inputs, invalid parameters)

### Example Test Pattern
```go
func TestShaderCompilation(t *testing.T) {
    // Initialize GLFW and OpenGL context
    if err := glfw.Init(); err != nil {
        t.Fatal(err)
    }
    defer glfw.Terminate()
    
    // Create window for OpenGL context
    window, err := createTestWindow()
    if err != nil {
        t.Fatal(err)
    }
    defer window.Destroy()
    
    // Test implementation
    // ...
}
```

## Common Patterns & Gotchas

### Known Constraints
- **macOS OpenGL**: Deprecated, Metal backend provides compatibility
- **Compute Shaders**: Require OpenGL 4.3+, unavailable on macOS
- **Context Creation**: Forward-compatible flag required on macOS
- **Validation**: Program validation on macOS generates expected warnings (not errors)

### Performance Optimization
- Cache compiled shader programs
- Batch state changes
- Minimize texture/buffer binding changes
- Use vertex array objects (VAOs) for geometry
- Profile on actual hardware, not software drivers

### Resource Management
- Always call Delete() on OpenGL resources
- Check for OpenGL errors after operations
- Validate resource IDs before use
- Implement proper cleanup in defer statements

## Documentation Standards
- Add package-level documentation with usage examples
- Document all exported functions with godoc comments
- Include parameter descriptions and validation rules
- Document error conditions
- Add cross-platform compatibility notes where relevant
- Update README.md for significant feature additions

## Current Development Focus
The project has a solid foundation with core shader system, basic rendering, and test infrastructure. Current expansion areas:

1. **Geometry Shaders** - Point/line expansion, wireframe rendering
2. **Compute Shaders** - GPU computation (OpenGL 4.3+, Linux/Windows only)
3. **Advanced Examples** - Complex shader demonstrations
4. **Performance Tools** - GPU profiling and optimization utilities

When working on new features, maintain the established patterns for error handling, resource management, and cross-platform compatibility.
