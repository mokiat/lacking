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

**Note:** In some literature the constraint gradient is also called the Jacobian. Since the $C_v$ function maps from $R^n$ to $R^1$, the Jacobian and gradient are the same thing, except that the former is represented by a single-row matrix and the latter is represented by a vactor. This also means that using a Jacobian, one has to use matrix multiplication and using a gradient one has to use the vector dot product respectively. Since it is easier to write, we will use $J$ to represent the above gradient. Furthermore, unless otherwise specified, $J$ indicates $J(p)$ (the Jacobian at point $p$).

## Solver

In general, there are two main ways to solve constraints.

The first one is to solve the mathematic equations for all constraints in parallel and then apply a single impulse per object that produces the desired output. This approach is difficult in that the complexity rises drastically with each new constraint and there can be situations where an exact solution does not even exist (e.g. two constraints that require the object be positioned in two different spots).

The second one is to solve constraints for maximum two bodies at a time. Each constraint is solved an applied in turn. The whole process is repeated a number of times until the system hopefully reaches a "stable" state. In practice this is cheaper to run and produces good results. It is very similar to how Neural Networks are trained.

For more information on constraint solvers, make sure to check the [References](./references.md) page.
