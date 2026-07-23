/* template "version.glsl" . */
/* template "precision.glsl" . */
/* template "outputs.glsl" . */
/* template "camera.glsl" . */
/* template "timing.glsl" . */
/* template "textures.glsl" . */
/* template "uniforms.glsl" . */
/* template "varyings.glsl" . */
/* template "public.glsl" . */
/* template "public.frag.glsl" . */

smooth in float spawnTimeInOut;
smooth in float custom0InOut;
smooth in float custom1InOut;
smooth in float custom2InOut;
smooth in vec3 normalInOut;
smooth in vec3 tangentInOut;
smooth in vec2 texCoordInOut;
smooth in vec4 colorInOut;

void main()
{
  vec4 color = colorInOut;
  float metallic = 0.0;
  vec3 normal_ws = normalize(normalInOut);
  float roughness = 1.0;
  /*- if .MainStatements */
  /*- range $statement := .MainStatements */
    /* $statement */
  /*- end */
  /*- end */
  fbColor0Out = vec4(color.xyz, metallic);
  fbColor1Out = vec4(normal_ws, roughness);
}
