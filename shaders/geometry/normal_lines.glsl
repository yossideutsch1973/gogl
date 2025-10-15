#version 410 core

layout(triangles) in;
layout(line_strip, max_vertices = 6) out;

uniform mat4 uProjection;
uniform mat4 uView;
uniform float uNormalLength;

in vec3 vNormal[];
in vec3 vFragPos[];

out vec3 fColor;

void main() {
    for (int i = 0; i < 3; i++) {
        vec3 normal = normalize(vNormal[i]);
        vec3 pos = vFragPos[i];
        
        // Start of normal line
        fColor = vec3(1.0, 1.0, 0.0);
        gl_Position = uProjection * uView * vec4(pos, 1.0);
        EmitVertex();
        
        // End of normal line
        fColor = vec3(1.0, 0.0, 0.0);
        gl_Position = uProjection * uView * vec4(pos + normal * uNormalLength, 1.0);
        EmitVertex();
        
        EndPrimitive();
    }
}
