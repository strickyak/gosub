#include "___.defs.h"

P_int builtin__len(P__any_ in) {
  switch (in.typecode[0]) {
    case 's':  // string
    {
      String* str = in.pointer;
      return (P_int)(str->len);
    } break;
    case 'S':  // Slice
    {
      Slice* slice = in.pointer;
      return (P_int)(slice->len);  // BUG: divide by element size.
    } break;
  }

  assert(0);
}
