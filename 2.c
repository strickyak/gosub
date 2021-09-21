#include <stdio.h>
#include "runtime_c.h"
t_int2  F_main__Hyp(t_int2 v_a, t_int2 v_b);

void F_main__main();

// package main
// ..... Imports .....
// ..... Consts .....
// ..... Types .....
// ..... Vars .....
// ..... Funcs .....
t_int2  F_main__Hyp(t_int2 v_a, t_int2 v_b) {

  return ((v_a) * (v_a)) + ((v_b) * (v_b));
}

void F_main__main() {

  t_int2 v_a = (t_int2 )(3);
  t_int2 v_b = (t_int2 )(4);
  t_int2 v_c = (t_int2 )((F_main__Hyp(v_a, v_b)));
  (void)((F_BUILTIN_println(v_c)));
}

// ..... Done .....
