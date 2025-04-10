func test() {
  if 10 > 5 {
    doFirst()
  }
  if 10 > 20 {
    doFirst()
  } else {
    doSecond()
  }
  if 10 > 20 {
    doFirst()
  } else if 10 > 5 {
    doSecond()
  } else {
    doThird()
  }
}