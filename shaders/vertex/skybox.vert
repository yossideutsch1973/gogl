#version 410 core

layout(location = 0) in vec3 aPosition;

uniform mat4 uView;
uniform mat4 uProjection;

out vec3 vTexCoord;

void main() {
    vTexCoord = aPosition;
    // Remove translation from view matrix for skybox
    mat4 viewNoTranslation = mat4(mat3(uView));
    vec4 pos = uProjection * viewNoTranslation * vec4(aPosition, 1.0);
    // Ensure skybox is at maximum depth
    gl_Position = pos.xyww;
}
