package parser

/*

Simplifications to Go:

No shadowing of identifiers in nested scopes.
Cannot reuse builtin names.  Cannot reuse global names.

`const` and `type` are only allowed at outer level.

`const` are untyped bool, integer, or string.

`type` is only used to define struct and interface.
Structs can only be defined by `type` at the global level;
there are no anonymous structs.  Similarly, interfaces can only
be defined by `type` at the global level, with the exception
of `interface {}`.   Thus with the exception of `interface {}`,
all structs and interfaces have a global type name.

`struct` is never used as a bare type; only pointer-to-struct
can be used.

All structs are allocated in the GC Heap, so pointer_to_struct
is a garbage-collecting pointer to GC Heap.  The term `handle`
will be used for these pointers.

Pointers are never used except for handles to structs.

Structs cannot be embedded -- they must be the toplevel thing
in a GC Heap object.

Methods are defined only on *struct, never on struct,
never on non-struct (since you can't define non-struct types
with `type`).

The GC Heap contains two hidden values for each allocation: its length
and its "class".  The length can be greater than the actual ask, so
that lengths can come from a small set of sizes (and there can be a free
list for each size).  The "class" is a byte that identifies the type of
the object.  Class is required in order for Garbage Collection to know
where handles are inside the object, but can also be used by languages
for dynamic dispatch of methods.  This subset of Go will use "class"
to do dynamic dispatch of methods on interfaces.

There are some special "internal classes" that are variable-lengthed.
These can be used for storing bytes in a string, handles in a slice,
or some more complicated cases (like a slice of slices).

Type names are easy to resolve because they are all known at the top
level.

Lambdas can only be defined at the first level in a func, and they cannot
be used after their parent scope is gone -- that is, they reference
variables directly on the C stack.

`defer` can only be used at the first level in a func.  So there is no
problem using `defer` to recover as the first thing in a function, or
using `defer` to close files that are opened in the first level of a func.

Slices are triples {handle, offset, length} as in normal Go.  The handle
is to GC Heap object of an internal struct type.

Strings are like slices, triples {handle, offset, length}.  To make
literal strings cheaper, we may allow the handle to be nil, and the
offset to locate a literal C string in a readonly OS9 module.

Maps are simple handles to a GC Heap object of an internal
struct type.

Interfaces can be implemented as two cases:
(1) `interface {}`
(2) interfaces that point to structs.
This is because only structs and interfaces can be defined with `type`.  
Case 1 will have to be big enough to hold a slice or string triple
and a type pointer.  Case 2 only needs to hold a handle, since
we can ask the class of the Heap object referenced by the handle.

Channels are not supported (yet).
When they are supported, they will be a simple handle to a GC Heap object
of an internal struct type.

Goroutines are not supported (yet).
When they are supported, they will be global function calls
or method invocations, not lambdas (since lambdas do not
survive their scope).  They will have to have their own C stack.

No renaming imports.  No "/" in import names (use a flat space).
No groups with ( and ) for imports, const, var, or type.
No types for enums.  No definitions with iota.

*/
