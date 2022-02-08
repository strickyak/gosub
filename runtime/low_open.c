#include "___.defs.h"

void low__Open(P_string filename, P_uint flags, P_uint mode, P_int* fd_out,
                P_int* errno_out) {
  const char* s = STRING_START(filename);
  int fd = open(s, flags, mode);
  *fd_out = fd;
  *errno_out = 0;
  if (fd < 0) {
    *errno_out = errno;
  }
}
