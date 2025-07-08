package shape3d

/*
DEVELOPMENT NOTES:

Passing spheres by ref to CheckIntersectionSphereWithSphere does not yield
faster performance.

Collecting intersections is faster than returning them as an optional. This is
because the Intersection object is large and in case of non-intersection
(the more common case), a duffcopy need not be performed.

However, using generics to pass the collector instead of an interface does not
yield any performance benefits.

Using 32bit floats for the spheres does make checking of intersections almost
twice faster but only in the API where an opt.T is returned, again, due to
the duffcopy problem. However, the in the collection case there is only a slight
performance boost.

EDIT: Using (Intersection, bool) is actually faster than opt.T and is very close
to the 32bit approach.

---

Checking if two spheres intersect is slightly faster if square distance is used
but not substantial. It's unclear to me which option reduces the chance of
floating point error.

*/
