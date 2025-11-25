mat4 billboard(mat4 model_matrix, mat4 camera_matrix)
{
  float xLength = length(model_matrix[0]);
  float yLength = length(model_matrix[1]);
  float zLength = length(model_matrix[2]);
  return mat4(
    camera_matrix[0] * xLength,
    camera_matrix[1] * yLength,
    camera_matrix[2] * zLength,
    model_matrix[3]
  );
}

mat4 billboardX(mat4 model_matrix, mat4 camera_matrix)
{
  float xLength = length(model_matrix[0]);
  float yLength = length(model_matrix[1]);
  float zLength = length(model_matrix[2]);

  vec3 xAxis = vec3(1.0, 0.0, 0.0);
  vec3 zAxis = normalize(cross(xAxis, vec3(camera_matrix[1])));
  vec3 yAxis = cross(zAxis, xAxis);

  return mat4(
    vec4(xAxis * xLength, 0.0),
    vec4(yAxis * yLength, 0.0),
    vec4(zAxis * zLength, 0.0),
    model_matrix[3]
  );
}

mat4 billboardY(mat4 model_matrix, mat4 camera_matrix)
{
  float xLength = length(model_matrix[0]);
  float yLength = length(model_matrix[1]);
  float zLength = length(model_matrix[2]);

  vec3 yAxis = vec3(0.0, 1.0, 0.0);
  vec3 zAxis = normalize(cross(vec3(camera_matrix[0]), yAxis));
  vec3 xAxis = cross(yAxis, zAxis);

  return mat4(
    vec4(xAxis * xLength, 0.0),
    vec4(yAxis * yLength, 0.0),
    vec4(zAxis * zLength, 0.0),
    model_matrix[3]
  );
}

mat4 billboardZ(mat4 model_matrix, mat4 camera_matrix)
{
  float xLength = length(model_matrix[0]);
  float yLength = length(model_matrix[1]);
  float zLength = length(model_matrix[2]);

  vec3 zAxis = vec3(0.0, 0.0, 1.0);
  vec3 xAxis = normalize(cross(vec3(camera_matrix[1]), zAxis));
  vec3 yAxis = cross(zAxis, xAxis);

  return mat4(
    vec4(xAxis * xLength, 0.0),
    vec4(yAxis * yLength, 0.0),
    vec4(zAxis * zLength, 0.0),
    model_matrix[3]
  );
}
