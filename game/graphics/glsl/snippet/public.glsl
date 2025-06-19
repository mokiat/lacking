mat3 extractRotation(mat4 matrix)
{
  // TODO: There might be a faster way if skewing is ignored.
  return inverse(transpose(mat3(matrix)));
}

vec3 normalFromTexel(vec3 texel, float scale)
{
  return normalize((texel * 2.0 - vec3(1.0)) * vec3(scale, scale, 1.0));
}

vec3 vectorToSurface(vec3 vector, vec3 normal, vec3 tangent)
{
  vec3 bitangent = cross(normal, tangent);
  mat3 tbn = mat3(tangent, bitangent, normal);
  return tbn * vector;
}
