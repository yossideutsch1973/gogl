#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uScreenTexture;
uniform float uBrightness;
uniform float uContrast;

void main() {
    vec3 color = texture(uScreenTexture, vTexCoord).rgb;
    
    // Apply contrast
    color = (color - 0.5) * uContrast + 0.5;
    
    // Apply brightness
    color += uBrightness;
    
    // Clamp to valid range
    color = clamp(color, 0.0, 1.0);
    
    fragColor = vec4(color, 1.0);
}
