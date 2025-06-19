# Shader

The Lacking game engine uses a custom shading language called LSL. This language is then transpiled to the respective GPU API shader language (e.g. GLSL, WGSL).
The syntax of the language is very similar to the Go syntax, though there are some notable differences. Using Go-like syntax allows one to write in a consistent way game code and shader code. Furthemore, the Go syntax is designed to be easily and unambiguously parsable.

## Comments

Comments can be added with `//` anywhere in the file, where the remainder of the line is considered a comment and is ignored.

```
// An example comment.
```

The LSL does not have support for multi-line comments.

## Literals

Literals are a mechanism to provide constant in-place values to a shader.

| Type | Example | Description |
| ---- | ------- | ----------- |
| bool | `true`, `false` | Specifies a boolean value. |
| int | `13`, `-24` | Specifies a 32bit signed integer value. |
| float | `-5.3`, `2.4` | Specifies a 32bit floating point value. |

## Operators

There are a number of operators supported by the LSL language.

### Assignment Operators

Following is a list of assignment operators that can be used to assign a value to a variable.

| Operator | Description |
| -------- | ----------- |
| `=` | Assigns the value to the variable. Both sides need to have the same type. |
| `:=` | Defines a new variable and initializes it with the contents and type of the value. |
| `+=` | Adds the value to the variable. |
| `-=` | Subtracts the value from the variable. |
| `*=` | Multiplies the variable by the value. |
| `/=` | Divides the variable by the value. |
| `%=` | Assigns the modulo of the variable and the value. |
| `<<=` | Shifts the variable to the left by the value number of bits. |
| `>>=` | Shifts the variable to the right by the value number of bits. |
| `&=` | Assigns the bitwise AND operation on the variable and the value to the variable. |
| `│=` | Assigns the bitwise OR operation on the variable and the value to the variable. |
| `^=` | Assigns the bitwise XOR operation on the variable and the value to the variable. |

### Unary Operators

The following unary operators can be used inside expressions to transform a single sub-expression.

| Operator | Description |
| -------- | ----------- |
| `!` | Inverts the value of the expression. It needs to be a boolean expression or a boolean vector expression. |
| `-` | Negates the value of the expression. It needs to be a numeric expression or a numeric vector expression. |
| `+` | A no-op change to the value of the expression. it needs to be a numeric expression vector expression. |
| `^` | Performs a bitwise NOT operation on the value of the expression. It needs to be a numeric expression or a numeric vector expression. |

### Binary Operators

The following binary operators can be used inside expressions to combine two sub-expressions.

| Operator | Description |
| -------- | ----------- |
| `+` | Returns the sum of the two expressions. Both sides need to be numeric expressions of the same type. |
| `-` | Returns the difference between the two expressions. Both sides need to be numeric expressions of the same type. |
| `*` | Returns the multiplication of the two expressions. Both sides need to be numeric expressions of the same type. |
| `/` | Returns the division of the two expressions. Both sides need to be numeric expressions of the same type. |
| `%` | Returns the modulo of the two expressions. Both sides need to be integer expressions of the same type. |
| `<<` | Returns the bitwise left-shift of the left expression by the right expression. Both sides need to be integer expressions. |
| `>>` | Returns the bitwise right-shift of the left expression by the right expression. Both sides need to be integer expressions. |
| `==` | Returns a boolean value indicating whether the two expressions are equal. Both sides need to be expressions of the same type and comparable. |
| `!=` | Returns a boolean value indicating whether the two expressions are different. Both sides need to be expressions of the same type and comparable. |
| `<` | Returns a boolean value indicating whether the first expression is smaller than the second expression. Both sides need to be expressions of the same type and be ordered. |
| `>` | Returns a boolean value indicating whether the second expression is smaller than the first expression. Both sides need to be expressions of the same type and be ordered. |
| `<=` | Returns a boolean value indicating whether the first expression is smaller than or equal to the second expression. Both sides need to be expressions of the same type and be ordered. |
| `>=` | Returns a boolean value indicating whether the second expression is smaller than or equal to the first expression. Both sides need to be expressions of the same type and be ordered. |
| `&` | Returns the result of a bitwise AND operation on the two expressions. Both sides need to be integer expressions of the same type. |
| `│` | Returns the result of a bitwise OR operation on the two expressions. Both sides need to be integer expressions of the same type. |
| `^` | Returns the result of a bitwise XOR operation on the two expressions. Both sides need to be integer expressions of the same type. |
| `&&` | Returns the result of a logical AND operation on the two expressions. Both sides need to be boolean expressions of the same type. |
| `││` | Returns the result of a logical OR operation on the two expressions. Both sides need to be boolean expressions of the same type. |

The operator precedence is similar to the official Go one and is described in the following table (higher is applied first).

| Precedence | Operator |
| ---------- | -------- |
| 5 | `*`, `/`, `%`, `<<`, `>>`, `&` |
| 4 | `+`, `-`, `│`, `^` |
| 3 | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| 2 | `&&` |
| 1 | `││` |

Operators with the same precedence associate from left to right (i.e. the operators are applied from left to right).

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


## Textures

Shaders often require textures to read texels off of. Such dependencies to external resources are declared through the `texture` keyword. It works in a similar way to the Go's `var` keyword, except that an initial value cannot be specified.

```
texture color sampler2D
texture env samplerCube
```

or

```
texture (
  color sampler2D
  env samplerCube
)
```

Texture fields are globally visible within the shader and are read-only.

## Uniforms

Shaders often require external data so that they can be reused. This is done through the `uniform` keyword. It works in a similar way to the Go's `var` keyword, except that an initial value cannot be specified.

```
uniform color vec4
uniform intensity float
```

or

```
uniform (
  color vec4
  intensity float
)
```

Uniform fields are globally visible within the shader and are read-only.

## Varying

Shader code is used in both vertex and fragment stages. Sometimes data needs to be passed from one stage onto the next. This is achieved via the `varying` keyword. It works in a similar way to the Go's `var` keyword, except that an initial value cannot be specified.

```
varying normal vec3
```

or

```
varying (
  normal vec3
)
```

Varying fields are globally visible within the shader and are read-only in the fragment shader stage.

## Structs

It is possible to define custom struct types.

```
type Vertex struct {
  position vec3
  uv vec2
}
```

## Functions

The following table lists constructor built-in functions.

| Name | Scope | Description |
| ---- | ----- | ----------- |
| `bool(v int) bool` | unbounded | Converts an integer into a boolean. |
| `bool(v uint) bool` | unbounded | Converts an unsigned integer into a boolean. |
| `bool(v float) bool` | unbounded | Converts a float into a boolean. |
| `int(v bool) int` | unbounded | Converts a boolean into an integer. |
| `int(v uint) int` | unbounded | Converts an unsigned integer into an integer. |
| `int(v float) int` | unbounded | Converts a float into an integer. |
| `uint(v bool) uint` | unbounded | Converts a boolean into an unsigned integer. |
| `uint(v int) uint` | unbounded | Converts an integer into an unsigned integer. |
| `uint(v float) uint` | unbounded | Converts a float into an unsigned integer. |
| `float(v bool) float` | unbounded | Converts a boolean into a float. |
| `float(v int) float` | unbounded | Converts an integer into a float. |
| `float(v uint) float` | unbounded | Converst an unsigned integer into a float. |
| `vec2(v float) vec2` | unbounded | Returns a `vec2` with all components equal to the value `v`. |
| `vec2(x, y float) vec2` | unbounded | Returns a `vec2` with the components set to `x` and `y` respectively. |
| `vec3(v float) vec3` | unbounded | Returns a `vec3` with all components equal to the value `v`. |
| `vec3(x, y, z float) vec3` | unbounded | Returns a `vec3` with the components set to `x`, `y`, and `z` respectively. |
| `vec3(a vec2, z float) vec3` | unbounded | Returns a `vec3` with the components set to `a.x`, `a.y`, and `z` respectively. |
| `vec3(x float, a vec2) vec3` | unbounded | Returns a `vec3` with the components set to `x`, `a.x`, and `a.y` respectively. |
| `vec4(v float) vec4` | unbounded | Returns a `vec4` with all components equal to the value `v`. |
| `vec4(x, y, z, w float) vec4` | unbounded | Returns a `vec4` with the components set to `x`, `y`, `z`, and `w` respectively. |
| `vec4(a vec2, z, w float) vec4` | unbounded | Returns a `vec4` with the components set to `a.x`, `a.y`, `z`, and `w` respectively. |
| `vec4(x float, a vec2, w float) vec4` | unbounded | Returns a `vec4` with the components set to `x`, `a.x`, `a.y`, and `w` respectively. |
| `vec4(x, y float, a vec2) vec4` | unbounded | Returns a `vec4` with the components set to `x`, `y`, `a.x`, and `a.y` respectively. |
| `vec4(a vec3, w float) vec4` | unbounded | Returns a `vec4` with the components set to `a.x`, `a.y`, `a.z`, and `w` respectively. |
| `vec4(x float, a vec3) vec4` | unbounded | Returns a `vec4` with the components set to `x`, `a.x`, `a.y`, and `a.z` respectively. |
| `bvec2(v bool) bvec2` | unbounded | Returns a `bvec2` with all components equal to the value `v`. |
| `bvec2(x, y bool) bvec2` | unbounded | Returns a `bvec2` with the components set to `x` and `y` respectively. |
| `bvec3(v bool) bvec3` | unbounded | Returns a `bvec3` with all components equal to the value `v`. |
| `bvec3(x, y, z bool) bvec3` | unbounded | Returns a `bvec3` with the components set to `x`, `y`, and `z` respectively. |
| `bvec3(a bvec2, z bool) bvec3` | unbounded | Returns a `bvec3` with the components set to `a.x`, `a.y`, and `z` respectively. |
| `bvec3(x bool, a bvec2) bvec3` | unbounded | Returns a `bvec3` with the components set to `x`, `a.x`, and `a.y` respectively. |
| `bvec4(v bool) bvec4` | unbounded | Returns a `bvec4` with all components equal to the value `v`. |
| `bvec4(x, y, z, w bool) bvec4` | unbounded | Returns a `bvec4` with the components set to `x`, `y`, `z`, and `w` respectively. |
| `bvec4(a bvec2, z, w bool) bvec4` | unbounded | Returns a `bvec4` with the components set to `a.x`, `a.y`, `z`, and `w` respectively. |
| `bvec4(x bool, a bvec2, w bool) bvec4` | unbounded | Returns a `bvec4` with the components set to `x`, `a.x`, `a.y`, and `w` respectively. |
| `bvec4(x, y bool, a bvec2) bvec4` | unbounded | Returns a `bvec4` with the components set to `x`, `y`, `a.x`, and `a.y` respectively. |
| `bvec4(a bvec3, w bool) bvec4` | unbounded | Returns a `bvec4` with the components set to `a.x`, `a.y`, `a.z`, and `w` respectively. |
| `bvec4(x bool, a bvec3) bvec4` | unbounded | Returns a `bvec4` with the components set to `x`, `a.x`, `a.y`, and `a.z` respectively. |
| `ivec2(v int) ivec2` | unbounded | Returns a `ivec2` with all components equal to the value `v`. |
| `ivec2(x, y int) ivec2` | unbounded | Returns a `ivec2` with the components set to `x` and `y` respectively. |
| `ivec3(v int) ivec3` | unbounded | Returns a `ivec3` with all components equal to the value `v`. |
| `ivec3(x, y, z int) ivec3` | unbounded | Returns a `ivec3` with the components set to `x`, `y`, and `z` respectively. |
| `ivec3(a ivec2, z int) ivec3` | unbounded | Returns a `ivec3` with the components set to `a.x`, `a.y`, and `z` respectively. |
| `ivec3(x int, a ivec2) ivec3` | unbounded | Returns a `ivec3` with the components set to `x`, `a.x`, and `a.y` respectively. |
| `ivec4(v int) ivec4` | unbounded | Returns a `ivec4` with all components equal to the value `v`. |
| `ivec4(x, y, z, w int) ivec4` | unbounded | Returns a `ivec4` with the components set to `x`, `y`, `z`, and `w` respectively. |
| `ivec4(a ivec2, z, w int) ivec4` | unbounded | Returns a `ivec4` with the components set to `a.x`, `a.y`, `z`, and `w` respectively. |
| `ivec4(x int, a ivec2, w int) ivec4` | unbounded | Returns a `ivec4` with the components set to `x`, `a.x`, `a.y`, and `w` respectively. |
| `ivec4(x, y int, a ivec2) ivec4` | unbounded | Returns a `ivec4` with the components set to `x`, `y`, `a.x`, and `a.y` respectively. |
| `ivec4(a ivec3, w int) ivec4` | unbounded | Returns a `ivec4` with the components set to `a.x`, `a.y`, `a.z`, and `w` respectively. |
| `ivec4(x int, a ivec3) ivec4` | unbounded | Returns a `ivec4` with the components set to `x`, `a.x`, `a.y`, and `a.z` respectively. |
| `uvec2(v uint) uvec2` | unbounded | Returns a `uvec2` with all components equal to the value `v`. |
| `uvec2(x, y uint) uvec2` | unbounded | Returns a `uvec2` with the components set to `x` and `y` respectively. |
| `uvec3(v uint) uvec3` | unbounded | Returns a `uvec3` with all components equal to the value `v`. |
| `uvec3(x, y, z uint) uvec3` | unbounded | Returns a `uvec3` with the components set to `x`, `y`, and `z` respectively. |
| `uvec3(a uvec2, z uint) uvec3` | unbounded | Returns a `uvec3` with the components set to `a.x`, `a.y`, and `z` respectively. |
| `uvec3(x uint, a uvec2) uvec3` | unbounded | Returns a `uvec3` with the components set to `x`, `a.x`, and `a.y` respectively. |
| `uvec4(v uint) uvec4` | unbounded | Returns a `uvec4` with all components equal to the value `v`. |
| `uvec4(x, y, z, w uint) uvec4` | unbounded | Returns a `uvec4` with the components set to `x`, `y`, `z`, and `w` respectively. |
| `uvec4(a uvec2, z, w uint) uvec4` | unbounded | Returns a `uvec4` with the components set to `a.x`, `a.y`, `z`, and `w` respectively. |
| `uvec4(x uint, a uvec2, w uint) uvec4` | unbounded | Returns a `uvec4` with the components set to `x`, `a.x`, `a.y`, and `w` respectively. |
| `uvec4(x, y uint, a uvec2) uvec4` | unbounded | Returns a `uvec4` with the components set to `x`, `y`, `a.x`, and `a.y` respectively. |
| `uvec4(a uvec3, w uint) uvec4` | unbounded | Returns a `uvec4` with the components set to `a.x`, `a.y`, `a.z`, and `w` respectively. |
| `uvec4(x uint, a uvec3) uvec4` | unbounded | Returns a `uvec4` with the components set to `x`, `a.x`, `a.y`, and `a.z` respectively. |
| `mat2(x, y vec2) mat2` | unbounded | Returns a `mat2` with columns `x` and `y` in order. |
| `mat2(m mat3) mat2` | unbounded | Returns a `mat2` using the upper left portion of the provided matrix. |
| `mat3(x, y, z vec3) mat3` | unbounded | Returns a `mat3` with columns `x`, `y` and `z` in order. |
| `mat3(m mat4) mat3` | unbounded | Returns a `mat3` using the upper left portion of the provided matrix. |
| `mat4(x, y, z, w vec4) mat4` | unbounded | Returns a `mat4` with columns `x`, `y`, `z` and `w` in order. |


The following table lists general math built-in functions.

| Name | Scope | Description |
| ---- | ----- | ----------- |
| `abs(v T) T` | unbounded | Returns the absolute value of the parameter. |
| `sign(v T) T` | unbounded | Returns `1` when positive, `0` when zero, and `-1` when negative. |
| `floor(v T) T` | unbounded | Returns the nearest whole number less than or equal to the parameter. |
| `trunc(v T) T` | unbounded | Returns the nearest whole number for which the absolute value is less or equal to the absolute value of the parameter. |
| `round(v T) T` | unbounded | Rounds the value to the nearest whole number. |
| `ceil(v T) T` | unbounded | Returns the nearest whole number greater than or equal to the parameter. |
| `fract(v T) T` | unbounded | In essence, returns `x - floor(x)`. |
| `min(a, b T) T` | unbounded | Returns the minimum of the two values. |
| `max(a, b T) T` | unbounded | Returns the maximum of the two values. |
| `clamp(v, lower, upper T) T` | unbounded | In essence, returns `max(lower, min(v, upper))`. |
| `mix(a, b, z T) T` | unbounded | Returns the linear interpolation between `a` and `b`. In essence, it calculates `a * (1-z) + b * z`. |
| `smoothstep(a, b, z T) T` | unbounded | Returns the Hermite interpolation between `a` and `b`. |
| `length(v T) S` | unbounded | Returns the length of the vector. |
| `distance(a, b T) S` | unbounded | Returns the distance between two vectors. In essence, it returns `length(b - a)`. |
| `dot(a, b T) S` | unbounded | Returns the dot product of two vectors. |
| `cross(a, b T) T` | unbounded | Returns the cross product of two vectors. |
| `normalize(v T) T` | unbounded | Resizes a vector to the unit length. |
| `faceforward(v, i, n T)` | unbounded | Returns the vector `v` oriented to point away from the surface as dictated by in normal vector `n` and the incident vector `i` (which visually "points" towards the surface). |
| `reflect(i, n T) T` | unbounded | Reflects the incident vector `i` (which visually "points" towards the surface) according to the normal vector `n`. |
| `refract(i, n T, e S) T` | unbounded | Refracts the incident vector `i` (which visually "points" towards the surface) according to the normal vector `n` and the ratio of refraction `e`. |
| `transpose(m T) T` | unbounded | Returns the transpose of the matrix `m`. As only square matrices are supported right now, the returned type is always the same. |
| `determinant(m T) K` | unbounded | Returns the determinant of the matrix `m`. |
| `any(v T) bool` | unbounded | Returns `true` if any of the components of `v` are `true`. |
| `all(v T) bool` | unbounded | Returns `true` if all of the components of `v` are `true`. |
| `cos(v T) T` | unbounded | Returns the cosine of the parameter `v`. |
| `sin(v T) T` | unbounded | Returns the sine of the parameter `v`. |

The following table lists general texture built-in functions.

| Name | Scope | Description |
| ---- | ----- | ----------- |
| `sample(s sampler2D, uv vec2) vec4` | unbounded | Samples the specified 2D sampler and returns the value at position `uv`. |
| `sample(s samplerCube, uv vec3) vec4` | unbounded | Samples the specified Cube sampler and returns the value at position `uv`. |

The following table lists LSL helper functions.

| Name | Scope | Description |
| ---- | ----- | ----------- |
| `extractRotation(matrix mat4) mat3` | unbounded | Extracts the rotation matrix from a general 3D transformation matrix. |
| `normalFromTexel(texel vec3, scale float) vec3` | unbounded | Converts a texel value from a texture into a normal, scaled as specified. |
| `vectorToSurface(vector, normal, tangent vec3) vec3` | unbounded | Transforms the specified `vector` according to the coordinate space defined by `normal` and `tangent`. This is usually used in normal mapping to transform a normal from local space into face orientation space. |
| `billboard(model, camera mat4) mat4` | unbounded | Takes a model and camera matrices and calculates and returns a new model matrix that will transform the model so that it is always aligned towards the camera. |
| `billboardX(model, camera mat4) mat4` | unbounded | Takes a model and camera matrices and calculates and returns a new model matrix that will transform the model so that its X axis matches the world X axis and the remaining axes are aligned with the camera's. |
| `billboardY(model, camera mat4) mat4` | unbounded | Takes a model and camera matrices and calculates and returns a new model matrix that will transform the model so that its Y axis matches the world Y axis and the remaining axes are aligned with the camera's. |
| `billboardZ(model, camera mat4) mat4` | unbounded | Takes a model and camera matrices and calculates and returns a new model matrix that will transform the model so that its Z axis matches the world Z axis and the remaining axes are aligned with the camera's. |

## Predefined Variables

Following are variables that are pre-defined for some of the render stages.

| Name | Type | Mode | Scope | Description |
| ---- | ---- | ---- | ----- | ----------- |
| `#vertexCoord` | `vec4` | ReadOnly | `#vertex` (shadow, geometry, forward) | Contains the position of the vertex in local space. Value is `vec4(0.0, 0.0, 0.0, 1.0)` if the mesh does not contain vertex coords. |
| `#vertexNormal` | `vec3` | ReadOnly | `#vertex` (shadow, geometry, forward) | Contains the normal of the vertex in local space. Value is `vec3(0.0, 0.0, 1.0)` if the mesh does not contain vertex normals. |
| `#vertexTangent` | `vec3` | ReadOnly | `#vertex` (shadow, geometry, forward) | Contains the tangent of the vertex in local space. Value is `vec3(1.0, 0.0, 0.0)` if the mesh does not contain vertex tangents. |
| `#vertexUV` | `vec2` | ReadOnly | `#vertex` (shadow, geometry, forward) | Contains the texture coordinates of the vertex. Value is `vec2(0.0, 0.0)` if the mesh does not contain vertex texture coords. |
| `#vertexColor` | `vec4` | ReadOnly | `#vertex` (shadow, geometry, forward) | Contains the color of the vertex. Value is `vec4(1.0, 1.0, 1.0, 1.0)` if the mesh does not contain vertex coloring. |
| `#modelMatrix` | `mat4` | ReadOnly | `#vertex` (shadow, geometry, forward) | Contains the model transformation matrix for the rendered object. |
| `#cameraMatrix` | `mat4` | ReadOnly | `#vertex`, `#fragment` (shadow, geometry, forward, sky) | Contains the camera transformation matrix. NOTE: Not to be confused with `#viewMatrix`, which is often the desired one. |
| `#viewMatrix` | `mat4` | ReadOnly | `#vertex`, `#fragment` (shadow, geometry, forward, sky) | Contains the view transformation matrix. It is equal to the inverse of the `#cameraMatrix` but is already pre-calculated. |
| `#projectionMatrix` | `mat4` | ReadOnly | `#vertex`, `#fragment` (shadow, geometry, forward, sky) | Contains the projection matrix. |
| `#viewport` | `vec4` | ReadOnly | `#vertex`, `#fragment` (shadow, geometry, forward, sky) | The `x` and `y` fields contain the positioning of the viewport and `z` and `w` contain the size of the viewport. |
| `#position` | `vec4` | ReadWrite | `#vertex` (shadow, geometry, forward) | Contains the output position of the vertex from a vertex shader. |
| `#color` | `vec4` | ReadWrite | `#fragment` (geometry, forward, sky) | Specifies the color to be placed on the screen. Default value is `vec4(0.0, 0.0, 0.0, 1.0)`. |
| `#normal` | `vec3` | ReadWrite | `#fragment` (geometry) | Contains the output normal of a texel from a geometry fragment shader. |
| `#metallic` | `float` | ReadWrite | `#fragment` (geometry) | Contains the output metallic value of a texel from a geometry fragment shader. |
| `#roughness` | `float` | ReadWrite | `#fragment` (geometry) | Contains the output roughness value of a texel from a geometry fragment shader. |
| `#varyingNormal` | `vec3` | ReadWrite | `#vertex`, `#fragment` (geometry) | A varying variable used to transfer a normal value between shader stages in a geometry shader. If a `#vertex` function is not specified, this value is automatically filled. **NOTE:** Make sure to normalize before usage due to interpolation. |
| `#varyingTangent` | `vec3` | ReadWrite | `#vertex`, `#fragment` (geometry) | A varying variable used to transfer a tangent value between shader stages in a geometry shader. If a `#vertex` function is not specified, this value is automatically filled. **NOTE:** Make sure to normalize before usage due to interpolation. |
| `#varyingUV` | `vec2` | ReadWrite | `#vertex`, `#fragment` (geometry) | A varying variable used to transfer a texture coordinate value between shader stages in a geometry shader. If a `#vertex` function is not specified, this value is automatically filled. |
| `#varyingColor` | `vec4` | ReadWrite | `#vertex`, `#fragment` (geometry) | A varying variable used to transfer a color value between shader stages in a geometry shader. If a `#vertex` function is not specified, this value is automatically filled. |
| `#direction` | `vec3` | ReadOnly | `#fragment` (sky) | Contains the world space direction of the ray that is being rendered. |
