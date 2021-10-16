# Reliable Data Transfer For UDP with Selective Repeat

- TODO
  - [x] Reliable Data Transfer
  - [ ] Packet Loss Simulation
  - [ ] Tests
  - [x] Documentation

# Explaination

This program has three major componetes
- packet: represents a packet with header to handle reliable data transfer
- rdt: the reliable data transfer component which uses a udt component and implements Selective Repeat
- udt: which handles unreliable data transfer which is implemented with udp as the transport layer

The main file runs loop to get user input and also run receiver and sender on different goroutines (basic units that are scheduled on green threads and they run concurrently).

# Run

No that you need to install go compiler so do that first

to run the project do as following
```bash
$ go run main.go <listen_addr> <peer_addr> [debug]
# examples
# peer 1
$ go run main.go :5000 :5001 debug
# peer 2
$ go run main.go :5001 :5000 debug

# another example
# peer 1
$ go run main.go :5000 192.168.1.10:5001 debug
# peer 2
$ go run main.go :5001 192.168.1.11:5000 debug
```

the debug flag stores extra logs to `host_<addr>.log` for example `host_:5000.log` or `host_192.168.1.10:5000.log`. These logs contains information about lost packets and when they are being retransmited. without this option you can't see this information.

# Documentation

To see the documentation you can browse through the code or you can run this command to see the doc
```go
$ go doc <pkg>
# example rdt is a package in rdt.go
$ go doc rdt
```
this will print the docs of this package to stdout so you can store it to a file
this way you don't need to see the code in details you just see the parts you are intended to see

