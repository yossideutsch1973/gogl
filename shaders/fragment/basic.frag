#version 410 core

in vec3 vColor;
out vec4 fragColor;

uniform float uTime;

void main() {
    // Simple color variation based on time for visual verification
    vec3 color = vColor * (0.8 + 0.2 * sin(uTime));
    fragColor = vec4(color, 1.0);
}