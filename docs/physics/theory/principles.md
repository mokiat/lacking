# Principles

Following are some physics principles that are useful to keep in mind.

## Tangental velocity

The velocity that a point $p$ on an object experiences is equal to the sum of the velocity of the object and the angular velocity times the radius.

$$
v_t = v + \omega r
$$

## Tangental Acceleration

From the tangental velocity, we can derive the tangental acceleration of an offset point to be as follows:

$$
a_t = a + \alpha r
$$

## Offset torque

If a torque is applied to an object at a point away from the center of mass, the behavior is the same as through the torque was applied at the center of mass.

This is probably not so unknown nowadays, as drones demonstrate this principle very well - this is the mechanism through which they rotate about their vertical axis (yaw), even though the propellers are at an offset and are parallel to the horizontal plane.

## Offset force

If a force is applied to an object at a point away from the center of mass, both a force at the center of mass and a torque are applied to the object. What is interesting here is that the magnitute of the force is the same as would be if the force were applied at the center of mass.

That is, given an object with a center of mass $\vec{p_{cm}}$ and a force $\vec{F}$ that is applied at point $\vec{p}$, the resulting force and torque arise.

$$
\vec{F_{cm}} = \vec{F}
$$

$$
\vec{\tau_{cm}} = (\vec{p} - \vec{p_{cm}}) \times \vec{F}
$$


> **NOTE:** Wrapping my head around the idea that a force applied at an offset affects the center of mass of the object in the same way as if it were applied at the center of mass was hard. If that were true, it felt that in the case where the force were applied at the center of mass there was a loss of energy since there was no rotational energy gain. I guess that one has to consider the motion of the object. An offset force will cause the object to spin faster and faster, leading to the motion being circular-like in shape and the object not gaining much potential energy. This is my personal way of thinking about this. Take this with a block of salt.
