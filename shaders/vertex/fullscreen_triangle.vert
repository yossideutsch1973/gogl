#version 410 core

// Emits a single oversized triangle that covers the entire screen
// without needing a vertex buffer. Call with glDrawArrays(GL_TRIANGLES, 0, 3).
//
// Reference: https://www.saschawillems.de/blog/2016/08/13/vulkan-tutorial-on-rendering-a-fullscreen-quad-without-buffers/

out vec2 vUV;

void main() {
    vec2 p = vec2((gl_VertexID << 1) & 2, gl_VertexID & 2);
    vUV = p;
    gl_Position = vec4(p * 2.0 - 1.0, 0.0, 1.0);
}
