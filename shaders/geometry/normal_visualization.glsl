#version 410 core

layout(triangles) in;
layout(line_strip, max_vertices = 6) out;

uniform mat4 uProjection;
uniform mat4 uView;
uniform mat4 uModel;
uniform float uNormalLength;

in vec3 vPosition[];
in vec3 vNormal[];

out vec3 fColor;

void main() {
    // Calculate the normal for the triangle
    vec3 edge1 = vPosition[1] - vPosition[0];
    vec3 edge2 = vPosition[2] - vPosition[0];
    vec3 normal = normalize(cross(edge1, edge2));
    
    // Calculate center of triangle
    vec3 center = (vPosition[0] + vPosition[1] + vPosition[2]) / 3.0;
    
    // Transform to world space
    vec4 worldCenter = uModel * vec4(center, 1.0);
    vec4 worldNormal = uModel * vec4(normal, 0.0);
    
    // Draw normal line from center
    fColor = vec3(1.0, 1.0, 0.0); // Yellow for normals
    
    // Start point (triangle center)
    gl_Position = uProjection * uView * worldCenter;
    EmitVertex();
    
    // End point (center + normal * length)
    gl_Position = uProjection * uView * (worldCenter + worldNormal * uNormalLength);
    EmitVertex();
    
    EndPrimitive();
}