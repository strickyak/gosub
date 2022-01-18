package main

import "errors"
import "log"

func main() {
	e := errors.New("I wish I were an Oscar Mayer Weiner")
	log.Fatalf("stopping with fatal %q", e, 1, 2, 3)
}
