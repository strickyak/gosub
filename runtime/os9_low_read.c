#include "___.defs.h"
#include "os9_base.h"

void low__Read(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno) {

  // int Os9Read(
  //     int path, char* buf, int buflen,
  //     int* bytes_read)

  *out_errno = Os9Read(in_fd, in_buf, in_size,
      out_count);
}
