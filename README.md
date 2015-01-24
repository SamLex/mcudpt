mcudpt
======

Multi-Client UDP Tunnel

Mcudpt is a small program written in Go (Go-lang) designed to mimic the functionality of ssh -R for UDP packets, in tunneling UDP packets over a TCP connection.

Note: this mimicing does not include encryption: the UDP packets are tunnled over an unencrypted TCP connection. If you wish to have encryption, use this on top of a ssh TCP tunnel.

Install
-------

This requires the Go (go-lang) dev tools and a properly setup go workspace (http://golang.org/doc/code.html)

Just execute 'go get github.com/SamLex/mcudpt'. This will create a mcudpt binary in your workspaces bin directory

Usage
-----

The program is split into two components: the receiver and the sender. The receiver listens for incoming UDP packets and sends them to the connected sender over TCP. The sender connects to the receiver and sends any UDP packets it is sent by the receiver to the address given.

Below is an example for forwarding DNS requests.

Receiver:

Set receiver to listen on port 53 for UDP connections and port 5353 for TCP connections

./mcudpt -mode receiver -udp :53 -tcp :5353

Receiver now listens for TCP connection from sender

Sender:

Set sender to connect to tunnel-server (setup above) on TCP port 5353 and send UDP packets to internal-dns.lan:53

./mcudpt -mode sender -udp internal-dns.lan:53 -tcp tunnel-server:5353

Now any DNS requests made to tunnel-server on port 53, will be tunneled to internal-dns port 53 and the reply sent back to the correct address

Security Note
-------------

Before the receiever accepts any messages from the sender it asks for user verification that it is the right sender. This is due to the fact that a malicious sender could use a receiver as a DDOS masker/amplifier. Once one sender has been accepted, however, no more can connect so only one verification is required.

Resource Note
-------------

Because of the way this program and the UDP protocol are designed, the UDP sockets opened by the sender are never closed during execution so during long-running exececutions or high-traffic situations, an occasional restart is advised to clear the open UDP sockets. This problem may be fixed in the future.
