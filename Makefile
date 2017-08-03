SHELL       := /bin/sh
COVERFILE   := system.out
TESTBIN     := siristest.bin

all: testbuild testrun clean

.PHONY: testbuild
testbuild:
	go test -c -covermode=atomic -coverpkg="github.com/go-siris/siris,github.com/go-siris/siris/context,github.com/go-siris/siris/cache,github.com/go-siris/siris/core/router" -tags testsiris -o $(TESTBIN)

.PHONY: testrun
testrun:
	./$(TESTBIN) -test.v -test.run "^TestSiris$$" -test.coverprofile=$(COVERFILE)

.PHONY: clean
clean:
	rm -f $(TESTBIN)
