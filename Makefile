T=demo/wc.go

all:
	bash ./gu build $T
	wc < demo/wc.go
	go run demo/wc.go < demo/wc.go
	./wc.bin < demo/wc.go

test: _FORCE_
	set -x; for x in test/t*.go ; do ./gu test $$x || { echo BROKEN: $$x; exit 63; } ; done
	echo ALL TESTS GOOD.

ci:
	set -x; ci-l runtime/*.c runtime/*.h Makefile *.go */*.go *.sh Makefile

fmt:
	gofmt -w *.go */*.go

clean:
	set -x ; rm -f *.s *.o *.bin a.out */*.go.c */*.want */*.got _ __ ___*

_FORCE_:
