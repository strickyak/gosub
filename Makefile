all:
	go run gosub.go  < 2.g   > 2.c 
	cat -n 2.c
	cc 2.c runtime_c.c
	./a.out

ci:
	ci-l *.c *.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go
