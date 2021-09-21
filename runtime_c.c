#include <stdio.h>
#include "runtime_c.h"

extern void F_main__main();

void F_BUILTIN_println(int i) {
    fprintf(stderr, " %d\n", i);
}

int main(int argc, const char* argv[]) {
  F_main__main();
  return 0;
}
