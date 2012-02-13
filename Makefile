include ${GOROOT}/src/Make.inc

TARG=forwarder

GOFILES=\
        charset_reader.go\
        forwarder.go\

include ${GOROOT}/src/Make.cmd
