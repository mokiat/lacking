# Physics

The Lacking framework features an impulse-based physics engine. Impulse-based physics engines use velocity corrections to adjust the state of objects and to correct their trajectories according to some physical constraints. In addition, the Lacking physics engine features a nudge system that operates directly on positions.

## Integration

Integration is one of the fundamental concepts of physics engines. We need to figure out how to move an object based on the forces that act on it and any existing velocity (both directional or rotational) it has.

For simple scenarios, like constant linear acceleration, this has often been done through equations like the following:

$$
\Delta s = v_0 \Delta t + \frac{1}{2}a{\Delta t}^2
$$

While this equation works, it assumes that a single force is applied on the object and the acceleration does not change throughout $\Delta t$. In reality, there can be multiple forces acting on an object that vary depending on the object's position. For some of those it might be easy to figure out the equation but for others (e.g. [The Three-Body Problem](https://en.wikipedia.org/wiki/Three-body_problem)) there is no [closed-form solution](https://en.wikipedia.org/wiki/Closed-form_expression).

In such cases we use a simplified sequence of equations that produce a somewhat correct result but compensate for flexibility and performance. After it is applied, corrections are performed to ensure that a physics constraint is not violated (e.g. an object should not go through another object).

There are a number of famous integration methods out there. Lacking uses the [semi-implicit Euler](https://en.wikipedia.org/wiki/Semi-implicit_Euler_method) integration method.

$$
v = v_0 + a \Delta t
\\
s = s_0 + v \Delta t
$$

More broadly, it performs the following sequence of steps:

1. Apply forces to all dynamic objects
1. Derive the new velocities of all dynamic objects based on the accumulated accelerations
1. Apply correction impulses to all dynamic objects that have constraints on them
1. Derive the new positions of all dynamic objects based on the evaluated velocities
1. Apply correction nudges to all dynamic objects that have constraints on them
1. Detect collisions and create temporary collision constraints

The correction impulses and the correction nudges are applied a number of times. This is due to the usage of local solvers instead of global ones, where the latter is much more expensive to achieve. This will be covered in the constraint section.
