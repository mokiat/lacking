# Vectors

This page presents equations for working with vectors.

## Length preservation

$$
|c\vec{a}| = c|\vec{a}|
$$

## Dot Product

$$
\vec{a} \cdot \vec{b} = |a||b|\cos{\alpha}
$$

The dot product is very useful in determining if two vectors are perpendicular or whether they point in the same direction. A value of $0$ indicates that they are perpendicular. A positive value indicates that they point in the same half-space direction. A negataive value indicates that they point in opposite half-space directions.

Furthermore, if one of the vectors is unit (has length of one), it returns the length of the other vector's component that is along the first vector's direction. This can be used to measure the distance of a point to a plane if we have the normal of the plane and an arbitrary point on it.

## Cross Product

$$
\vec{a} \times \vec{b} = |\vec{a}||\vec{b}|\sin(\alpha)\vec{n}
$$

The cross product returns a new vector that is perpendicular to both of the initial vectors and has a length equal to the surface area that is bounded by the initial vectors. If the initial vectors are collinear, the resulting vector is the zero vector.

This is very useful in physics equations where the resulting concept is perpendicular to the two initial vectors (e.g. the torque vector is perpendicular to the force and the radius).

The order of the two vectors matters and flipping the order results in the inverse output vector.

$$
\vec{a} \times \vec{b} = - \vec{b} \times \vec{a}
$$

The cross product is distributable over addition.

$$
\vec{a} \times (\vec{b} + \vec{c}) = \vec{a} \times \vec{b} + \vec{a} \times \vec{c}
$$

## Triple Product

$$
\vec{a} \cdot (\vec{b} \times \vec{c})
=
\vec{b} \cdot (\vec{c} \times \vec{a})
=
\vec{c} \cdot (\vec{a} \times \vec{b})
$$

This gives the signed volume of the parallelepiped formed by the three vectors.

## Matrix - Vector multiplication

The standard formula for matrix to vector multiplication is as follows:

$$
\begin{bmatrix}
m_1 & m_2 & m_3 \\
m_4 & m_5 & m_6 \\
m_7 & m_8 & m_9 \\
\end{bmatrix}
\begin{bmatrix}
a_1 \\
a_2 \\
a_3 \\
\end{bmatrix}
=
\begin{bmatrix}
m_1a_1+m_2a_2+m_3a_3 \\
m_4a_1+m_5a_2+m_6a_3 \\
m_7a_1+m_8a_2+m_9a_3 \\
\end{bmatrix}
$$

This can be expressed in terms of vector dot products:

$$
\begin{bmatrix}
m_1 & m_2 & m_3 \\
m_4 & m_5 & m_6 \\
m_7 & m_8 & m_9 \\
\end{bmatrix}
\begin{bmatrix}
a_1 \\
a_2 \\
a_3 \\
\end{bmatrix}
=
\begin{bmatrix}
\vec{row_1} \\
\vec{row_2} \\
\vec{row_3} \\
\end{bmatrix}
\vec{a}
=
\begin{bmatrix}
\vec{row_1} \cdot \vec{a} \\
\vec{row_2} \cdot \vec{a} \\
\vec{row_3} \cdot \vec{a} \\
\end{bmatrix}
$$

Or it can also be expressed in terms of scaled vector sums:

$$
\begin{bmatrix}
m_1 & m_2 & m_3 \\
m_4 & m_5 & m_6 \\
m_7 & m_8 & m_9 \\
\end{bmatrix}
\begin{bmatrix}
a_1 \\
a_2 \\
a_3 \\
\end{bmatrix}
=
\begin{bmatrix}
m_1 \\
m_4 \\
m_7 \\
\end{bmatrix}
a_1
+
\begin{bmatrix}
m_2 \\
m_5 \\
m_8 \\
\end{bmatrix}
a_2
+
\begin{bmatrix}
m_3 \\
m_6 \\
m_9 \\
\end{bmatrix}
a_3
=
\vec{col_1}a_1
+
\vec{col_2}a_2
+
\vec{col_3}a_3
$$
