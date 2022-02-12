#include "___.defs.h"
#include "os9_base.h"

// Borrow os9 system calls from NCL.
typedef unsigned int uint;
#define OMIT_stkcheck
#define OMIT_exit

// #include "../../doing_os9/picol/os9.c"
// #include "../../doing_os9/picol/puthex.c"

#include "picol/os9.c"
#include "picol/puthex.c"
