T=demo/wc.go

all:
	go run gosub.go  < $T   > $T.c
	clang-format -i --style=Google $T.c
	cat -n $T.c >&2
	cat $T.c | sed 's;//.*;;' | sed '/^ *$$/d' >&2
	cc -g -Iruntime $T.c runtime/runt.c runtime/bigmem.c
	./a.out

test: _FORCE_
	set -x; for x in test/t*.go ; do ./gu test $$x ; done
	echo ALL TESTS GOOD.

ci:
	set -x; ci-l runtime/*.c runtime/*.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go

clean:
	set -x ; rm -f *.s *.o a.out */*.go.c */*.want */*.got _ __ _[0-9]*

_FORCE_:
