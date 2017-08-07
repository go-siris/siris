SHELL         := /bin/sh
GO            := $(firstword $(subst :, ,$(GOPATH)))
GOCOV         := $(GOPATH)/bin/gocov
GOCOVMERGE    := $(GOPATH)/bin/gocovmerge
COVERFILEBIN  := $(addsuffix /coverage.out,$(addprefix coverage/,github.com/go-siris/siris))
TESTBIN       := siristest.exe

# helper
comma:= ,
empty:=
space:= $(empty) $(empty)

# List of pkgs for the project
PKGS          = $(shell go list ./... | grep -v vendor | grep -v "github.com/go-siris/siris/httptest" | grep -v "github.com/go-siris/siris/sessions/" | grep -v "github.com/go-siris/siris/typescript"  | grep -v "github.com/go-siris/siris/websocket")
PKGSLIST      = $(subst $(space),$(comma),$(PKGS))

# Coverage output: coverage/$PKG/coverage.out
COVPKGS=$(addsuffix /coverage.out,$(addprefix coverage/,$(PKGS)))

.FORCE:
all: testbuild coverage/all.out mergecoverfiles clean

coverage/all.out: testrunbin $(COVPKGS)
	echo "mode: set" >$@
	grep -hv "mode: set" $(wildcard $^) >>$@

$(COVPKGS): .FORCE
	@ mkdir -p $(dir $@)
	@ go test -coverprofile $@ $(patsubst coverage/%/coverage.out,%,$@)


.PHONY: testbuild
testbuild:
	go test -c -coverpkg="$(PKGSLIST)" -tags testsiris -o $(TESTBIN)

.PHONY: mergecoverfiles
mergecoverfiles:
	echo "" > coverage.txt
	cat coverage/all.out >> coverage.txt
	sed -i '/^$$/d' coverage.txt

.PHONY: testrunbin
testrunbin:
	@ mkdir -p $(dir $(COVERFILEBIN))
	@ ./$(TESTBIN) -test.v -test.short -test.run "^TestSiris$$" -test.coverprofile=$(COVERFILEBIN)

.PHONY: clean
clean:
	rm -rf coverage/*
	rm -f $(TESTBIN)
