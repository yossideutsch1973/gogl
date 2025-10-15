#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uScreenTexture;

void main() {
    vec3 color = texture(uScreenTexture, vTexCoord).rgb;
    
    // Convert to grayscale using luminance weights
    float gray = dot(color, vec3(0.299, 0.587, 0.114));
    
    fragColor = vec4(vec3(gray), 1.0);
}
