package builtin

type error interface {
	Error() string
}

func println(args ...interface{})
func make(t _type_, args ...int) interface{} // not really.
func len(coll interface{}) int
func panic(arg interface{})
