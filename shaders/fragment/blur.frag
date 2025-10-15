#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uScreenTexture;
uniform float uBlurRadius;

void main() {
    vec2 texelSize = 1.0 / vec2(textureSize(uScreenTexture, 0));
    vec4 result = vec4(0.0);
    float totalWeight = 0.0;
    
    int radius = int(uBlurRadius);
    for (int x = -radius; x <= radius; x++) {
        for (int y = -radius; y <= radius; y++) {
            vec2 offset = vec2(float(x), float(y)) * texelSize;
            float weight = 1.0;
            result += texture(uScreenTexture, vTexCoord + offset) * weight;
            totalWeight += weight;
        }
    }
    
    fragColor = result / totalWeight;
}
