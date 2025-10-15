#version 410 core

in vec3 vFragPos;
in vec3 vNormal;
in vec3 vColor;

out vec4 fragColor;

uniform vec3 uLightPos;
uniform vec3 uViewPos;
uniform vec3 uLightColor;
uniform float uAmbientStrength;
uniform float uSpecularStrength;
uniform float uShininess;

void main() {
    // Ambient
    vec3 ambient = uAmbientStrength * uLightColor;
    
    // Diffuse
    vec3 norm = normalize(vNormal);
    vec3 lightDir = normalize(uLightPos - vFragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * uLightColor;
    
    // Specular
    vec3 viewDir = normalize(uViewPos - vFragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), uShininess);
    vec3 specular = uSpecularStrength * spec * uLightColor;
    
    // Combine with vertex color
    vec3 result = (ambient + diffuse + specular) * vColor;
    
    fragColor = vec4(result, 1.0);
}
