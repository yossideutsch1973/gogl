#version 410 core

// Raymarched showcase scene for GoGL.
//
// Renders a rotating, soft-shadowed scene of an analytical sphere and
// torus on an infinite checkered plane. Intended as a flagship demo
// for the gogl library — all geometry is procedural in the fragment
// shader, so the host code only needs a full-screen triangle.

in vec2 vUV;
out vec4 fragColor;

uniform float uTime;
uniform vec2  uResolution;

// --- SDF primitives ---------------------------------------------------------

float sdSphere(vec3 p, float r) {
    return length(p) - r;
}

float sdTorus(vec3 p, vec2 t) {
    vec2 q = vec2(length(p.xz) - t.x, p.y);
    return length(q) - t.y;
}

float sdPlane(vec3 p) {
    return p.y;
}

float opSmoothUnion(float a, float b, float k) {
    float h = clamp(0.5 + 0.5 * (b - a) / k, 0.0, 1.0);
    return mix(b, a, h) - k * h * (1.0 - h);
}

mat3 rotY(float a) {
    float c = cos(a), s = sin(a);
    return mat3(c, 0.0, s, 0.0, 1.0, 0.0, -s, 0.0, c);
}

// --- Scene ------------------------------------------------------------------

struct Hit {
    float d;   // distance
    int   id;  // material id: 0=plane, 1=sphere, 2=torus
};

Hit map(vec3 p) {
    Hit h;

    vec3 sp = p - vec3(0.0, 1.0 + 0.15 * sin(uTime * 1.7), 0.0);
    float ds = sdSphere(sp, 0.8);

    vec3 tp = rotY(uTime * 0.6) * (p - vec3(0.0, 1.0, 0.0));
    float dt = sdTorus(tp, vec2(1.4, 0.12));

    float plane = sdPlane(p);

    float blob = opSmoothUnion(ds, dt, 0.18);

    if (plane < blob) {
        h.d = plane;
        h.id = 0;
    } else {
        h.d = blob;
        h.id = (ds < dt) ? 1 : 2;
    }
    return h;
}

vec3 calcNormal(vec3 p) {
    const float e = 0.0008;
    vec2 k = vec2(1.0, -1.0);
    return normalize(
        k.xyy * map(p + k.xyy * e).d +
        k.yyx * map(p + k.yyx * e).d +
        k.yxy * map(p + k.yxy * e).d +
        k.xxx * map(p + k.xxx * e).d
    );
}

// Soft shadow march. Iq's classic.
float softShadow(vec3 ro, vec3 rd, float mint, float maxt, float k) {
    float res = 1.0;
    float t = mint;
    for (int i = 0; i < 48; i++) {
        float h = map(ro + rd * t).d;
        if (h < 0.0005) return 0.0;
        res = min(res, k * h / t);
        t += clamp(h, 0.02, 0.4);
        if (t > maxt) break;
    }
    return clamp(res, 0.0, 1.0);
}

Hit raymarch(vec3 ro, vec3 rd) {
    Hit h;
    h.d = 0.0;
    h.id = -1;
    float t = 0.0;
    for (int i = 0; i < 128; i++) {
        vec3 p = ro + rd * t;
        Hit s = map(p);
        if (s.d < 0.0008 * t) {
            h.d = t;
            h.id = s.id;
            return h;
        }
        t += s.d;
        if (t > 40.0) break;
    }
    h.d = -1.0;
    return h;
}

// Checkerboard albedo for the plane
vec3 checker(vec2 p) {
    vec2 q = floor(p);
    float m = mod(q.x + q.y, 2.0);
    return mix(vec3(0.18), vec3(0.32), m);
}

vec3 shade(vec3 ro, vec3 rd) {
    // Animated gradient background
    vec3 sky = mix(vec3(0.04, 0.06, 0.10), vec3(0.55, 0.65, 0.95), pow(max(rd.y, 0.0), 0.6));
    sky += vec3(0.7, 0.5, 0.3) * pow(max(-rd.y * 0.5 + 0.5, 0.0), 4.0) * 0.25;

    Hit h = raymarch(ro, rd);
    if (h.d < 0.0) return sky;

    vec3 p = ro + rd * h.d;
    vec3 n = calcNormal(p);

    vec3 lightDir = normalize(vec3(0.6, 0.9, 0.4));
    float diff = clamp(dot(n, lightDir), 0.0, 1.0);
    float sh   = softShadow(p + n * 0.002, lightDir, 0.02, 12.0, 16.0);

    vec3 albedo;
    if (h.id == 0) {
        albedo = checker(p.xz * 0.5);
    } else if (h.id == 1) {
        albedo = vec3(0.95, 0.35, 0.25);
    } else {
        albedo = vec3(0.20, 0.55, 0.85);
    }

    vec3 col = albedo * (0.18 + 0.82 * diff * sh);

    // Cheap specular for non-plane hits
    if (h.id != 0) {
        vec3 r = reflect(-lightDir, n);
        float spec = pow(max(dot(r, -rd), 0.0), 32.0);
        col += vec3(1.0) * spec * sh * 0.5;
    }

    // Distance fog blending to sky
    float fog = 1.0 - exp(-0.025 * h.d * h.d);
    col = mix(col, sky, fog);
    return col;
}

void main() {
    vec2 uv = (vUV * 2.0 - 1.0);
    uv.x *= uResolution.x / uResolution.y;

    // Slow orbit camera
    float a = uTime * 0.25;
    vec3 ro = vec3(4.0 * cos(a), 2.2, 4.0 * sin(a));
    vec3 ta = vec3(0.0, 1.0, 0.0);

    vec3 ww = normalize(ta - ro);
    vec3 uu = normalize(cross(ww, vec3(0.0, 1.0, 0.0)));
    vec3 vv = cross(uu, ww);
    vec3 rd = normalize(uv.x * uu + uv.y * vv + 1.6 * ww);

    vec3 col = shade(ro, rd);

    // Gamma + subtle vignette
    col = pow(col, vec3(1.0 / 2.2));
    float vig = smoothstep(1.4, 0.4, length(uv));
    col *= mix(0.85, 1.0, vig);

    fragColor = vec4(col, 1.0);
}
