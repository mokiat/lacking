# Equations

| Equation | Vector form | Notes |
| -------- | ----------- | ----- |
| $F = ma$ | $\vec{F} = M \vec{a}$ | The force is collinear to the acceleration and is proportional to the mass. |
| $p = mv$ | $\vec{p} = M \vec{v}$ | The momentum is collinear to the velocity and is proportional to the mass. |
| $\tau = I \alpha$ | $\vec{\tau} = I \vec{\alpha} + \vec{\omega} \times I \vec{\omega}$ | The vector form in 3D is the more accurate representation. It takes into account the fact that there can be a torque even without an angular acceleration, just because of the shape of the object. Check [Moment of Intertia](./moment-of-intertia.md) for more information. The torque might not be collinear with the angular acceleration. |
| $\tau = rF$ | $\vec{\tau} = \vec{r} \times \vec{F}$ | The cross product handles situations where the radius is not perpendicular to the force. |
| $L = I\omega$ | $\vec{L} = I \vec{\omega}$ | The angular momentum need not be collinear with the angular velocity. Check [Moment of Intertia](./moment-of-intertia.md) for more information. |
| $v_t = \omega r$ | $\vec{v_t} = \vec{\omega} \times \vec{r}$ | The cross product handles situations where the radius is not perpendicular to the angular velocity. The resulting tangential velocity, when not zero, is perpendicular to the angular velocity. |
| $a_t = \alpha r$ | $\vec{a_t} = \vec{\alpha} \times \vec{r}$ | The cross product handles situations where the radius is not perpendicular to the angular acceleration. |
| $F = \mu F_n$ | $\vec{F_{max}} = - \mu \hat{v_{t}} \|\vec{F_n}\|$ | This returns the maximum friction force. The actual force could be less if it would be sufficient to keep the object from moving. |
| $F_d = \frac{1}{2} \rho v^2 C_d A$ | $\vec{F_d} = \frac{1}{2} \rho C_d A \|\vec{v}\| \vec{v}$ | The velocity in this equation is the relative velocity of the wind to the object. |
| $F_l = \frac{1}{2} \rho v^2 C_L A$ | $\vec{F_l} = \frac{1}{2} \rho C_L A (\vec{v} \cdot \vec{v}) \hat{n}$ | The velocity in this equation is the relative velocity of the wind to the object. The $\hat{n}$ term defines the lift direction of the wing, perpendicular to the wind direction. |
| $F_s = k x$ | $\vec{F_s} = -k \vec{x}$ | This is according to [Hooke's law](https://en.wikipedia.org/wiki/Hooke%27s_law), though not all springs follow that law. The force is in the opposite direction to the displacement. |
