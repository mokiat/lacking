---
title: Overview
---

# Physics

The Lacking framework features an impulse-based physics engine. Impulse-based physics engines use velocity corrections to adjust the state of objects and to correct their trajectories according to physical constraints. In addition, the Lacking physics engine features a nudge system that operates directly on positions.

## Integration

Integration is one of the fundamental concepts of physics engines. We need to figure out how to move an object based on the forces that act on it and the current velocity (both directional or rotational) it has.

For simple scenarios, like constant linear acceleration, this has often been done through equations like the following:

$$
\Delta p = v_0 \Delta t + \frac{1}{2}a{\Delta t}^2
$$

While this equation works, it assumes that a single force is applied on the object and the acceleration does not change throughout $\Delta t$. In reality, there can be multiple forces acting on an object that vary depending on the object's position or rotation. For some of those it might be easy to figure out the equation but for others (e.g. [The Three-Body Problem](https://en.wikipedia.org/wiki/Three-body_problem)) there is no [closed-form solution](https://en.wikipedia.org/wiki/Closed-form_expression).

In such cases, we use a simplified sequence of operations that do not produce an exact solution but compensate with flexibility and performance. After it is applied, corrections are performed to ensure that a physics constraint is not violated (e.g. an object should not go through another object).

There are a number of famous integration methods out there. Lacking uses the [semi-implicit Euler](https://en.wikipedia.org/wiki/Semi-implicit_Euler_method) integration method.

$$
v_1 = v_0 + a \Delta t
$$

$$
s_1 = s_0 + v_1 \Delta t
$$

More broadly, it performs the following sequence of steps.

1. Applies forces to all dynamic objects.
1. Derives the new velocities of all dynamic objects based on the accumulated accelerations.
1. Applies correction impulses to all dynamic objects that have constraints on them.
1. Derives the new positions of all dynamic objects based on the evaluated velocitiess
1. Applies correction nudges to all dynamic objects that have constraints on them.
1. Detects collisions and creates temporary collision constraints.

For more information on integration, make sure to check the [References](./references.md) page.

## Impulses

Outside the standard integration, when adjustments are needed to an object's velocity, the engine uses impulses.

Impulses are like forces, except that they deal with the velocity instead of the acceleration.

$$
P = \Delta{t}F = vm
$$

And just as forces applied at an offset to an object induce both a change in linear and angular accelerations, impulses induce both a change in linear and angular velocities.

$$
\Delta{\vec{v}} = m^{-1} \vec{P}
$$

$$
\Delta{\vec{w}} = I^{-1}(\vec{r} \times \vec{P})
$$

## Constraints

Constraints are a mechanism to enforce a physics rule or restriction on an object. Examples include having an object always point towards a point in space, preventing an object from falling through the ground, restricting the motion of an object to a single axis, etc.

The way constraints are expressed mathematically is through equations that equal zero when the constraint is satisfied.

$$
C = 0
$$

For example, the following constraint requires that an object has a position $p$ a specific distance $l$ away from a point $p_0$.

$$
C_p(p) = |\vec{p} - \vec{p_0}| - l
$$

The above equation is equal to $0$ only when the object is $l$ distance away from $p_0$.

As mentioned before, an impulse engine uses velocity adjustments (impulses) to enforce constraints. As such, we require constraint functions that take the object's velocity as an argument. We achieve this by using the positional constraint and differentiation over time. For the above positional constraint, we get the following velocity constraint.

$$
C_v(v) = C_p' = \frac{\vec{p} - \vec{p_0}}{|\vec{p} - \vec{p_0}|} \cdot \vec{v}
$$

Our next step is to get the gradient of the constraint. This allows us to use gradient descent to make the optimal velocity adjustments. We do this through differentiation over the velocity.

$$
\nabla{C_v}(v) = \frac{\vec{p} - \vec{p_0}}{|\vec{p} - \vec{p_0}|}
$$

**Note:** In some literature the constraint gradient is also called the Jacobian. Since the $C_v$ function maps from $R^n$ to $R^1$, the Jacobian and gradient are the same thing, except that the former is represented by a single-row matrix and the latter is represented by a vactor. This also means that using a Jacobian, one has to use matrix multiplication and using a gradient one has to use the vector dot product respectively. Since it is easier to write, we will use $J$ to represent the above gradient. Furthermore, unless otherwise specified, $J$ indicates $J(v)$ (the Jacobian at velocity $v$).

Once we have the Jacobian, we can use the direction it implies to apply an impulse on the object.

$$
\vec{P} = J^T \lambda
$$

Note: we transpose the jacobian to convert it from a $1 \times 3$ (when working in 3D) matrix, to a 3D vector.

While the Jacobian $J$ (or rather the inverse) determines the direction, the $\lambda$ scalar determines the strength of the impulse. Where $\lambda$ is calculated as follows.

$$
\lambda = - \frac{J\vec{v_0}}{JM^{-1}J^T}
$$

Here $\vec{v_0}$ is the current velocity of the object and $M^{-1}$ is the inverse mass matrix, though $\frac{1}{m}$ works just as well in the general case. It is also the case that $JM^{-1}J^T$ produces the [inverse effective mass](./theory.md#effective-mass).

This brings the equation down to:

$$
\vec{P} = - J^T \frac{J\vec{v_0}}{JM^{-1}J^T}
$$

In practice, we often have an offset impulse, in which case we need to take the moment of inertia and the current angular velocity into account. The equation is pretty much the same, except that the velocity vector now includes the angular components as well and the mass matrix includes the moment of inertia.

$$
\vec{v_0} =
\begin{bmatrix}
v_x \\
v_y \\
v_z \\
w_x \\
w_y \\
w_z \\
\end{bmatrix}
$$

$$
M =
\begin{bmatrix}
m & 0 & 0 & 0 & 0 & 0 \\
0 & m & 0 & 0 & 0 & 0 \\
0 & 0 & m & 0 & 0 & 0 \\
0 & 0 & 0 & I_{xx} & I_{xy} & I_{xz} \\
0 & 0 & 0 & I_{yx} & I_{yy} & I_{yz} \\
0 & 0 & 0 & I_{zx} & I_{zy} & I_{zz} \\
\end{bmatrix}
$$

**Note:** Don't forget that the **inverse** of the matrix $M$ is used in the equations above.

More information on how the above equation was derived can be found on the [Impulse Derivation](./derivations/impulse-derivation.md) page.

## Solver

In general, there are two main ways to solve constraints.

The first one is to solve the mathematic equations for all constraints in parallel and then apply a single impulse per object that produces the desired output. This approach is difficult in that the complexity rises drastically with each new constraint and there can be situations where an exact solution does not even exist (e.g. two constraints that require the object be positioned in two different spots).

The second one is to solve constraints for maximum two bodies at a time. Each constraint is solved an applied in turn. The whole process is repeated a number of times until the system hopefully reaches a "stable" state. In practice this is cheaper to run and produces good results. It is very similar to how Neural Networks are trained.

For more information on constraint solvers, make sure to check the [References](./references.md) page.
