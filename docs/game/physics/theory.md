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
| $\alpha = \frac{\tau}{I}$ | The angular acceleration that an object experiences can be dervied from the torque acting on the object and its angular momentum. |
| $\tau = rF$ | A force $F$ applied tangentially at distance $r$ from the center of mass of an object induces a torque. |
| $L = I\omega$ | The angular momentum is proportional to the moment of inertia of the object and its angular velocity. |
| $v_t = \omega r$ | The tangential velocity of a point at $r$ distance from the center of mass depends on the angular velocity and the radius. |
| $a_t = \alpha r$ | The tangential acceleration of a point at $r$ distance from the center of mass depends on the angular acceleration and the radius. This is derived from the tangential velocity equation. |

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

I was unable to find a correct term for this. The articles sometimes refer to it as _Effective Mass_, though it means a different thing in reality. The term _Reduced Mass_ is also a good candidate and it does cover part of the equation we will present here but in general covers a different scenario. So I will be using _Apparent Mass_ here.

Let us consider an object that is experiencing both an acceleration and an angular acceleration (e.g. a spaceship that has only one of its engines working). Now, suppose that the object is attached at some offset point to an anchor that cannot be moved. We would like to know what force would the anchor experience.

The equation is as follows (in non-vector form):

$$
F_{anchor} = (a + \alpha r)\frac{1}{\frac{1}{m} + \frac{r^2}{I}}
$$

We can check what happens with the equation if the anchor was exactly at the center of mass ($r = 0$).

$$
F_{anchor} = (a + \alpha 0)\frac{1}{\frac{1}{m} + \frac{0^2}{I}}
= a\frac{1}{\frac{1}{m}} = am
$$

The force that the anchor would experience is equal to the force that is causing the object to accelerate. That is, the anchor would need to handle the full force, which makes sense.

So where does the term _Effective mass_ come into play. Well, if we were to take the above equation but were to ignore the fact that we have a whole object but rather wanted to simplify it to a point mass that was accelerating around some axis at distance $r$ but was also accelerating linearly, we would get the following.

$$
(a + \alpha r) m_{point} = (a + \alpha r)\frac{1}{\frac{1}{m} + \frac{r^2}{I}}
$$

$$
\Downarrow
$$

$$
m_{point} = m_{apparent} = \frac{1}{\frac{1}{m} + \frac{r^2}{I}}
$$

While mostly a thought experiment and not really grounded in reality, this representation is really important when we start dealing with two objects that both exert an offset force on each other and we would like to find out what proportion of the net force needs to go to which object. We cannot use their full mass and we have to take the moment of inertia into consideration as well. This is where the apparent mass comes into play. We can just sum the two apparent masses and continue from there.

So how did we get to this equation. We should consider the point $p$ on the object that is attached to the anchor. Since the anchor does not budge, we expect that point $p$ has to be stationary as well. Let's look at what accelerations point $p$ experiences.

The first acceleration it experiences is the one from the object's acceleration and angular acceleration.

$$
a_{object} = a + \alpha r
$$

The equation above is the one we already looked at in the [tangental acceleration](#tangental-acceleration) section.

The second acceleration it experiences is the one from the anchor, which is trying to resist the point's motion.

$$
a_{anchor} = -\frac{F_{anchor}}{m} - \frac{F_{anchor}r^2}{I}
$$

NOTE: The sign is negative, as the anchor acceleration acts in the opposite direction.

This one is a bit more complicated. Let's look at how we arived at the two terms. The first one represents the acceleration that the whole object would experience because of $F_{anchor}$ and the second one is the tangential acceleration.

Recall from [offset force](#offset-force) that a force applied to an object induces a force on the center of mass and a torque.

Linear acceleration:

$$
F_{cm} = -F_{anchor}
$$

$$
\Downarrow
$$

$$
a_{cm} = -\frac{F_{cm}}{m} = -\frac{F_{anchor}}{m}
$$

Angular acceleration:

$$
\tau_{cm} = -F_{anchor} r
$$

$$
\Downarrow
$$

$$
\frac{\tau_{cm}}{I} = -\frac{F_{anchor} r}{I}
$$

$$
\Downarrow
$$

$$
\alpha = -\frac{F_{anchor} r}{I}
$$

$$
\Downarrow
$$

$$
\alpha r = -\frac{F_{anchor} r r}{I} = -\frac{F_{anchor} r^2}{I}
$$

$$
\Downarrow
$$

$$
a_{tangent} = -\frac{F_{anchor} r r}{I} = -\frac{F_{anchor} r^2}{I}
$$

Hence:

$$
a_{anchor} = a_{cm} + a_{tangent} = - \frac{F_{anchor}}{m} - \frac{F_{anchor}r^2}{I}
$$

Now, we want to get back to the two original equations. If the point $p$ is to be stationary, we want the acceleration induced by the object to be negated by the acceleration induced by the anchor force.

$$
a_{object} = - a_{anchor}
$$

$$
\Downarrow
$$

$$
a + \alpha r = \frac{F_{anchor}}{m} + \frac{F_{anchor}r^2}{I}
$$

$$
\Downarrow
$$

$$
F_{anchor} (\frac{1}{m} + \frac{r^2}{I}) = (a + \alpha r)
$$

$$
\Downarrow
$$

$$
F_{anchor} = (a + \alpha r) \frac{1}{\frac{1}{m} + \frac{r^2}{I}}
$$

And so we have arrived at our initial equation.

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
