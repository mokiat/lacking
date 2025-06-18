/* template "version.glsl" . */
/* template "attributes.glsl" . */
/* template "camera.glsl" . */
/* template "model.glsl" . */
/* template "armature.glsl" . */
/* template "textures.glsl" . */
/* template "uniforms.glsl" . */
/* template "varyings.glsl" . */
/* template "public.vert.glsl" . */

void main()
{
  /*- if .HasAttributeArmature */
  mat4 model_matrix =
    boneMatrixIn[attrJoints.x] * attrWeights.x + 
    boneMatrixIn[attrJoints.y] * attrWeights.y +
    boneMatrixIn[attrJoints.z] * attrWeights.z +
    boneMatrixIn[attrJoints.w] * attrWeights.w;
  /*- else */
  mat4 model_matrix = modelMatrixIn[gl_InstanceID];
  /*- end */

  gl_Position = projectionMatrixIn * (viewMatrixIn * (model_matrix * attrCoord));
}
