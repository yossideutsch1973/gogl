#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uScreenTexture;

void main() {
    vec3 color = texture(uScreenTexture, vTexCoord).rgb;
    
    // Invert colors
    fragColor = vec4(1.0 - color, 1.0);
}
