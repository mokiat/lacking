/* template "version.glsl" . */
/* template "precision.glsl" . */
/* template "outputs.glsl" . */
/* template "camera.glsl" . */
/* template "textures.glsl" . */
/* template "uniforms.glsl" . */
/* template "varyings.glsl" . */

smooth in vec3 normalInOut;
smooth in vec3 tangentInOut;
smooth in vec2 texCoordInOut;
smooth in vec4 colorInOut;

vec3 mapNormal(vec3 texel, float scale)
{
  vec3 normal_ls = (texel * 2.0 - vec3(1.0)) * vec3(scale, scale, 1.0);
  vec3 surface_normal_ws = normalize(normalInOut);
  vec3 surface_tangent_ws = normalize(tangentInOut);
  vec3 surface_bitangent_ws = normalize(cross(surface_normal_ws, surface_tangent_ws));
  mat3 tbn = mat3(surface_tangent_ws, surface_bitangent_ws, surface_normal_ws);
  return tbn * normalize(normal_ls);
}

void main()
{
  vec3 normal = normalize(normalInOut);
  vec2 tex_coord = texCoordInOut;
  vec4 vertex_color = colorInOut;
  float metallic = 0.0;
  float roughness = 1.0;
  vec4 color = vertex_color;
/*- range $statement := .MainStatements */
  /* $statement */
/*- end */
  fbColor0Out = vec4(color.xyz, metallic);
  fbColor1Out = vec4(normal, roughness);
}
