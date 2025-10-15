# GoGL Shader Library

This directory contains a comprehensive collection of GLSL shaders for common rendering tasks. All shaders are written with OpenGL 4.1 compatibility for maximum cross-platform support.

## Directory Structure

```
shaders/
├── vertex/          # Vertex shader programs
├── fragment/        # Fragment shader programs
├── geometry/        # Geometry shader programs
└── compute/         # Compute shader programs (requires OpenGL 4.3+)
```

## Vertex Shaders

### basic.vert
Basic vertex shader with position, color, and MVP matrix.
- **Inputs**: `aPosition` (vec3), `aColor` (vec3)
- **Outputs**: `vColor` (vec3)
- **Uniforms**: `uModelViewProjection` (mat4)
- **Use Case**: Simple colored geometry

### textured.vert
Vertex shader for textured meshes with lighting support.
- **Inputs**: `aPosition` (vec3), `aTexCoord` (vec2), `aNormal` (vec3)
- **Outputs**: `vTexCoord` (vec2), `vNormal` (vec3), `vFragPos` (vec3)
- **Uniforms**: `uModel` (mat4), `uView` (mat4), `uProjection` (mat4)
- **Use Case**: Textured objects with lighting

### phong.vert
Vertex shader for Phong lighting model.
- **Inputs**: `aPosition` (vec3), `aNormal` (vec3), `aColor` (vec3)
- **Outputs**: `vFragPos` (vec3), `vNormal` (vec3), `vColor` (vec3)
- **Uniforms**: `uModel` (mat4), `uView` (mat4), `uProjection` (mat4)
- **Use Case**: Per-vertex lighting calculations

### flat_color.vert
Simple flat color shader with uniform color.
- **Inputs**: `aPosition` (vec3)
- **Outputs**: `vColor` (vec3)
- **Uniforms**: `uModel` (mat4), `uView` (mat4), `uProjection` (mat4), `uColor` (vec3)
- **Use Case**: Solid colored objects

### skybox.vert
Specialized vertex shader for skybox rendering.
- **Inputs**: `aPosition` (vec3)
- **Outputs**: `vTexCoord` (vec3)
- **Uniforms**: `uView` (mat4), `uProjection` (mat4)
- **Use Case**: Skybox/environment mapping

### screen_quad.vert
Screen-space quad for post-processing.
- **Inputs**: `aPosition` (vec2), `aTexCoord` (vec2)
- **Outputs**: `vTexCoord` (vec2)
- **Use Case**: Full-screen post-processing effects

### standard.vert
Standard PBR-ready vertex shader.
- **Inputs**: `aPosition` (vec3), `aNormal` (vec3), `aTexCoord` (vec2)
- **Outputs**: `vFragPos` (vec3), `vNormal` (vec3), `vTexCoord` (vec2)
- **Uniforms**: `uModel` (mat4), `uView` (mat4), `uProjection` (mat4), `uNormalMatrix` (mat3)
- **Use Case**: General-purpose rendering

## Fragment Shaders

### basic.frag
Basic fragment shader with time-based color animation.
- **Inputs**: `vColor` (vec3)
- **Uniforms**: `uTime` (float)
- **Use Case**: Simple animated rendering

### textured.frag
Fragment shader with texture sampling and Phong lighting.
- **Inputs**: `vTexCoord` (vec2), `vNormal` (vec3), `vFragPos` (vec3)
- **Uniforms**: `uTexture` (sampler2D), `uLightPos` (vec3), `uViewPos` (vec3), `uLightColor` (vec3)
- **Use Case**: Textured objects with lighting

### phong.frag
Phong lighting model fragment shader.
- **Inputs**: `vFragPos` (vec3), `vNormal` (vec3), `vColor` (vec3)
- **Uniforms**: `uLightPos` (vec3), `uViewPos` (vec3), `uLightColor` (vec3), `uAmbientStrength` (float), `uSpecularStrength` (float), `uShininess` (float)
- **Use Case**: Per-fragment Phong lighting

### flat_color.frag
Simple flat color output.
- **Inputs**: `vColor` (vec3)
- **Use Case**: Solid color rendering

### skybox.frag
Skybox fragment shader with cubemap sampling.
- **Inputs**: `vTexCoord` (vec3)
- **Uniforms**: `uSkybox` (samplerCube)
- **Use Case**: Skybox/environment rendering

### simple_texture.frag
Basic texture sampling without lighting.
- **Inputs**: `vTexCoord` (vec2)
- **Uniforms**: `uTexture` (sampler2D)
- **Use Case**: Simple texture display

## Post-Processing Fragment Shaders

### blur.frag
Box blur post-processing effect.
- **Uniforms**: `uScreenTexture` (sampler2D), `uBlurRadius` (float)
- **Use Case**: Blur effects, bloom, depth of field

### edge_detection.frag
Edge detection using convolution kernel.
- **Uniforms**: `uScreenTexture` (sampler2D)
- **Use Case**: Edge highlighting, artistic effects

### grayscale.frag
Convert color to grayscale using luminance.
- **Uniforms**: `uScreenTexture` (sampler2D)
- **Use Case**: Black and white effect

### invert.frag
Invert screen colors.
- **Uniforms**: `uScreenTexture` (sampler2D)
- **Use Case**: Negative image effect

### brightness_contrast.frag
Adjust brightness and contrast.
- **Uniforms**: `uScreenTexture` (sampler2D), `uBrightness` (float), `uContrast` (float)
- **Use Case**: Image adjustment, tone mapping

### gamma_correction.frag
Apply gamma correction.
- **Uniforms**: `uScreenTexture` (sampler2D), `uGamma` (float)
- **Use Case**: Color space correction

## Geometry Shaders

### point_to_quad.glsl
Expand points to billboard quads.
- **Input Primitive**: points
- **Output Primitive**: triangle_strip (max 4 vertices)
- **Uniforms**: `uProjection` (mat4), `uView` (mat4), `uPointSize` (float)
- **Use Case**: Particle systems, sprites

### wireframe.glsl
Generate wireframe from triangles.
- **Input Primitive**: triangles
- **Output Primitive**: line_strip (max 4 vertices)
- **Use Case**: Wireframe rendering, debug visualization

### normal_visualization.glsl
Visualize surface normals.
- **Input Primitive**: triangles
- **Output Primitive**: line_strip (max 6 vertices)
- **Uniforms**: `uProjection` (mat4), `uView` (mat4), `uNormalLength` (float)
- **Use Case**: Debug normal directions

### normal_lines.glsl
Draw lines along vertex normals.
- **Input Primitive**: triangles
- **Output Primitive**: line_strip (max 6 vertices)
- **Uniforms**: `uProjection` (mat4), `uView` (mat4), `uNormalLength` (float)
- **Use Case**: Normal visualization, debug rendering

### explode.glsl
Explode triangles along their face normal.
- **Input Primitive**: triangles
- **Output Primitive**: triangle_strip (max 3 vertices)
- **Uniforms**: `uProjection` (mat4), `uView` (mat4), `uExplodeDistance` (float)
- **Use Case**: Explosion effects, model dissection

## Compute Shaders

**Note**: Compute shaders require OpenGL 4.3+. They are not available on macOS (limited to OpenGL 4.1).

### particle_simulation.glsl
GPU-based particle physics simulation.
- **Work Group Size**: 16x16x1
- **Uniforms**: `uDeltaTime` (float), `uGravity` (float), `uAttractor` (vec2), `uAttractorStrength` (float), `uViewportSize` (vec2)
- **Use Case**: Large-scale particle systems

### image_processing.glsl
Image processing with multiple filter types.
- **Work Group Size**: 16x16x1
- **Uniforms**: `uBlurRadius` (float), `uBrightness` (float), `uContrast` (float), `uFilterType` (int)
- **Use Case**: Real-time image filters

## Usage Examples

### Basic Rendering
```go
vertexShader, _ := shader.CompileShaderFromFile("shaders/vertex/basic.vert", shader.VertexShader)
fragmentShader, _ := shader.CompileShaderFromFile("shaders/fragment/basic.frag", shader.FragmentShader)
program, _ := shader.CreateProgram(vertexShader, fragmentShader)
```

### Textured Object with Lighting
```go
vertexShader, _ := shader.CompileShaderFromFile("shaders/vertex/textured.vert", shader.VertexShader)
fragmentShader, _ := shader.CompileShaderFromFile("shaders/fragment/textured.frag", shader.FragmentShader)
program, _ := shader.CreateProgram(vertexShader, fragmentShader)
```

### Geometry Shader Pipeline
```go
vertexShader, _ := shader.CompileShaderFromFile("shaders/vertex/standard.vert", shader.VertexShader)
geometryShader, _ := shader.CompileShaderFromFile("shaders/geometry/explode.glsl", shader.GeometryShader)
fragmentShader, _ := shader.CompileShaderFromFile("shaders/fragment/phong.frag", shader.FragmentShader)
program, _ := shader.CreateProgram(vertexShader, geometryShader, fragmentShader)
```

### Post-Processing Effect
```go
vertexShader, _ := shader.CompileShaderFromFile("shaders/vertex/screen_quad.vert", shader.VertexShader)
fragmentShader, _ := shader.CompileShaderFromFile("shaders/fragment/blur.frag", shader.FragmentShader)
program, _ := shader.CreateProgram(vertexShader, fragmentShader)
```

## Platform Compatibility

- **Vertex/Fragment Shaders**: OpenGL 4.1+ (all platforms including macOS)
- **Geometry Shaders**: OpenGL 3.2+ (all platforms including macOS)
- **Compute Shaders**: OpenGL 4.3+ (Linux/Windows only, not available on macOS)

## Contributing

When adding new shaders:
1. Follow the OpenGL 4.1 baseline for maximum compatibility
2. Document inputs, outputs, and uniforms clearly
3. Include use case examples
4. Test on multiple platforms when possible
5. Use consistent naming conventions (u prefix for uniforms, v for varyings, a for attributes)
