export const vertexShaderSource = `#version 300 es

// An attribute is an input to a vertex shader.
// It will receive data from a buffer.
in vec4 a_position;
in vec2 a_texcoord;
in vec4 a_color;

// Translation to add to position
uniform vec2 u_translation;

out highp vec2 v_texcoord;
out highp vec4 v_color;

// All shaders have a main function.
void main() {

  vec4 position = a_position + vec4(u_translation, 0, 0);

  // gl_Position is a special variable a vertex shader is responsible for setting
  gl_Position = position;

  v_texcoord = a_texcoord;
  v_color = a_color;
}
`;

export const fragmentShaderSource = `#version 300 es

// Fragment shaders don't have a default precision so we need to pick one.
// highp (high precision) is a good default.
precision highp float;

in highp vec2 v_texcoord;
in highp vec4 v_color;

// Take a uniform color input.
uniform vec3 u_color;

uniform sampler2D u_texture;

// We need to declare an output for the fragment shader.
out vec4 outColor;

float median(float r, float g, float b) {
  return max(min(r, g), min(max(r, g), b));
}


// void main() {
//   vec4 texColor = texture(u_texture, v_texcoord);
//   float sigDist = median(texColor.r, texColor.g, texColor.b) - 0.5;

//   float alpha = clamp(sigDist / fwidth(sigDist) + 0.5, 0.0, 1.0);
  
//   // float alpha = sigDist > 0.0 ? 1.0 : 0.0;

//   // vec2 ddist = vec2(dFdx(sigDist), dFdy(sigDist));
//   // float alpha = sigDist / length(ddist) + 0.5;
//   // float alpha = clamp(sigDist / length(ddist) + 0.5, 0.0, 1.0);

//   outColor = mix(vec4(u_color, 1.0), vec4(0.0, 0.0, 0.0, 0.0),  1.0-alpha);
  
//   if (outColor.a < 0.0001) discard;
// }

float screenPxRange() {
  return 6.0;
}

float norm(float x, float a, float b) {
  return b != a ? (x - a) / (b - a) : 0.0;
}

float fitClamped(float x, float a, float b, float c, float d) {
  return c + (d - c) * clamp(norm(x, a, b), 0.0, 1.0);
}

void main() {
  const vec4 bgColor = vec4(0.0, 0.0, 0.0, 0.0);
  vec3 msd = texture(u_texture, v_texcoord).rgb;
  float sd = median(msd.r, msd.g, msd.b);

  // Original
  
  // float screenPxDistance = screenPxRange()*(sd - 0.5);
  // float opacity = clamp(screenPxDistance + 0.5, 0.0, 1.0);
  // outColor = mix(v_color, vec4(u_color, 1.0), opacity);

  // https://observablehq.com/@stwind/msdf-text-label
  
  vec2 duv = fwidth(v_texcoord);
  // antialising factor
  float su = 0.0078125;
  float aa = max(dot(vec2(su), 0.5 / duv), 1.0);
  // edge boosting for low resolutions
  float boost = fitClamped(0.5 * su / length(duv), 1.0, 0.5, 0.0, 0.5);
  float t = clamp((sd - 0.5) * aa + 0.5, 0.0, 1.0);
  t = norm(t, 0.0, 1.0 - boost);
  outColor = mix(v_color, vec4(u_color, 1.0), t);
}



`;
