#include "___.defs.h"
#include "os9_base.h"

void low__Write(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno) {
  *out_errno = Os9Write(in_fd, (const char*)in_buf, in_size, out_count);
}
