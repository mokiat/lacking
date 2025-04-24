func test() {

  alpha := texture(uv).a // get alpha

  // alpha check
  if (alpha < 0.5) {

      discard // drop this fragment
  }

  return
}
