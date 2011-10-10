# Makefile for mtc-rpl-monitor.

include $(GOROOT)/src/Make.inc

TARG=bin/mtc-rpl-monitor
GOFILES=\
		mtc-rpl-monitor.go\

include $(GOROOT)/src/Make.cmd