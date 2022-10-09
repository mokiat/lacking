# Theory

This page recaps and expands upon some fundamental physics theory. Knowledge in this area is necessary if one is to correctly model physics constraints.

## Terminology

| Symbol | Description |
| ------ | ----------- |
| $v$ | The [velocity](https://en.wikipedia.org/wiki/Velocity) of an object. Indicates the rate at which the object's position changes. |
| $m$ | The [mass](https://en.wikipedia.org/wiki/Mass) of an object. Larger mass means that the object is harder to translate. |
| $a$ | The [acceleration](https://en.wikipedia.org/wiki/Acceleration) of an object. Indicates the rate at which the velocity of an object changes. |
| $F$ | A [force](https://en.wikipedia.org/wiki/Force) acting on an object. |
| $p$ | The [momentum](https://en.wikipedia.org/wiki/Momentum) of an object. Can be thought of as the impact potential of the object. |
| | |
| $\omega$ | The [angular velocity](https://en.wikipedia.org/wiki/Angular_velocity) of an object. Indicates the rate at which the object's rotation changes. |
| $I$ | The [moment of inertia](https://en.wikipedia.org/wiki/Moment_of_inertia) of an object. Larger moment of inertia means that the object is harder to rotate. |
| $\alpha$ | The [angular acceleration](https://en.wikipedia.org/wiki/Angular_acceleration) of an object. Indicates the rate at which the angular velocity of an object changes. |
| $\tau$ | The [torque](https://en.wikipedia.org/wiki/Torque) acting on an object. |
| $L$| The [angular momentum](https://en.wikipedia.org/wiki/Angular_momentum) of an object. Can be thought of as the rotational impact potential of the object. |


## Equations

| Equation | Description |
| -------- | ----------- |
| $a = \frac{F}{m}$ | The acceleration that an object experiences can be derived from the force acting on the object and its mass. |
| $p = mv$ | The momentum is proportional to the mass and its relative velocity. |
| | |
| $\alpha = \frac{\tau}{I}$ | The angular acceleration that an object experiences can be dervied from the torque acting on the object and its angular momentum. |
| $L = I\omega$ | The angular momentum is proportional to the moment of inertia of the object and its angular velocity. |

## Representation in 3D

It is important to note that the above equations look slightly different in 3D space. The reason for this is that concepts like velocity and acceleration are actually vectors, whereas mass and moment of inertia are matrices.

Example:

$$
\Delta
\begin{bmatrix}
v_x \\
v_y \\
v_z \\
\end{bmatrix}
=
\begin{bmatrix}
a_x \\
a_y \\
a_z \\
\end{bmatrix}
\Delta{t}
$$

Example:

$$
\vec{F}=M\vec{a}
$$

$$
\begin{bmatrix}
F_x \\
F_y \\
F_z \\
\end{bmatrix}
=
\begin{bmatrix}
m & 0 & 0 \\
0 & m & 0 \\
0 & 0 & m \\
\end{bmatrix}
\begin{bmatrix}
a_x \\
a_y \\
a_z \\
\end{bmatrix}
$$

Example:

$$
\vec{\tau}=I\vec{\alpha}
$$

$$
\begin{bmatrix}
\tau_x \\
\tau_y \\
\tau_z \\
\end{bmatrix}
=
\begin{bmatrix}
I_{xx} & I_{xy} & I_{xz} \\
I_{yx} & I_{yy} & I_{yz} \\
I_{zx} & I_{zy} & I_{zz} \\
\end{bmatrix}
\begin{bmatrix}
\alpha_x \\
\alpha_y \\
\alpha_z \\
\end{bmatrix}
$$
