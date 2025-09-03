#version 410 core

layout(local_size_x = 16, local_size_y = 16, local_size_z = 1) in;

// Input and output textures
layout(rgba8, binding = 0) restrict readonly uniform image2D inputImage;
layout(rgba8, binding = 1) restrict writeonly uniform image2D outputImage;

// Uniforms for filter parameters
uniform float uBlurRadius;
uniform float uBrightness;
uniform float uContrast;
uniform int uFilterType; // 0=blur, 1=edge detection, 2=brightness/contrast

// Convolution kernels
const float[9] edgeKernel = float[](
    -1.0, -1.0, -1.0,
    -1.0,  8.0, -1.0,
    -1.0, -1.0, -1.0
);

const float[9] blurKernel = float[](
    1.0/16.0, 2.0/16.0, 1.0/16.0,
    2.0/16.0, 4.0/16.0, 2.0/16.0,
    1.0/16.0, 2.0/16.0, 1.0/16.0
);

vec4 sampleImage(ivec2 coord) {
    ivec2 size = imageSize(inputImage);
    coord = clamp(coord, ivec2(0), size - 1);
    return imageLoad(inputImage, coord);
}

vec4 applyBlur(ivec2 coord) {
    vec4 color = vec4(0.0);
    int radius = int(uBlurRadius);
    
    for (int y = -radius; y <= radius; y++) {
        for (int x = -radius; x <= radius; x++) {
            ivec2 sampleCoord = coord + ivec2(x, y);
            color += sampleImage(sampleCoord);
        }
    }
    
    float samples = float((radius * 2 + 1) * (radius * 2 + 1));
    return color / samples;
}

vec4 applyEdgeDetection(ivec2 coord) {
    vec4 color = vec4(0.0);
    
    for (int i = 0; i < 9; i++) {
        int x = i % 3 - 1;
        int y = i / 3 - 1;
        ivec2 sampleCoord = coord + ivec2(x, y);
        color += sampleImage(sampleCoord) * edgeKernel[i];
    }
    
    return vec4(abs(color.rgb), 1.0);
}

vec4 applyBrightnessContrast(ivec2 coord) {
    vec4 color = sampleImage(coord);
    
    // Apply contrast first
    color.rgb = (color.rgb - 0.5) * uContrast + 0.5;
    
    // Then apply brightness
    color.rgb += uBrightness;
    
    // Clamp to valid range
    color.rgb = clamp(color.rgb, 0.0, 1.0);
    
    return color;
}

void main() {
    ivec2 coord = ivec2(gl_GlobalInvocationID.xy);
    ivec2 size = imageSize(outputImage);
    
    if (coord.x >= size.x || coord.y >= size.y) {
        return;
    }
    
    vec4 result;
    
    switch (uFilterType) {
        case 0:
            result = applyBlur(coord);
            break;
        case 1:
            result = applyEdgeDetection(coord);
            break;
        case 2:
            result = applyBrightnessContrast(coord);
            break;
        default:
            result = sampleImage(coord);
            break;
    }
    
    imageStore(outputImage, coord, result);
}