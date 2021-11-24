package builtin

type error interface {
	Error() string
}

func println(args ...interface{})
func make(t _type_, args ...int)

const true bool = 1
const false bool = 0
