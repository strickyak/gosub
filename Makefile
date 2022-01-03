T=demo/wc.go
all:
	go run gosub.go  < $T   > $T.c
	clang-format -i --style=Google $T.c
	cat -n $T.c
	# cc -I. $T.c runt.c bigmem.c
	# ./a.out

ci:
	ci-l runtime/*.c runtime/*.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go

clean:
	rm *.s *.o a.out
