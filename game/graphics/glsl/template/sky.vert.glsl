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
  vec4 viewPosition = viewMatrixIn * vec4(attrCoord.xyz, 0.0);
  // restore w to 1.0 so that projection works
  vec4 position = projectionMatrixIn * vec4(viewPosition.xyz, 1.0);
  // set z to w so that it has maximum depth (1.0) after projection division
  gl_Position = vec4(position.xy, position.w, position.w);
}
