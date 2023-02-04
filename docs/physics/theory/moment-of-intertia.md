# Moment of Intertia

While mass represents the reluctance of an object to change its linear velocity, moment of inertia represents the reluctance of an object to change its angular momentum.
However, while having a lot of things in common, mass is much easier to reason about, whereas moment of inertia induces some strange properties on the motion of an object.
This is why this page is dedicated to discussing Moment of Inertia.

> NOTE: I have not studied physics past high school so more of the advanced stuff here is based on internet articles, tutorials, and personal reasoning. Take everything with a grain of salt.

The two main equations related to mass are as follows.

$$
\vec{p} = M \vec{v}
$$

and

$$
\vec{F} = M \vec{a}
$$

Where the second equation is derived by taking the derivative of the first one over time. In fact, it seems that Newton's second law of motion describes Force as the rate at which the momentum of an object changes with time.

For moment of inertia the equations are fairly similar.

$$
\vec{L} = I \vec{\omega}
$$

and

$$
\vec{\tau} = I \vec{\omega}
$$

Except that the second equation is actually wrong. While very common in a lot of text books, it works correctly only for 2D scenarios or in 3D scenarios where the rotation occurs over one of the principal axes (more on that later) of the object.

The correct equation is actually as follows.

$$
\vec{\tau} = I \vec{\alpha} + \vec{\omega} \times I \vec{\omega}
$$

In fact, similarly to the equation for force, this one is also derived by taking the derivative of the angular momentum equation (the first one) over time.

A key thing to note here is that unlike the force equation, where the mass is a scalar, in this equation the moment of inertia is actually a 3x3 matrix called a tensor. The reason for this is because the resistence to rotation differs depending on the angle of rotation. What's even more, the angular momentum or the torque (depending on which equation above is used) need not point in the same direction as the angular velocity or angular acceleration.

> NOTE: This last bit was hard for me to understand or create a mental image of. In the following text I will try to create a mostly intuitive explanation as to why the equation for torque is so complicated.

## The Intertia Tensor

We should first explore the Inertia Tensor $I$. As mentioned, it is a 3x3 matrix that is defined as follows.

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

Where $m_i$ represents the mass of an individual particle of the object and $x_i$, $y_i$ and $z_i$ represent the location of the particle relative to the point of rotation.

One can use integration to solve the above equations for various shape types.

Check the [References](../references.md) page for links to resources with more detailed information on how this is derived.

So what is the difference between a tensor and a 3x3 matrix. Well, to my understanding, a tensor describes a transformation under a change of coordinates. It can produce a scalar, a vector, or more complicated outputs.

In the case of the moment of inertia, an input vector is transformed to an output vector and the transformation is linear, which is a tensor of second order and is described by a symmetric matrix.

The components $I_{xx}$, $I_{yy}$, and $I_{zz}$ are called _the moments of inertia_, whereas the other components, $I_{xy}$, $I_{yx}$, $I_{xz}$, $I_{zx}$, $I_{yz}$, and $I_{zy}$, are called _the products of inertia_.

> My understanding is that for any object, if you position the point of rotation to be the center of mass (which would be the common case for a physics engine), you can find three orthogonal axes, such that _the products of inertia_ are zero. In such cases, the axes are called principal axes and rotation over them does not induce any torque (i.e. we can the more simple $\vec{\tau} = I \vec{\alpha}$ equation). The moment of inertia becomes a diagonal matrix.

## The torque equation

So now that we know how the moment of inertia is calculated and represented, let's get back to the intimidating torque equation.

$$
\vec{\tau} = I \vec{\alpha} + \vec{\omega} \times I \vec{\omega}
$$

### Velocity-induced torque

Let us consider the second part of the equation.

$$
\vec{\omega} \times I \vec{\omega}
$$

It tells us that a torque can be induced by the angular velocity alone, even if there is no angular acceleration.

> NOTE: This torque does not appear if an object is rotated about one of its principal axes.

TODO
