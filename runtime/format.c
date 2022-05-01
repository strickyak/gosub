#include "___.defs.h"

byte Guard0[16];
byte Buffer[256];
byte Guard1[16];
byte* BufferP;
byte* BufferEnd;

String low__BufferToString() {
  int len = BufferP - Buffer;
  assert(len >= 0);
  assert(len < INF);

  String z = {
        oalloc((byte)(len), C_Bytes), // base
        0, // offset
        len,
        };

    byte* src = Buffer;
    byte* dest = (byte*)z.base;
    for (int i = 0; i < len; i++) {
        *dest++ = *src++;
    }
    return z;
}

void PUTCHAR(byte x) {
  *BufferP = x;
  BufferP++;
  assert (BufferP < BufferEnd);
}

void PUTSTR(const char* s) {
  for (; *s; s++) {
    PUTCHAR(*s);
  }
}

void PUTSTRN(const char* s, byte n) {
  for (byte i=0; i<n; i++) {
    assert(s[i]);
    PUTCHAR(s[i]);
  }
}

void PUTDEC(P_byte x) {
  assert(x < 10);
  PUTCHAR('0' + x);
}

void PUTU(P_uintptr x) {
  if (x > 9) {
    PUTU(x / 10);
    PUTDEC((byte)(x % 10));
  } else {
    PUTDEC((byte)x);
  }
}
void PUTI(P_int x) {
  if (x<0) {
    PUTCHAR('-');
    PUTU(-x);
  } else {
    PUTU(x);
  }
}

void PUTHEX(byte x) {
  assert(x < 16);
  if (x < 10) {
    PUTDEC(x);
  } else {
    PUTCHAR('a' + x - 10);
  }
}

void PutX(P_uint x) {
  if (x > 15) {
    PutX(x >> 4);
    PUTHEX((byte)(x & 15));
  } else {
    PUTHEX((byte)x);
  }
}

void PutCurly(byte c) {
          PUTCHAR('{');
          PUTU((word)c);
          PUTCHAR('}');
}

void FormatQ(byte* str, int n) {
  PUTCHAR('"');
  for (int i = 0; i < n; i++) {
    byte c = str[i];
    if (32 <= c && c <= 127) {
      switch (c) {
        case '"':
        case '\'':
        case '\\':
        case '{':
        case '}':
          PutCurly(c);
            break;
        default:
          PUTCHAR(c);
      }
    } else {
          PutCurly(c);
    }
  }
  PUTCHAR('"');
}

int low__FormatToBuffer(String s, Slice args) {
  BufferP = Buffer;
  BufferEnd = Buffer + sizeof(Buffer);

  P__any_* a = (P__any_*)(args.base + args.offset);
  P__any_* a_end = a + (args.len / sizeof(*a));

  byte* p = (byte*)(s.base + s.offset);
  byte* p_end = p + s.len;

  for (; p < p_end; p++) {
    byte c = *p;
    if (c == 0/*EOS*/) goto RETURN;
    if (c=='%') {
        ++p;
        c = *p;

        if (a >= a_end) {
          PUTSTR("<end>");
        } else {

          if (c == 0/*EOS*/) goto RETURN;

          switch (a->typecode[0]) {
            case 's': // case string
            case 'S': // case Slice (actually for slice of bytes) pun as String:
              {
                String* xp = (String*)a->pointer;
                if (c=='q')
                  FormatQ((byte*)(xp->base + xp->offset), xp->len);
                else
                  PUTSTRN((const char*)(xp->base + xp->offset), xp->len);
              }
              break;
            case 'z': // case bool
              PUTSTR( (*(P_bool*)a->pointer) ? "true" : "false");
              break;
            case 'b':  // case byte
              PUTU(*(P_byte*)a->pointer);
              break;
            case 'i':  // case int
              PUTI(*(P_int*)a->pointer);
              break;
            case 'u': // case uint
              PUTU(*(P_uint*)a->pointer);
              break;
            case 'p': // case pointer
              PUTSTR( "(*)" );
              PUTU((P_uintptr)*(void**)a->pointer);
              break;
            default: // default: Unhandled Type
              PUTSTR("(percent "); PUTU(c);
              PUTSTR(" typecode "); PUTSTR(a->typecode);
              PUTSTR(" pointer "); PUTU((P_uintptr)a->pointer);
              PUTSTR(" * "); PUTU(((P_uintptr*)a->pointer)[0]);
              PUTSTR(" * "); PUTU(((P_uintptr*)a->pointer)[1]);
              PUTSTR(")");
              break;
          }
      }
      a++;
    } else {
      // Not a % escape -- just a normal char
      PUTCHAR(*p);
    }
  }  // next byte *p
RETURN:
  return BufferP - Buffer;
}  // end low__FormatToBuffer
