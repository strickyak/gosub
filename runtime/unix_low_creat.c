#include "___.defs.h"

void low__Creat(P_string filename, P_uint mode, P_int* fd_out,
                 P_int* errno_out) {
  const char* s = STRING_START(filename);
  int fd = creat(s, mode);
  *fd_out = fd;
  *errno_out = 0;
  if (fd < 0) {
    *errno_out = errno;
  }
}
