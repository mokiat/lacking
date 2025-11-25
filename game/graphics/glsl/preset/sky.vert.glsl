/* template "version.glsl" . */
/* template "attributes.glsl" . */
/* template "camera.glsl" . */
/* template "textures.glsl" . */
/* template "uniforms.glsl" . */
/* template "varyings.glsl" . */

smooth out vec3 varyingDirectionWS;

void main()
{
  varyingDirectionWS = attrCoord.xyz;
  // ensure that translations are ignored by setting w to 0.0
  vec4 position_vs = viewMatrixIn * vec4(attrCoord.xyz, 0.0);
  // restore w to 1.0 so that projection works
  vec4 position_clip = projectionMatrixIn * vec4(position_vs.xyz, 1.0);
  // set z to w so that it has maximum depth (1.0) after projection division
  gl_Position = vec4(position_clip.xy, position_clip.w, position_clip.w);
}
