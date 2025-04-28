/*- if .HasAttributeArmature */
layout (std140) uniform Armature
{
	mat4 boneMatrixIn[256];
};
/*- end */