# Effective Mass

I was unable to find a correct term for this. The articles refer to it as _Effective Mass_, though it means a different thing in practice. Regardless, I will stick with the term _Effective Mass_ to keep this consistent with other tutorials out there.

The effective mass is represented with the following equation.

$$
m_{eff} = \frac{1}{m^{-1} + I^{-1}(\vec{r} \times \hat{n})^2}
$$

It represents the inertial mass of an object when interacted with at an offset (e.g. an airplane wing-mounted engine pushing on the airplane).

In practice we will be using the inverse effective mass which has the following equation.

$$
m_{eff}^{-1} = m^{-1} + I^{-1}(\vec{r} \times \hat{n})^2
$$

Lastly, if dealing with two objects, the inverse of the reduced effective mass of the system can be calculated as follows.

$$
m_{eff}^{-1} = m_1^{-1} + I_1^{-1}(\vec{r_1} \times \hat{n})^2 + m_2^{-1} + I_2^{-1}(\vec{r_2} \times \hat{n})^2
$$

A detailed derivation of the above equation can be found on the [Effective Mass Derivation](../derivations/effective-mass-derivation.md) page.
