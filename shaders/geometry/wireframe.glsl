#version 410 core

layout(triangles) in;
layout(line_strip, max_vertices = 4) out;

in vec3 vColor[];
out vec3 fColor;

void main() {
    // Emit the three edges of the triangle
    for(int i = 0; i < 3; i++) {
        gl_Position = gl_in[i].gl_Position;
        fColor = vColor[i];
        EmitVertex();
    }
    
    // Complete the triangle by connecting back to the first vertex
    gl_Position = gl_in[0].gl_Position;
    fColor = vColor[0];
    EmitVertex();
    
    EndPrimitive();
}