mcudpt
======

Multi-Client UDP Tunnel

Mcudpt is a small program written in Go (Go-lang) designed to mimic the functionality of ssh -R for UDP packets.

Note: this mimicing does not include encryption: the UDP packets are tunnled over an unencrypted TCP connection. If you wish to have encryption, use this on top of a ssh TCP tunnel.

Install
-------

This requires the Go (go-lang) dev tools (go build command)

Clone this repostitory
Execute 'go build mcudpt'
Executable is created as 'mcudpt'

Usage
-----

Server:

Set server to listen on port 53 for UDP connections and port 5353 for TCP connections

./mcudpt -mode server -udp :53 -tcp :5353

Server now listens for TCP connection from client

Client:

Set client to connect to tunnel server (setup above) on TCP port 5353 and send tunneled UDP packets to internal-dns.lan:53

./mcudpt -mode client -udp internal-dns.lan:53 -tcp tunnel-server:5353

Now any DNS requests (or any other UDP packets) made to tunnel-server on port 53, will be tunneled to internal-dns port 53 and the reply sent back to the correct address
