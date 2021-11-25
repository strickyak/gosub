package builtin

type int int
type uint uint
type byte byte
type _type_ _type_

const nil = nil
const true bool = 1
const false bool = 0

type error interface {
	Error() string
}

func println(args ...interface{})
func make(t _type_, args ...int)
