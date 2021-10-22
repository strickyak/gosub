DEMO=demo/wc.go
all:
	go run gosub.go  < $(DEMO)   > $(DEMO).c 
	clang-format -i --style=Google $(DEMO).c
	cat -n $(DEMO).c
	# cc -I. $(DEMO).c runt.c bigmem.c
	# ./a.out

ci:
	ci-l *.c *.h Makefile *.go */*.go

fmt:
	gofmt -w *.go */*.go

clean:
	rm *.s *.o a.out
