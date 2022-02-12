#include "___.defs.h"
#include "os9_base.h"

void low__Read(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno) {
  *out_count = 0;
  *out_errno = 0;
  // Simulate EOF.
}
