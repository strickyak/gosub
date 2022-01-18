#include "___.defs.h"
#include "runtime/runt.h"

#ifdef USING_MODULE_unix
P_int unix__Close(P_int fd) { return close(fd); }
#endif
