package log

import "low"

func Printf(format string, args ...interface{}) {
	// TODO: Don't add trailing newline, if already there.
	low.FormatToBuffer("# "+format+"\n", args...)
	low.WriteBuffer(2)
}

func Fatalf(format string, args ...interface{}) {
	Printf("FATAL: "+format, args)
	low.Exit(13)
}
