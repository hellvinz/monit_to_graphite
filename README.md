# Monit to Graphite

The code is intended to forward events emitted by monit via it's m/monit interface to Graphite

## Compiling

You need go r1

go get github.com/hellvinz/monit_to_graphite

## Setup

in your monitrc:

set mmonit http://127.0.0.1:3005/collector

adapt the address to where you want monit to send the xml report it is generating

run the forwarder:

~/GOPATH/bin/monit_to_graphite
