/* template "version.glsl" . */
/* template "attributes.glsl" . */
/* template "camera.glsl" . */
/* template "model.glsl" . */
/* template "armature.glsl" . */
/* template "textures.glsl" . */
/* template "uniforms.glsl" . */
/* template "varyings.glsl" . */

void main()
{
  /* if .HasAttributeArmature */
  mat4 model_matrix =
    boneMatrixIn[attrJoints.x] * attrWeights.x + 
    boneMatrixIn[attrJoints.y] * attrWeights.y +
    boneMatrixIn[attrJoints.z] * attrWeights.z +
    boneMatrixIn[attrJoints.w] * attrWeights.w;
  /* else */
  mat4 model_matrix = modelMatrixIn[gl_InstanceID];
  /* end */
  vec4 position_ws = model_matrix * attrCoord;
  gl_Position = projectionMatrixIn * (viewMatrixIn * position_ws);
}
