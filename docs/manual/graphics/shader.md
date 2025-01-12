# Shader

The Lacking game engine uses a custom shading language called LSL. This language is then transpiled to the respective GPU API shader language (e.g. GLSL, WGSL).
The syntax of the language is very similar to the Go syntax, though there are some notable differences. Using Go-like syntax allows one to write in a consistent way game code and shader code. Furthemore, the Go syntax is designed to be easily and unambiguously parsable.

## Comments

Comments can be added with `//` anywhere in the file, where the remainder of the line is considered a comment and is ignored.

```
// An example comment.
```

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
| `\|=` | Assigns the bitwise OR operation on the variable and the value to the variable. |
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
| `\|` | Returns the result of a bitwise OR operation on the two expressions. Both sides need to be integer expressions of the same type. |
| `^` | Returns the result of a bitwise XOR operation on the two expressions. Both sides need to be integer expressions of the same type. |
| `&&` | Returns the result of a logical AND operation on the two expressions. Both sides need to be boolean expressions of the same type. |
| `\|\|` | Returns the result of a logical OR operation on the two expressions. Both sides need to be boolean expressions of the same type. |


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
