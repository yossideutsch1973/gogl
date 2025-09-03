#version 410 core

layout(local_size_x = 16, local_size_y = 16, local_size_z = 1) in;

// Particle data structure
struct Particle {
    vec2 position;
    vec2 velocity;
    vec4 color;
    float life;
    float size;
};

// Shader storage buffer objects
layout(std430, binding = 0) restrict buffer ParticleBuffer {
    Particle particles[];
};

// Uniforms
uniform float uDeltaTime;
uniform float uGravity;
uniform vec2 uAttractor;
uniform float uAttractorStrength;
uniform vec2 uViewportSize;

// Random number generation
uint hash(uint x) {
    x += (x << 10u);
    x ^= (x >> 6u);
    x += (x << 3u);
    x ^= (x >> 11u);
    x += (x << 15u);
    return x;
}

float random(uint seed) {
    return float(hash(seed)) / 4294967296.0;
}

void main() {
    uint index = gl_GlobalInvocationID.x + gl_GlobalInvocationID.y * gl_NumWorkGroups.x * gl_WorkGroupSize.x;
    
    if (index >= particles.length()) {
        return;
    }
    
    Particle particle = particles[index];
    
    // Skip dead particles
    if (particle.life <= 0.0) {
        return;
    }
    
    // Update life
    particle.life -= uDeltaTime;
    
    // Apply gravity
    particle.velocity.y -= uGravity * uDeltaTime;
    
    // Apply attractor force
    vec2 toAttractor = uAttractor - particle.position;
    float distance = length(toAttractor);
    if (distance > 0.01) {
        vec2 force = normalize(toAttractor) * uAttractorStrength / (distance * distance);
        particle.velocity += force * uDeltaTime;
    }
    
    // Update position
    particle.position += particle.velocity * uDeltaTime;
    
    // Bounce off walls
    if (particle.position.x < 0.0 || particle.position.x > uViewportSize.x) {
        particle.velocity.x *= -0.8;
        particle.position.x = clamp(particle.position.x, 0.0, uViewportSize.x);
    }
    if (particle.position.y < 0.0 || particle.position.y > uViewportSize.y) {
        particle.velocity.y *= -0.8;
        particle.position.y = clamp(particle.position.y, 0.0, uViewportSize.y);
    }
    
    // Fade out over time
    float lifeRatio = particle.life / 5.0; // Assume max life is 5 seconds
    particle.color.a = lifeRatio;
    
    // Respawn particle if dead
    if (particle.life <= 0.0) {
        uint seed = index + uint(uDeltaTime * 1000.0);
        particle.position = vec2(
            random(seed) * uViewportSize.x,
            uViewportSize.y + random(seed + 1u) * 100.0
        );
        particle.velocity = vec2(
            (random(seed + 2u) - 0.5) * 200.0,
            -random(seed + 3u) * 100.0 - 50.0
        );
        particle.color = vec4(
            random(seed + 4u),
            random(seed + 5u),
            random(seed + 6u),
            1.0
        );
        particle.life = 3.0 + random(seed + 7u) * 2.0;
        particle.size = 2.0 + random(seed + 8u) * 4.0;
    }
    
    particles[index] = particle;
}