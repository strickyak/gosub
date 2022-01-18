#include "runtime/runt.h"
#include "___.defs.h"

#ifdef USING_MODULE_unix
P_int unix__Close(P_int fd) {
  return close(fd);
}
#endif
