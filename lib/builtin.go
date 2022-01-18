package builtin

type error interface {
	Error() string
}

func println(args ...interface{}) // not really.

func make(t _type_, args ...int) interface{} // not really.

func len(coll interface{}) int
