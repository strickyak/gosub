package unsafe

// Address for first element, of string or slice.
// This is equivalent to Base + Offset.
func AddrOfFirstElement(stringOrSlice interface{}) uintptr

// These can be used before & after calling OS9 system calls with a string.
func SetHighBitOfFinalChar(s string) byte

// Call this with the byte returned by SetHighBitOfFinalChar to restore the string.
func RestoreFinalChar(s string, b byte)

// Raw memory access.
func Peek(addr uintptr) byte
func Poke(addr uintptr, value byte)
func Peek2(addr uintptr) uint
func Poke2(addr uintptr, value uint)

// Find the address of a variable.
// Be careful, it may be tricky -- the compiler might make unwanted copies of variables.
func AddressOf(thing interface{}) uintptr
