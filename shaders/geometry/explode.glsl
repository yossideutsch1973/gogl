#version 410 core

layout(triangles) in;
layout(triangle_strip, max_vertices = 3) out;

uniform mat4 uProjection;
uniform mat4 uView;
uniform float uExplodeDistance;

in vec3 vColor[];
in vec3 vNormal[];

out vec3 fColor;

vec3 getTriangleNormal() {
    vec3 a = vec3(gl_in[1].gl_Position) - vec3(gl_in[0].gl_Position);
    vec3 b = vec3(gl_in[2].gl_Position) - vec3(gl_in[0].gl_Position);
    return normalize(cross(a, b));
}

void main() {
    vec3 normal = getTriangleNormal();
    
    for (int i = 0; i < 3; i++) {
        fColor = vColor[i];
        vec4 pos = gl_in[i].gl_Position + vec4(normal * uExplodeDistance, 0.0);
        gl_Position = uProjection * uView * pos;
        EmitVertex();
    }
    
    EndPrimitive();
}
