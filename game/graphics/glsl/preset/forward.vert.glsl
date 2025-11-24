/* template "version.glsl" . */
/* template "attributes.glsl" . */
/* template "camera.glsl" . */
/* template "model.glsl" . */
/* template "timing.glsl" . */
/* template "armature.glsl" . */
/* template "textures.glsl" . */
/* template "uniforms.glsl" . */
/* template "varyings.glsl" . */
/* template "public.glsl" . */
/* template "public.vert.glsl" . */

smooth out float spawnTimeInOut;

void main()
{
  spawnTimeInOut = timeIn - timingIn[gl_InstanceID].x;
  /*- if .HasAttributeCoord */
  vec4 coord_ls = attrCoord;
  /*- else */
  vec4 coord_ls = vec4(0.0, 0.0, 0.0, 1.0);
  /*- end */
  /*- if .HasAttributeNormal */
  vec3 normal_ls = attrNormal;
  /*- else */
  vec3 normal_ls = vec3(0.0, 0.0, 1.0);
  /*- end */
  /*- if .HasAttributeTangent */
  vec3 tangent_ls = attrTangent;
  /*- else */
  vec3 tangent_ls = vec3(1.0, 0.0, 0.0);
  /*- end */
  /*- if .HasAttributeTexCoord */
  vec2 tex_coord = attrTexCoord;
  /*- else */
  vec2 tex_coord = vec2(0.0, 0.0);
  /*- end */
  /*- if .HasAttributeColor */
  vec4 color = attrColor;
  /*- else */
  vec4 color = vec4(1.0);
  /*- end */
  /*- if .HasAttributeArmature */
  mat4 model_matrix =
    boneMatrixIn[attrJoints.x] * attrWeights.x + 
    boneMatrixIn[attrJoints.y] * attrWeights.y +
    boneMatrixIn[attrJoints.z] * attrWeights.z +
    boneMatrixIn[attrJoints.w] * attrWeights.w;
  /*- else */
  mat4 model_matrix = modelMatrixIn[gl_InstanceID];
  /*- end */
  vec4 position = vec4(0.0, 0.0, 0.0, 1.0);
  /*- if .MainStatements */
  /*- range $statement := .MainStatements */
    /* $statement */
  /*- end */
  /*- else */
  position = projectionMatrixIn * (viewMatrixIn * (model_matrix * coord_ls));
  /*- end */
  gl_Position = position;
}
