DEMO=2.g
all:
	go run gosub.go  < $(DEMO)   > $(DEMO).c 
	cat -n $(DEMO).c
	cc -I. $(DEMO).c runt.c bigmem.c
	./a.out

ci:
	ci-l *.c *.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go

clean:
	rm *.s *.o a.out
