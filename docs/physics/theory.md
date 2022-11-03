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
| $\omega$ | The [angular velocity](https://en.wikipedia.org/wiki/Angular_velocity) of an object. Indicates the rate at which the object's rotation changes. |
| $I$ | The [moment of inertia](https://en.wikipedia.org/wiki/Moment_of_inertia) of an object. Larger moment of inertia means that the object is harder to rotate. |
| $\alpha$ | The [angular acceleration](https://en.wikipedia.org/wiki/Angular_acceleration) of an object. Indicates the rate at which the angular velocity of an object changes. |
| $\tau$ | The [torque](https://en.wikipedia.org/wiki/Torque) acting on an object. |
| $L$| The [angular momentum](https://en.wikipedia.org/wiki/Angular_momentum) of an object. Can be thought of as the rotational impact potential of the object. |
| $\textit{e}$ | The [coefficient of restitution](https://en.wikipedia.org/wiki/Coefficient_of_restitution) of two objects. It describes the bounciness that an object experiences when colliding with another object, where the value differs depending on the characteristics of the two objects. |
| $\mu$ | The [coefficient of friction](https://en.wikipedia.org/wiki/Friction) of two objects. It describes how hard it is for two objects to slide when in contact, where the value differs depending on the characteristics of the two objects. It also differs depending on whether the two objects are initially at relative rest or are already sliding across each other. |
| $\rho$ | The [density](https://en.wikipedia.org/wiki/Density) of a fluid or object. Determines how much mass there is per unit of volume. For a fluid it also determines how hard it is for an object to go through the fluid. |
| $C_d$ | The [drag coefficient](https://en.wikipedia.org/wiki/Drag_coefficient) of an object. It is determined by the shape of an object and is specific to a given orientation of the object. It describes how much an object's shape opposes the object's motion through a fluid. |


## Equations

| Equation | Description |
| -------- | ----------- |
| $a = \frac{F}{m}$ | The acceleration that an object experiences can be derived from the force acting on the object and its mass. |
| $p = mv$ | The momentum is proportional to the mass and its relative velocity. |
| $\alpha = \frac{\tau}{I}$ | The angular acceleration that an object experiences can be dervied from the torque acting on the object and its angular momentum. |
| $\tau = rF$ | A force $F$ applied tangentially at distance $r$ from the center of mass of an object induces a torque. |
| $L = I\omega$ | The angular momentum is proportional to the moment of inertia of the object and its angular velocity. |
| $v_t = \omega r$ | The tangential velocity of a point at $r$ distance from the center of mass depends on the angular velocity and the radius. |
| $a_t = \alpha r$ | The tangential acceleration of a point at $r$ distance from the center of mass depends on the angular acceleration and the radius. This is derived from the tangential velocity equation. |
| $F = \mu F_n$ | The maximum friction force depends on the normal force that is pressing the two objects together. |
| $F_d = \frac{1}{2} \rho v^2 C_d A$ | The drag force that an object experiences when moving through a fluid depends on the velocity, drag coefficient and surface area of the object, as well as on the density of the fluid. |
| $F_s = k x$ | The force that a spring exerts on an object is proportional to the displacement $x$ of the spring. The constant factor $k$ is specific to the spring and determines how stiff it is. This is according to [Hooke's law](https://en.wikipedia.org/wiki/Hooke%27s_law), though not all springs follow that law. |

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

## Principles

Following are some not so intuitive and common principles that for me personally have taken time to find and understand.

### Tangental velocity

The velocity that a point $p$ on an object experiences is equal to the sum of the velocity of the object and the angular velocity times the radius.

$$
v_t = v + \omega r
$$

### Tangental Acceleration

From the tangental velocity, we can derive the tangental acceleration of an offset point to be as follows:

$$
a_t = a + \alpha r
$$


### Offset torque

If a torque is applied to an object at a point away from the center of mass, the behavior is the same as through the torque was applied at the center of mass.

This is probably not so unknown nowadays, as drones demonstrate this principle very well - this is the mechanism through which they rotate about their vertical axis (yaw), even though the propellers are at an offset and are parallel to the horizontal plane.

### Offset force

If a force is applied to an object at a point away from the center of mass, both a force at the center of mass and a torque are applied to the object. What is interesting here is that the magnitute of the force is the same as would be if the force is applied at the center of mass.

That is, given an object with a center of mass $p_{cm}$ and a force $\vec{f}$ that is applied at point $p$, the resulting force and torque are applied.

$$
\vec{f_{cm}} = \vec{f}
$$

$$
\vec{\tau_{cm}} = \vec{f} \times (\vec{p - p_{cm}})
$$

We use the cross product to get the tangential component of the force, have it multiplied by the radius, and have the resulting torque perpendicular to radius and force, as is physically correct.

**NOTE:** Wrapping my head around the idea that a force applied at an offset affects the center of mass of the object in the same way as if it were applied at the center of mass was hard. If that were true, it felt that in the case where the force were applied at the center of mass there was a loss of energy since there was no rotational energy gain. I guess that one has to consider the motion of the object. An offset force will cause the object to spin faster and faster, leading to the motion being circular-like in shape and the object not gaining much potential energy. This is my personal way of thinking about this. Take this with a block of salt.

### Apparent mass

I was unable to find a correct term for this. The articles sometimes refer to it as _Effective Mass_, though it means a different thing in reality. The term _Reduced Mass_ is also a good candidate and it does cover part of the equation we will present here but in general relates to a different scenario. So I will be using _Apparent Mass_ here.

The apparent mass is represented with the following equation.

$$
m_{apparent} = \frac{1}{\frac{1}{m} + \frac{r^2}{I}}
$$

It represents the mass that an object appears to have when analyzed at some distant point. Another way of thinking about this is to ask what would the mass of the object "feel" to an observer if some distant (at $r$ distance) point of the object were to press against the observer.

While mostly a thought experiment and not really grounded in reality, this representation is really important when we start dealing with two objects that both exert an offset force on each other and we would like to find out what proportion of the net force needs to go to which object. We cannot use their full masses and we have to take the moment of inertia into consideration as well.

A detailed derivation of the above equation can be found on the [Apparent Mass Derivation](../explanations/apparent-mass-derivation.md) page.

### The Intertia Tensor

While in a 2D coordinate system it is sufficient to represent the moment of inertia as a scalar, since rotation can only occur about one axis (the one perpendicular to the screen), in 3D we need a more sophisticated representation that can handle various rotation vectors. This is where the Intertia Tensor comes into play.

The Inertia Tensor is a $3 \times 3$ matrix:

$$
I =
\begin{bmatrix}
I_{xx} & I_{xy} & I_{xz} \\
I_{yx} & I_{yy} & I_{yz} \\
I_{zx} & I_{zy} & I_{zz} \\
\end{bmatrix}
$$

The components can be calculated as follows:

$$
I_{xx} = \sum_i{m_i (y_i^2 + z_i^2)}
$$

$$
I_{yy} = \sum_i{m_i (x_i^2 + z_i^2)}
$$

$$
I_{zz} = \sum_i{m_i (x_i^2 + y_i^2)}
$$

$$
I_{xy} = I_{yx} = - \sum_i{m_i x_i y_i}
$$

$$
I_{xz} = I_{zx} = - \sum_i{m_i x_i z_i}
$$

$$
I_{yz} = I_{zy} = - \sum_i{m_i y_i z_i}
$$

Where $m_i$ represents the mass of an individual particle of the object and $x_i$, $y_i$ and $z_i$ represent the location of the particle relative to the center of mass.

One can use integration to solve the above equations for various shape types.

Having the Inertia Tensor, one can calculate the angular acceleration vector as follows:

$$
\vec{\alpha} = I^{-1} \vec{\tau} = I^{-1} (\vec{r} \times \vec{F})
$$

Check the [References](./references.md) page for links to resources with more detailed information.
