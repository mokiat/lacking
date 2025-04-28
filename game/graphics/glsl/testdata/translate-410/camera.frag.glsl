#version 410

layout (std140) uniform Camera
{
	mat4 projectionMatrixIn;
	mat4 viewMatrixIn;
	mat4 cameraMatrixIn;
	vec4 viewportIn;
	float lackingTime; // FIXME: rename to timeIn
};

