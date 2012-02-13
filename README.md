# Monit to Graphite

The code is intended to forward events emitted by monit via it's m/monit interface to Graphite

## Compiling

You need go r60.3

In the directory type make

## Setup

in your monitrc:

set mmonit http://127.0.0.1:3005/collector

adapt the address to where you want monit to send the xml report it is generating

run the forwarder:

./forwarder
