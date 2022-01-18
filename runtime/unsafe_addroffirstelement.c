// #include "runtime/runt.h"
#include "___.defs.h"

#ifdef USING_MODULE_unsafe

P_uintptr unsafe__AddrOfFirstElement(P__any_ in) {
  switch (in.typecode[0]) {
    case 's':  // string
    {
      String* str = in.pointer;
      return (P_uintptr)(str->base + str->offset);
    } break;
    case 'S':  // Slice
    {
      Slice* slice = in.pointer;
      return (P_uintptr)(slice->base + slice->offset);
    } break;
  }

  assert(0);
}
#endif
