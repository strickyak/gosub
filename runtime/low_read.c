#include "___.defs.h"

void low__Read(P_int fd, P_uintptr buf, P_int size, P_int* count_out,
                P_int* errno_out) {
  int cc = read(fd, (char*)buf, size);
  *count_out = cc;
  *errno_out = errno;
}
