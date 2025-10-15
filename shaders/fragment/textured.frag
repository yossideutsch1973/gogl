#version 410 core

in vec2 vTexCoord;
in vec3 vNormal;
in vec3 vFragPos;

out vec4 fragColor;

uniform sampler2D uTexture;
uniform vec3 uLightPos;
uniform vec3 uViewPos;
uniform vec3 uLightColor;

void main() {
    // Ambient
    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * uLightColor;
    
    // Diffuse
    vec3 norm = normalize(vNormal);
    vec3 lightDir = normalize(uLightPos - vFragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * uLightColor;
    
    // Specular
    float specularStrength = 0.5;
    vec3 viewDir = normalize(uViewPos - vFragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32.0);
    vec3 specular = specularStrength * spec * uLightColor;
    
    // Combine with texture
    vec4 texColor = texture(uTexture, vTexCoord);
    vec3 result = (ambient + diffuse + specular) * texColor.rgb;
    
    fragColor = vec4(result, texColor.a);
}
