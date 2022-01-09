T=demo/wc.go
all:
	go run gosub.go  < $T   > $T.c
	clang-format -i --style=Google $T.c
	cat -n $T.c >&2
	cat $T.c | sed 's;//.*;;' | sed '/^ *$$/d' >&2
	cc -g -Iruntime $T.c runtime/runt.c runtime/bigmem.c
	./a.out

ci:
	set -x; ci-l runtime/*.c runtime/*.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go

clean:
	rm *.s *.o a.out
