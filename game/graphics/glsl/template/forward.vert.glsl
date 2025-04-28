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
	mat4 boneMatrixA = boneMatrixIn[attrJoints.x];
	mat4 boneMatrixB = boneMatrixIn[attrJoints.y];
	mat4 boneMatrixC = boneMatrixIn[attrJoints.z];
	mat4 boneMatrixD = boneMatrixIn[attrJoints.w];
	vec4 worldPosition =
		boneMatrixA * (attrCoord * attrWeights.x) +
		boneMatrixB * (attrCoord * attrWeights.y) +
		boneMatrixC * (attrCoord * attrWeights.z) +
		boneMatrixD * (attrCoord * attrWeights.w);
	/* else */
	mat4 modelMatrix = modelMatrixIn[gl_InstanceID];
	vec4 worldPosition = modelMatrix * attrCoord;
	/* end */
  gl_Position = projectionMatrixIn * (viewMatrixIn * worldPosition);  
}
