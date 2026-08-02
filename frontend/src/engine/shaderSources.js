// GLSL fragment shader sources for all node types
// Each shader renders a fullscreen quad with the node's effect

export const vertexShader = `#version 300 es
precision highp float;
in vec2 a_position;
in vec2 a_uv;
out vec2 v_uv;
void main() {
  v_uv = a_uv;
  gl_Position = vec4(a_position, 0.0, 1.0);
}
`

// ============ SOURCE SHADERS ============

export const solidColorShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform vec4 u_color;
out vec4 fragColor;
void main() {
  fragColor = u_color;
}
`

export const gradientShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform vec3 u_color1;
uniform vec3 u_color2;
uniform float u_angle;
out vec4 fragColor;
void main() {
  float rad = u_angle * 3.14159265 / 180.0;
  vec2 dir = vec2(cos(rad), sin(rad));
  float d = dot(v_uv - 0.5, dir) + 0.5;
  d = clamp(d, 0.0, 1.0);
  vec3 color = mix(u_color1, u_color2, d);
  fragColor = vec4(color, 1.0);
}
`

// ============ EFFECT SHADERS ============

export const blurHShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_input;
uniform vec2 u_resolution;
uniform float u_radius;
out vec4 fragColor;
void main() {
  vec2 texel = 1.0 / u_resolution;
  vec4 result = vec4(0.0);
  float total = 0.0;
  float r = u_radius;
  for (float x = -r; x <= r; x += 1.0) {
    float weight = exp(-x * x / (2.0 * r * r));
    result += texture(u_input, v_uv + vec2(x * texel.x, 0.0)) * weight;
    total += weight;
  }
  fragColor = result / total;
}
`

export const blurVShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_input;
uniform vec2 u_resolution;
uniform float u_radius;
out vec4 fragColor;
void main() {
  vec2 texel = 1.0 / u_resolution;
  vec4 result = vec4(0.0);
  float total = 0.0;
  float r = u_radius;
  for (float y = -r; y <= r; y += 1.0) {
    float weight = exp(-y * y / (2.0 * r * r));
    result += texture(u_input, v_uv + vec2(0.0, y * texel.y)) * weight;
    total += weight;
  }
  fragColor = result / total;
}
`

export const transformShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_input;
uniform vec2 u_resolution;
uniform vec2 u_translation;
uniform vec2 u_scale;
uniform float u_rotation;
uniform float u_opacity;
out vec4 fragColor;
void main() {
  vec2 center = vec2(0.5);
  vec2 uv = v_uv - center;

  // Apply rotation
  float rad = u_rotation * 3.14159265 / 180.0;
  float c = cos(rad);
  float s = sin(rad);
  uv = vec2(uv.x * c - uv.y * s, uv.x * s + uv.y * c);

  // Apply scale
  uv /= u_scale;

  // Apply translation
  uv += center;
  uv -= u_translation / u_resolution;

  // Sample
  if (uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0) {
    fragColor = vec4(0.0);
  } else {
    vec4 color = texture(u_input, uv);
    fragColor = vec4(color.rgb, color.a * u_opacity);
  }
}
`

export const mergeShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_fg;
uniform sampler2D u_bg;
uniform int u_mode;
out vec4 fragColor;

vec3 blendNormal(vec3 fg, vec3 bg) { return fg; }
vec3 blendMultiply(vec3 fg, vec3 bg) { return fg * bg; }
vec3 blendScreen(vec3 fg, vec3 bg) { return 1.0 - (1.0 - fg) * (1.0 - bg); }
vec3 blendOverlay(vec3 fg, vec3 bg) {
  return mix(2.0 * fg * bg, 1.0 - 2.0 * (1.0 - fg) * (1.0 - bg), step(0.5, bg));
}
vec3 blendAdd(vec3 fg, vec3 bg) { return min(fg + bg, 1.0); }
vec3 blendSubtract(vec3 fg, vec3 bg) { return max(bg - fg, 0.0); }
vec3 blendDifference(vec3 fg, vec3 bg) { return abs(fg - bg); }
vec3 blendLighten(vec3 fg, vec3 bg) { return max(fg, bg); }
vec3 blendDarken(vec3 fg, vec3 bg) { return min(fg, bg); }
vec3 blendColorDodge(vec3 fg, vec3 bg) {
  return mix(min(bg / (1.0 - fg), 1.0), vec3(1.0), step(1.0 - fg, 0.001));
}
vec3 blendColorBurn(vec3 fg, vec3 bg) {
  return mix(1.0 - min((1.0 - bg) / max(fg, 0.001), 1.0), vec3(0.0), step(fg, 0.001));
}

void main() {
  vec4 fg = texture(u_fg, v_uv);
  vec4 bg = texture(u_bg, v_uv);

  vec3 result;
  if (u_mode == 0) result = blendNormal(fg.rgb, bg.rgb);
  else if (u_mode == 1) result = blendMultiply(fg.rgb, bg.rgb);
  else if (u_mode == 2) result = blendScreen(fg.rgb, bg.rgb);
  else if (u_mode == 3) result = blendOverlay(fg.rgb, bg.rgb);
  else if (u_mode == 4) result = blendAdd(fg.rgb, bg.rgb);
  else if (u_mode == 5) result = blendSubtract(fg.rgb, bg.rgb);
  else if (u_mode == 6) result = blendDifference(fg.rgb, bg.rgb);
  else if (u_mode == 7) result = blendLighten(fg.rgb, bg.rgb);
  else if (u_mode == 8) result = blendDarken(fg.rgb, bg.rgb);
  else if (u_mode == 9) result = blendColorDodge(fg.rgb, bg.rgb);
  else if (u_mode == 10) result = blendColorBurn(fg.rgb, bg.rgb);
  else result = fg.rgb;

  float alpha = max(fg.a, bg.a);
  fragColor = vec4(result, alpha);
}
`

export const colorCorrectShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_input;
uniform float u_brightness;
uniform float u_contrast;
uniform float u_saturation;
uniform float u_hue;
out vec4 fragColor;

vec3 rgb2hsv(vec3 c) {
  vec4 K = vec4(0.0, -1.0/3.0, 2.0/3.0, -1.0);
  vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
  vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));
  float d = q.x - min(q.w, q.y);
  float e = 1.0e-10;
  return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c) {
  vec4 K = vec4(1.0, 2.0/3.0, 1.0/3.0, 3.0);
  vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
  return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
  vec4 color = texture(u_input, v_uv);

  // Brightness
  color.rgb += u_brightness;

  // Contrast
  color.rgb = (color.rgb - 0.5) * (1.0 + u_contrast) + 0.5;

  // Saturation via HSV
  vec3 hsv = rgb2hsv(color.rgb);
  hsv.y *= (1.0 + u_saturation);
  color.rgb = hsv2rgb(hsv);

  // Hue rotation
  hsv = rgb2hsv(color.rgb);
  hsv.x = fract(hsv.x + u_hue / 360.0);
  color.rgb = hsv2rgb(hsv);

  fragColor = clamp(color, 0.0, 1.0);
}
`

export const chromaKeyShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_input;
uniform vec3 u_keyColor;
uniform float u_similarity;
uniform float u_smoothness;
out vec4 fragColor;

void main() {
  vec4 color = texture(u_input, v_uv);
  float d = distance(color.rgb, u_keyColor);
  float alpha = smoothstep(u_similarity, u_similarity + u_smoothness, d);
  fragColor = vec4(color.rgb, color.a * alpha);
}
`

export const passthroughShader = `#version 300 es
precision highp float;
in vec2 v_uv;
uniform sampler2D u_input;
out vec4 fragColor;
void main() {
  fragColor = texture(u_input, v_uv);
}
`

// ============ BLEND MODE MAP ============
export const blendModeMap = {
  normal: 0,
  multiply: 1,
  screen: 2,
  overlay: 3,
  add: 4,
  subtract: 5,
  difference: 6,
  lighten: 7,
  darken: 8,
  colordodge: 9,
  colorburn: 10,
}
