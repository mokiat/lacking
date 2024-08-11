# Shader

The lacking game engine uses a custom shading language. This language is then transpiled to the respective GPU API shader language (e.g. GLSL, WGSL).
The syntax of the language is very similar to the Go syntax, though there are some notable differences. Using Go-like syntax allows one to write in a consistent way game code and shader code. Furthemore, the Go syntax is designed to be easily and unambiguously parsable.

## Types

The following table lists the supported built-in types.

| Name | Description |
| ---- | ----------- |
| `bool` | Boolean type |
| `int` | 32 bit signed integer type |
| `uint` | 32 bit unsigned integer type |
| `float` | 32 bit floating point type |
| `vec2` | 2D vector type with two 32 bit floating point components |
| `vec3` | 3D vector type with three 32 bit floating point components |
| `vec4` | 4D vector type with four 32 bit floating point components |
| `bvec2` | 2D vector type with two boolean components |
| `bvec3` | 3D vector type with three boolean components |
| `bvec4` | 4D vector type with four boolean components |
| `ivec2` | 2D vector type with two 32 bit signed integer components |
| `ivec3` | 3D vector type with three 32 bit signed integer components |
| `ivec4` | 4D vector type with four 32 bit signed integer components |
| `uvec2` | 2D vector type with two 32 bit unsigned integer components |
| `uvec3` | 3D vector type with three 32 bit unsigned integer components |
| `uvec4` | 4D vector type with four 32 bit unsigned integer components |
| `mat2` | 2x2 matrix type with four 32 bit floating point components |
| `mat3` | 3x3 matrix type with nine 32 bit floating point components |
| `mat4` | 4x4 matrix type with sixteen 32 bit floating point components |
| `sampler2D` | sampler to a 2D texture |
| `samplerCube` | sampler to a Cube texture |

## Functions

The following table lists built-in functions.

| Name | Variants | Description |
| ---- | -------- | ----------- |
| `sin(float)` | `sin(vec2)`, `sin(vec3)`, `sin(vec4)` | Returns the sine of a value |
| `cos(float)` | `cos(vec2)`, `cos(vec3)`, `cos(vec4)` | Returns the cosine of a value |
