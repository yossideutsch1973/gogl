#version 410 core

layout(location = 0) in vec3 aPosition;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProjection;
uniform vec3 uColor;

out vec3 vColor;

void main() {
    vColor = uColor;
    gl_Position = uProjection * uView * uModel * vec4(aPosition, 1.0);
}
