#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uScreenTexture;
uniform float uGamma;

void main() {
    vec3 color = texture(uScreenTexture, vTexCoord).rgb;
    
    // Apply gamma correction
    color = pow(color, vec3(1.0 / uGamma));
    
    fragColor = vec4(color, 1.0);
}
