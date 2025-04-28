/*- if .HasAttributeCoord */
layout(location = 0) in vec4 attrCoord;
/*- end */
/*- if .HasAttributeNormal */
layout(location = 1) in vec3 attrNormal;
/*- end */
/*- if .HasAttributeTangent */
layout(location = 2) in vec3 attrTangent;
/*- end */
/*- if .HasAttributeTexCoord */
layout(location = 3) in vec2 attrTexCoord;
/*- end */
/*- if .HasAttributeColor */
layout(location = 4) in vec4 attrColor;
/*- end */
/*- if .HasAttributeArmature */
layout(location = 5) in vec4 attrWeights;
layout(location = 6) in uvec4 attrJoints;
/*- end */
