include $(GOROOT)/src/Make.inc

TARG=mtclib
GOFILES=\
	mtclib.go\

GOPATH=../..
GCIMPORTS=$(patsubst %,-I %/pkg/$(GOOS)_$(GOARCH),$(subst :, ,$(GOPATH)))
LDIMPORTS=$(patsubst %,-L %/pkg/$(GOOS)_$(GOARCH),$(subst :, ,$(GOPATH)))

include $(GOROOT)/src/Make.pkg
