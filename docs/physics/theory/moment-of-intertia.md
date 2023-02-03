# Moment of Intertia

TODO

## The Intertia Tensor

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

Check the [References](../references.md) page for links to resources with more detailed information.
