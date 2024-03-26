# Derivatives

The following is a table of first order derivatives.

| Description | Derivative | Result |
| ----------- | ---------- | ------ |
| Constant | $c'$ | $0$ |
| Variable | $x'$  | $1$ |
| Scaled variable | $(cx)'$ | $cx'$ |
| Sum | $(x+y)'$ | $x' + y'$ |
| Product | $(xy)'$ | $xy'+x'y$ |
| Quot | $\frac{x}{y}$ | $\frac{x.y'+x'.y}{y^2}$ |
| Reciprocal | $\frac{1}{x}$ | $\frac{-x'}{x^2}$ |
| Power | $(x^y)'$ | $yx^{y-1}$ |
| Square root | $\sqrt{x}'$ | $\frac{1}{2\sqrt{x}}$ |
| Chained rule | $\frac{\partial f(g(x))}{\partial x}$ | $\frac{\partial f(g(x))}{\partial g(x)} \frac{\partial g(x)}{\partial x} $ |
| Multivariable rule | $\frac{\partial f(u(x), v(x))}{\partial x}$ | $\frac{\partial f(u(x), v(x))}{\partial u(x)}\frac{\partial f(u(x), v(x))}{\partial u(x)} + \frac{\partial f(u(x), v(x))}{\partial x}$ |
| Vector square length | $(\vec{v}.\vec{v})'$ | $2\vec{v}\vec{v}'$ |
| Vector length | $\|\vec{v}\|'$ | $\hat{v}\vec{v}'$ |
