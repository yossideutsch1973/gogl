#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uScreenTexture;

const mat3 edgeKernel = mat3(
    -1.0, -1.0, -1.0,
    -1.0,  8.0, -1.0,
    -1.0, -1.0, -1.0
);

void main() {
    vec2 texelSize = 1.0 / vec2(textureSize(uScreenTexture, 0));
    vec3 color = vec3(0.0);
    
    for (int x = -1; x <= 1; x++) {
        for (int y = -1; y <= 1; y++) {
            vec2 offset = vec2(float(x), float(y)) * texelSize;
            vec3 sample = texture(uScreenTexture, vTexCoord + offset).rgb;
            color += sample * edgeKernel[x+1][y+1];
        }
    }
    
    fragColor = vec4(abs(color), 1.0);
}
