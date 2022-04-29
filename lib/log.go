package log

import "low"

func Printf(format string, args ...interface{}) {
	// if len(format) > 0 && format[len(format)-1] != '\n' {
	// format = format + "\n"
	// }
	low.FormatToBuffer("# "+format, args...)
	low.WriteBuffer(2)
	low.FormatToBuffer("\n")
	low.WriteBuffer(2)
}

func Fatalf(format string, args ...interface{}) {
	Printf("FATAL: "+format, args)
	low.Exit(13)
}
