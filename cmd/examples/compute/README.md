# Compute Shader Example

This example demonstrates GPU-accelerated particle simulation using OpenGL compute shaders.

## Platform Compatibility

**⚠️ Important:** Compute shaders require OpenGL 4.3 or higher.

### macOS Limitations
- OpenGL 4.1 is the maximum supported version on macOS
- Compute shaders are **not available** on macOS
- Use the platform detection example to check capabilities: `go run cmd/examples/platform/main.go`

### Alternative Approaches
If compute shaders are not available on your platform:
1. Use transform feedback with vertex shaders
2. Implement CPU-based particle simulation
3. Use geometry shaders for simpler effects
4. Consider platform-specific solutions (Metal on macOS, DirectCompute on Windows)

## Running the Example

The example will automatically detect if compute shaders are supported and show an appropriate message. On unsupported platforms, it will exit gracefully with information about alternatives.

## Supported Platforms
- Windows: OpenGL 4.3+ (most modern GPUs)
- Linux: OpenGL 4.3+ (most modern GPUs)
- macOS: Not supported (OpenGL 4.1 limitation)

## Future Enhancements
- Fallback to transform feedback when compute shaders unavailable
- CPU-based particle simulation alternative
- Metal compute shader version for macOS