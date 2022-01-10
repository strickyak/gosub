T=demo/wc.go

all:
	go run gosub.go  < $T   > $T.c
	clang-format -i --style=Google $T.c
	cat -n $T.c >&2
	cat $T.c | sed 's;//.*;;' | sed '/^ *$$/d' >&2
	cc -g -Iruntime $T.c runtime/runt.c runtime/bigmem.c
	./a.out

test: _FORCE_
	set -x; make T=test/t1.go >&2 && ./a.out > _ && diff -b test/t1.want _
	set -x; make T=test/t2.go >&2 && ./a.out > _ && diff -b test/t2.want _
	set -x; make T=test/t3.go >&2 && ./a.out > _ && diff -b test/t3.want _
	set -x; make T=test/t4.go >&2 && ./a.out > _ && diff -b test/t4.want _
	set -x; make T=test/t5.go >&2 && ./a.out > _ && diff -b test/t5.want _
	set -x; make T=test/t6.go >&2 && ./a.out > _ && diff -b test/t6.want _
	set -x; make T=test/t8.go >&2 && ./a.out > _ && diff -b test/t8.want _
	echo ALL TESTS GOOD.

ci:
	set -x; ci-l runtime/*.c runtime/*.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go

clean:
	rm *.s *.o a.out */*.go.c

_FORCE_:
