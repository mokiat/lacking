#if defined(GL_FRAGMENT_PRECISION_HIGH)
precision highp float;
precision highp sampler2DShadow;
precision highp sampler2DArrayShadow;
#else
precision mediump float;
precision mediump sampler2DShadow;
precision mediump sampler2DArrayShadow;
#endif