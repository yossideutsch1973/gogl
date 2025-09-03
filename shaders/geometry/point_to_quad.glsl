#version 410 core

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
}