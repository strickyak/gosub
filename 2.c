#include <stdio.h>
#include "runt.h"
t_int2  F_main__Hyp(t_int2 v_a, t_int2 v_b);

void F_main__main();

t_int2  F_main__Triangle(t_int2 v_n);

// package main
// ..... Imports .....
// ..... Consts .....
// ..... Types .....
// ..... Vars .....
// ..... Funcs .....
t_int2  F_main__Triangle(t_int2 v_n) {

  t_int2 v_sum = (t_int2 )(0);
  while(1) {
    t_bool _while_ = (t_bool)((v_n) > (0));
    if (!_while_) break;
  v_sum = (t_int2 )((v_sum) + (v_n));
  v_n = (t_int2 )((v_n) - (1));
  }
  return v_sum;
}

t_int2  F_main__Hyp(t_int2 v_a, t_int2 v_b) {

  return ((v_a) * (v_a)) + ((v_b) * (v_b));
}

void F_main__main() {

  t_int2 v_a = (t_int2 )(3);
  t_int2 v_b = (t_int2 )(4);
  t_int2 v_c = (t_int2 )((F_main__Hyp(v_a, v_b)));
  (void)((F_BUILTIN_println(v_c)));
  (void)((F_BUILTIN_println((F_main__Triangle(10)))));
}

// ..... Done .....
