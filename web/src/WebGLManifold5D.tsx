import React, { useEffect, useRef } from "react";

export type Coords5D = { x1: number; x2: number; x3: number; x4: number; x5: number };
export type AgentPath = {
  agentId: string;
  color: string;
  current: Coords5D;
  history: Coords5D[];
};

type Props = {
  agents: AgentPath[];
  size?: number;
  selectedAgentId?: string | null;
  helixPointRef?: React.MutableRefObject<{x: number, y: number, z: number} | null>;
  manifoldPointRef?: React.MutableRefObject<{x: number, y: number, z: number} | null>;
  spectrumRef?: React.MutableRefObject<Float32Array>;
};

const vsRender = `#version 300 es
precision highp float;
uniform sampler2D uPosTex;
uniform sampler2D uColorTex;
uniform int uAgentCount;
uniform float uTheta;
uniform mat4 uProjection;
uniform mat4 uView;
out vec3 vColor;
void main() {
    float u = (float(gl_VertexID) + 0.5) / float(uAgentCount);
    vec3 pos3D = texture(uPosTex, vec2(u, 0.5)).xyz;
    vec3 color = texture(uColorTex, vec2(u, 0.5)).xyz;
    float c = cos(uTheta);
    float s = sin(uTheta);
    mat3 rotY = mat3(c, 0.0, -s, 0.0, 1.0, 0.0, s, 0.0, c);
    vec3 rotated = rotY * pos3D;
    gl_Position = uProjection * uView * vec4(rotated, 1.0);
    gl_PointSize = 8.0;
    vColor = color;
}
`;

const fsRender = `#version 300 es
precision highp float;
in vec3 vColor;
out vec4 outColor;
void main() {
    vec2 coord = gl_PointCoord * 2.0 - 1.0;
    float r = length(coord);
    float alpha = smoothstep(1.0, 0.0, r);
    vec3 base = vColor;
    vec3 glow = base * 1.5;
    outColor = vec4(mix(base, glow, 0.5), alpha);
}
`;

const vsQuad = `#version 300 es
precision highp float;
out vec2 vUv;
void main() {
    float x = float((gl_VertexID & 1) << 2);
    float y = float((gl_VertexID & 2) << 1);
    vUv = vec2(x * 0.5, y * 0.5);
    gl_Position = vec4(x - 1.0, y - 1.0, 0.0, 1.0);
}
`;

const fsTrailFade = `#version 300 es
precision highp float;
uniform sampler2D uPrevTrail;
uniform sampler2D uAgentMask;
uniform float uDecay;
uniform float uIntensity;
in vec2 vUv;
out vec4 outColor;
void main() {
    vec4 prev = texture(uPrevTrail, vUv);
    vec4 faded = prev * uDecay;
    vec4 agents = texture(uAgentMask, vUv);
    vec4 added = agents * uIntensity;
    outColor = faded + added;
}
`;

const fsComposite = `#version 300 es
precision highp float;
uniform sampler2D uTrail;
uniform float uSpectrum[512]; 

in vec2 vUv;
out vec4 outColor;
void main() {
    vec4 trail = texture(uTrail, vUv);
    float d = clamp(trail.r + trail.g + trail.b, 0.0, 1.0);
    
    float lf = 0.0;
    float mf = 0.0;
    float hf = 0.0;

    for (int i = 1; i < 4; i++) {
        lf += uSpectrum[i];
    }
    lf /= 3.0;

    for (int i = 5; i < 32; i++) {
        mf += uSpectrum[i];
    }
    mf /= 27.0;

    for (int i = 33; i < 128; i++) {
        hf += uSpectrum[i];
    }
    hf /= 95.0;

    vec3 cold = vec3(0.0, 0.1, 0.2);
    vec3 hot  = vec3(1.0, 0.9, 0.6);
    vec3 heat = mix(cold, hot, d);

    float halo = 0.8 + lf * 0.6;
    float pulse = 0.7 + mf * 1.0;
    float jitter = 0.5 + hf * 1.5;

    vec3 glow = heat * halo;
    glow += trail.rgb * pulse;
    
    // Crackling jitter
    float noise = fract(sin(dot(vUv * 43758.5453, vec2(12.9898,78.233))) * 43758.5453);
    glow += vec3(hf * noise * 0.3);
    
    outColor = vec4(glow, 1.0);
}`;
}
