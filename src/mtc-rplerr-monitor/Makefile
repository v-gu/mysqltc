include $(GOROOT)/src/Make.inc

TARG=mtc-rplerr-monitor
GOFILES=\
	mtc-rplerr-monitor.go\

GOPATH=../..
GCIMPORTS=$(patsubst %,-I %/pkg/$(GOOS)_$(GOARCH),$(subst :, ,$(GOPATH)))
LDIMPORTS=$(patsubst %,-L %/pkg/$(GOOS)_$(GOARCH),$(subst :, ,$(GOPATH)))

include $(GOROOT)/src/Make.cmd
