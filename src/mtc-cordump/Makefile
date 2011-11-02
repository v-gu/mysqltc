include $(GOROOT)/src/Make.inc

TARG=mtc-cordump
GOFILES=\
	mtc-cordump.go\

GOPATH=../..
LDIMPORTS=$(patsubst %,-L %/pkg/$(GOOS)_$(GOARCH),$(subst :, ,$(GOPATH)))
GCIMPORTS=$(patsubst %,-I %/pkg/$(GOOS)_$(GOARCH),$(subst :, ,$(GOPATH)))

include $(GOROOT)/src/Make.cmd
